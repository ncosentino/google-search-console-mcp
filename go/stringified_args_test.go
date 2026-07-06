package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-search-console-mcp/go/internal/searchconsole"
)

// newTestServerAndClient wires a real in-memory MCP client/server pair around
// the actual production query_search_analytics registration (including any
// middleware added by registerMiddleware), backed by a fake Search Console
// API that always returns an empty rows result. This exercises the full
// request pipeline -- schema validation, middleware, and the typed handler --
// exactly as a real MCP client would, rather than calling internal validation
// functions directly.
func newTestServerAndClient(t *testing.T, registerMiddleware func(*mcp.Server)) *mcp.ClientSession {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rows": []}`))
	}))
	t.Cleanup(srv.Close)
	t.Cleanup(searchconsole.SetTestAPIBaseURL(srv.URL))

	client := searchconsole.NewTestClient(srv.Client())

	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "test"}, nil)
	if registerMiddleware != nil {
		registerMiddleware(server)
	}
	mcp.AddTool(server,
		&mcp.Tool{Name: "query_search_analytics", Description: "test"},
		func(ctx context.Context, _ *mcp.CallToolRequest, input querySearchAnalyticsInput) (*mcp.CallToolResult, any, error) {
			return querySearchAnalytics(ctx, client, input)
		},
	)

	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession
}

// TestQuerySearchAnalytics_StringifiedDimensions_Fails is the RED half of the
// TDD cycle for the "dimensions arrives as a JSON-encoded string" bug (#7,
// the same class of bug as google-keyword-planner-mcp#2/#4). It reproduces
// the exact failure end-to-end, through the real schema-validation pipeline,
// with no middleware installed.
//
// Schema validation failures are surfaced as a normal CallToolResult with
// IsError=true (via CallToolResult.SetError), not as a protocol-level error --
// same go-sdk v1.5.0+ behavior verified in google-keyword-planner-mcp's own
// SDK modernization (issue #10/PR #13).
func TestQuerySearchAnalytics_StringifiedDimensions_Fails(t *testing.T) {
	t.Parallel()

	clientSession := newTestServerAndClient(t, nil)

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "query_search_analytics",
		Arguments: map[string]any{
			"site_url":   "devleader.ca",
			"start_date": "2025-01-01",
			"end_date":   "2025-12-31",
			// Simulates a client that double-encodes the array into a JSON string
			// before sending it, instead of a genuine JSON array.
			"dimensions": `["query"]`,
		},
	})
	if err != nil {
		t.Fatalf("unexpected protocol-level error: %v", err)
	}

	if !result.IsError {
		t.Fatal("expected a schema validation error result for stringified dimensions, got a successful result")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "dimensions") {
		t.Errorf("result text = %q, want it to mention dimensions", text)
	}
}

// TestQuerySearchAnalytics_StringifiedDimensions_CoercedByMiddleware is the
// GREEN half: with coerceStringifiedArrayArgs installed, the same stringified
// input is repaired before schema validation runs, and the call succeeds.
func TestQuerySearchAnalytics_StringifiedDimensions_CoercedByMiddleware(t *testing.T) {
	t.Parallel()

	clientSession := newTestServerAndClient(t, func(s *mcp.Server) {
		s.AddReceivingMiddleware(coerceStringifiedArrayArgs(toolArrayFields))
	})

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "query_search_analytics",
		Arguments: map[string]any{
			"site_url":   "devleader.ca",
			"start_date": "2025-01-01",
			"end_date":   "2025-12-31",
			"dimensions": `["query"]`,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Errorf("result.IsError = true, want false: %+v", result.Content)
	}
}

// TestQuerySearchAnalytics_GenuineArrayDimensions_StillWorks confirms the
// coercion middleware is a no-op for well-formed clients that already send a
// genuine JSON array -- it must not interfere with the standard-compliant path.
func TestQuerySearchAnalytics_GenuineArrayDimensions_StillWorks(t *testing.T) {
	t.Parallel()

	clientSession := newTestServerAndClient(t, func(s *mcp.Server) {
		s.AddReceivingMiddleware(coerceStringifiedArrayArgs(toolArrayFields))
	})

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "query_search_analytics",
		Arguments: map[string]any{
			"site_url":   "devleader.ca",
			"start_date": "2025-01-01",
			"end_date":   "2025-12-31",
			"dimensions": []string{"query"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error for a genuine JSON array: %v", err)
	}
	if result.IsError {
		t.Errorf("result.IsError = true, want false: %+v", result.Content)
	}
}

// TestQuerySearchAnalytics_OmittedDimensions_StillWorks confirms the coercion
// middleware is a no-op when dimensions is omitted entirely (aggregate totals,
// per the tool's documented behavior) -- it must not require the field to be
// present just because it's in the array-fields map.
func TestQuerySearchAnalytics_OmittedDimensions_StillWorks(t *testing.T) {
	t.Parallel()

	clientSession := newTestServerAndClient(t, func(s *mcp.Server) {
		s.AddReceivingMiddleware(coerceStringifiedArrayArgs(toolArrayFields))
	})

	result, err := clientSession.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "query_search_analytics",
		Arguments: map[string]any{
			"site_url":   "devleader.ca",
			"start_date": "2025-01-01",
			"end_date":   "2025-12-31",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error when dimensions is omitted: %v", err)
	}
	if result.IsError {
		t.Errorf("result.IsError = true, want false: %+v", result.Content)
	}
}
