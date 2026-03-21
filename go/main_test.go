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
			InputSchema: listSitesInputSchema,
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
