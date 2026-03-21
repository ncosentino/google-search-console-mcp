package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-search-console-mcp/go/internal/searchconsole"
)

// TestNewServer_RegistersTools verifies that the MCP server can be created and all
// tools can be registered without panicking. This catches invalid struct tags or
// schema-generation failures at test time rather than at runtime.
func TestNewServer_RegistersTools(_ *testing.T) {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-search-console-mcp",
		Version: "test",
	}, nil)

	// Use a nil client -- tools are registered only, not called.
	var client *searchconsole.Client

	mcp.AddTool(srv,
		&mcp.Tool{Name: "query_search_analytics", Description: "test"},
		func(ctx context.Context, _ *mcp.CallToolRequest, input querySearchAnalyticsInput) (*mcp.CallToolResult, any, error) {
			return querySearchAnalytics(ctx, client, input)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "list_sites",
			Description: "test",
			InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[],"additionalProperties":false}`),
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ listSitesInput) (*mcp.CallToolResult, any, error) {
			return listSites(ctx, client)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{Name: "list_sitemaps", Description: "test"},
		func(ctx context.Context, _ *mcp.CallToolRequest, input listSitemapsInput) (*mcp.CallToolResult, any, error) {
			return listSitemaps(ctx, client, input.SiteURL)
		},
	)
}

// TestListSites_InputSchema verifies that the list_sites tool exposes an explicit
// input schema compatible with strict MCP clients (e.g. Copilot CLI) that reject
// tools whose schema is missing a "properties" field.
func TestListSites_InputSchema(t *testing.T) {
	tool := &mcp.Tool{
		Name:        "list_sites",
		Description: "test",
		InputSchema: json.RawMessage(`{"type":"object","properties":{},"required":[],"additionalProperties":false}`),
	}

	rawSchema, ok := tool.InputSchema.(json.RawMessage)
	if !ok {
		t.Fatal("InputSchema is not a json.RawMessage")
	}

	var schema map[string]any
	if err := json.Unmarshal(rawSchema, &schema); err != nil {
		t.Fatalf("unmarshal InputSchema: %v", err)
	}

	if _, ok := schema["properties"]; !ok {
		t.Error("list_sites InputSchema is missing the 'properties' field; strict MCP clients will reject it")
	}

	if _, ok := schema["required"]; !ok {
		t.Error("list_sites InputSchema is missing the 'required' field")
	}

	// Verify the tool can be registered without panicking.
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-search-console-mcp",
		Version: "test",
	}, nil)
	var client *searchconsole.Client
	mcp.AddTool(srv,
		tool,
		func(ctx context.Context, _ *mcp.CallToolRequest, _ listSitesInput) (*mcp.CallToolResult, any, error) {
			return listSites(ctx, client)
		},
	)
}
