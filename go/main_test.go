package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-search-console-mcp/go/internal/searchconsole"
)

// TestNewServer_RegistersTools verifies that newServer builds a server with all
// three tools registered and listable via a real client session, catching invalid
// struct tags or schema-generation failures at test time rather than at runtime.
func TestNewServer_RegistersTools(t *testing.T) {
	t.Parallel()

	client := searchconsole.NewTestClient(http.DefaultClient)
	srv := newServer(client)

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	defer serverSession.Close()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	var names []string
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	for _, want := range []string{"query_search_analytics", "list_sites", "list_sitemaps"} {
		found := false
		for _, n := range names {
			if n == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tool %q not registered; got tools %v", want, names)
		}
	}
}

// TestQuerySearchAnalytics_Success_ReturnsMarshaledResult confirms the
// query_search_analytics handler marshals a successful client response into
// CallToolResult content, with no error set.
func TestQuerySearchAnalytics_Success_ReturnsMarshaledResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[{"keys":["hello world"],"clicks":5,"impressions":100,"ctr":0.05,"position":3.2}]}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	result, _, err := querySearchAnalytics(context.Background(), client, querySearchAnalyticsInput{
		SiteURL:   "devleader.ca",
		StartDate: "2025-01-01",
		EndDate:   "2025-12-31",
		RowLimit:  10,
	})
	if err != nil {
		t.Fatalf("querySearchAnalytics: %v", err)
	}
	if result.IsError {
		t.Errorf("result.IsError = true, want false: %+v", result.Content)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "hello world") {
		t.Errorf("result text = %q, want it to contain %q", text, "hello world")
	}
}

// TestQuerySearchAnalytics_APIError_ReturnsErrorContent confirms an API-level
// error (e.g. a non-403 failure, or a 403 with no matching resolvable
// property) is surfaced as CallToolResult content rather than a Go error,
// matching this repo's established error-handling convention.
func TestQuerySearchAnalytics_APIError_ReturnsErrorContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	result, _, err := querySearchAnalytics(context.Background(), client, querySearchAnalyticsInput{
		SiteURL:   "devleader.ca",
		StartDate: "2025-01-01",
		EndDate:   "2025-12-31",
	})
	if err != nil {
		t.Fatalf("querySearchAnalytics returned a Go error instead of error content: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "querying search analytics:") {
		t.Errorf("result text = %q, want it to mention %q", text, "querying search analytics:")
	}
}

// TestListSites_Success_ReturnsMarshaledResult confirms the list_sites
// handler marshals a successful client response into CallToolResult content.
func TestListSites_Success_ReturnsMarshaledResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"siteEntry":[{"siteUrl":"sc-domain:devleader.ca","permissionLevel":"siteFullUser"}]}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	result, _, err := listSites(context.Background(), client)
	if err != nil {
		t.Fatalf("listSites: %v", err)
	}
	if result.IsError {
		t.Errorf("result.IsError = true, want false: %+v", result.Content)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "sc-domain:devleader.ca") {
		t.Errorf("result text = %q, want it to contain %q", text, "sc-domain:devleader.ca")
	}
}

// TestListSites_APIError_ReturnsErrorContent confirms an API-level error is
// surfaced as CallToolResult content rather than a Go error.
func TestListSites_APIError_ReturnsErrorContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	result, _, err := listSites(context.Background(), client)
	if err != nil {
		t.Fatalf("listSites returned a Go error instead of error content: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "listing sites:") {
		t.Errorf("result text = %q, want it to mention %q", text, "listing sites:")
	}
}

// TestListSitemaps_Success_ReturnsMarshaledResult confirms the list_sitemaps
// handler marshals a successful client response into CallToolResult content.
func TestListSitemaps_Success_ReturnsMarshaledResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sitemap":[{"path":"https://devleader.ca/sitemap.xml","isPending":false,"isSitemapsIndex":false,"type":"sitemap","warnings":0,"errors":0}]}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	result, _, err := listSitemaps(context.Background(), client, "devleader.ca")
	if err != nil {
		t.Fatalf("listSitemaps: %v", err)
	}
	if result.IsError {
		t.Errorf("result.IsError = true, want false: %+v", result.Content)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "sitemap.xml") {
		t.Errorf("result text = %q, want it to contain %q", text, "sitemap.xml")
	}
}

// TestListSitemaps_APIError_ReturnsErrorContent confirms an API-level error
// is surfaced as CallToolResult content rather than a Go error.
func TestListSitemaps_APIError_ReturnsErrorContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	result, _, err := listSitemaps(context.Background(), client, "devleader.ca")
	if err != nil {
		t.Fatalf("listSitemaps returned a Go error instead of error content: %v", err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "listing sitemaps:") {
		t.Errorf("result text = %q, want it to mention %q", text, "listing sitemaps:")
	}
}

