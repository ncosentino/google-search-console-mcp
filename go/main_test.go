package main

import (
	"context"
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
		&mcp.Tool{Name: "list_sites", Description: "test"},
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