// TestNewServer_CallQuerySearchAnalyticsTool_ViaRealSession confirms the
// query_search_analytics tool, as actually registered by newServer (not just
// the underlying Go function called directly), works end-to-end through a
// real MCP client session: argument binding, schema validation, and tool
// dispatch all have to agree for this to pass.
func TestNewServer_CallQuerySearchAnalyticsTool_ViaRealSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows":[]}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	mcpServer := newServer(client)

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	defer serverSession.Close()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "query_search_analytics",
		Arguments: map[string]any{
			"site_url":   "devleader.ca",
			"start_date": "2025-01-01",
			"end_date":   "2025-12-31",
			"dimensions": []string{"query"},
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Errorf("CallTool returned an error result: %+v", result.Content)
	}
}

// TestNewServer_CallListSitesTool_ViaRealSession is the list_sites
// equivalent of TestNewServer_CallQuerySearchAnalyticsTool_ViaRealSession.
func TestNewServer_CallListSitesTool_ViaRealSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"siteEntry":[]}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	mcpServer := newServer(client)

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	defer serverSession.Close()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_sites",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Errorf("CallTool returned an error result: %+v", result.Content)
	}
}

// TestNewServer_CallListSitemapsTool_ViaRealSession is the list_sitemaps
// equivalent of TestNewServer_CallQuerySearchAnalyticsTool_ViaRealSession.
func TestNewServer_CallListSitemapsTool_ViaRealSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sitemap":[]}`))
	}))
	defer srv.Close()
	defer searchconsole.SetTestAPIBaseURL(srv.URL)()

	client := searchconsole.NewTestClient(srv.Client())
	mcpServer := newServer(client)

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	defer serverSession.Close()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name:      "list_sitemaps",
		Arguments: map[string]any{"site_url": "devleader.ca"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Errorf("CallTool returned an error result: %+v", result.Content)
	}
}

// TestQuerySearchAnalyticsInput_RowLimit_IsNotRequired confirms row_limit is
// absent from the schema's required list, matching its description ("Defaults
// to 1000 if omitted") and the defaulting behavior in
// Client.querySearchAnalyticsWithURL. Regression test for a real bug found
// while writing this repo's characterization tests ahead of the go-sdk
// modernization (issue #5): row_limit had no ",omitempty" on its json tag, so
// the generated schema wrongly marked it required, and a real MCP client
// session omitting it (relying on the documented default) failed schema
// validation entirely.
func TestQuerySearchAnalyticsInput_RowLimit_IsNotRequired(t *testing.T) {
	t.Parallel()

	schema, err := jsonschema.For[querySearchAnalyticsInput](nil)
	if err != nil {
		t.Fatalf("schema inference failed: %v", err)
	}

	if slices.Contains(schema.Required, "row_limit") {
		t.Errorf("row_limit must not be in schema.Required (got %v)", schema.Required)
	}
}

// TestListSites_InputSchema validates the production listSitesInputSchema variable to
// ensure it is compatible with strict MCP clients (e.g. Copilot CLI) that require
// explicit properties, required, and additionalProperties fields.
func TestListSites_InputSchema(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(listSitesInputSchema, &schema); err != nil {
		t.Fatalf("unmarshal listSitesInputSchema: %v", err)
	}

	properties, ok := schema["properties"]
	if !ok {
		t.Error("list_sites InputSchema is missing the 'properties' field; strict MCP clients will reject it")
	} else {
		if _, ok := properties.(map[string]any); !ok {
			t.Error("list_sites InputSchema 'properties' field must be a JSON object")
		}
	}

	required, ok := schema["required"]
	if !ok {
		t.Error("list_sites InputSchema is missing the 'required' field")
	} else {
		if _, ok := required.([]any); !ok {
			t.Error("list_sites InputSchema 'required' field must be a JSON array (often empty)")
		}
	}

	additionalProperties, ok := schema["additionalProperties"]
	if !ok {
		t.Error("list_sites InputSchema is missing the 'additionalProperties' field")
	} else {
		additionalPropertiesBool, ok := additionalProperties.(bool)
		if !ok {
			t.Error("list_sites InputSchema 'additionalProperties' field must be a boolean")
		} else if additionalPropertiesBool {
			t.Error("list_sites InputSchema 'additionalProperties' must be false for strict MCP clients")
		}
	}
}
