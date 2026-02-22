// Command google-search-console-mcp is an MCP server that exposes Google Search Console
// search analytics as tools for AI assistants. It communicates via STDIO using the MCP protocol.
//
// Usage:
//
//	google-search-console-mcp [--service-account-file <path>]
//
// Credential resolution order: --service-account-file flag,
// GOOGLE_SERVICE_ACCOUNT_FILE env var, GOOGLE_SERVICE_ACCOUNT_JSON env var, .env file.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-search-console-mcp/go/internal/config"
	"github.com/ncosentino/google-search-console-mcp/go/internal/searchconsole"
)

var version = "dev"

func main() {
	serviceAccountFile := flag.String("service-account-file", "", "Path to Google service account JSON key file")
	flag.Parse()

	// All diagnostic output must go to stderr to avoid corrupting the MCP STDIO stream.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Resolve(*serviceAccountFile)
	if len(cfg.ServiceAccountJSON) == 0 {
		slog.Error("no service account credentials provided",
			"hint", "set --service-account-file flag, GOOGLE_SERVICE_ACCOUNT_FILE env var, or GOOGLE_SERVICE_ACCOUNT_JSON env var")
		os.Exit(1)
	}

	client, err := searchconsole.NewClient(cfg.ServiceAccountJSON)
	if err != nil {
		slog.Error("failed to create Search Console client", "err", err)
		os.Exit(1)
	}

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-search-console-mcp",
		Version: version,
	}, nil)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "query_search_analytics",
			Description: "Query Google Search Console search analytics. Returns clicks, impressions, CTR, and average position grouped by the specified dimensions (query, page, country, device, date).",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input querySearchAnalyticsInput) (*mcp.CallToolResult, any, error) {
			return querySearchAnalytics(ctx, client, input)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "list_sites",
			Description: "List all Google Search Console properties (sites) the service account has access to.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ listSitesInput) (*mcp.CallToolResult, any, error) {
			return listSites(ctx, client)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "list_sitemaps",
			Description: "List sitemaps submitted to Google Search Console for a specific property.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input listSitemapsInput) (*mcp.CallToolResult, any, error) {
			return listSitemaps(ctx, client, input.SiteURL)
		},
	)

	slog.Info("google-search-console-mcp starting", "version", version, "transport", "stdio")
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

// querySearchAnalyticsInput is the input schema for the query_search_analytics tool.
type querySearchAnalyticsInput struct {
	SiteURL    string   `json:"site_url"`
	StartDate  string   `json:"start_date"`
	EndDate    string   `json:"end_date"`
	Dimensions []string `json:"dimensions"`
	RowLimit   int      `json:"row_limit"`
}

// listSitesInput is the input schema for the list_sites tool (no parameters required).
type listSitesInput struct{}

// listSitemapsInput is the input schema for the list_sitemaps tool.
type listSitemapsInput struct {
	SiteURL string `json:"site_url"`
}

func querySearchAnalytics(ctx context.Context, client *searchconsole.Client, input querySearchAnalyticsInput) (*mcp.CallToolResult, any, error) {
	result, err := client.QuerySearchAnalytics(ctx, input.SiteURL, input.StartDate, input.EndDate, input.Dimensions, input.RowLimit)
	if err != nil {
		errResult := map[string]string{"error": fmt.Sprintf("querying search analytics: %v", err)}
		b, _ := json.Marshal(errResult)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
	}
	b, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling result: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func listSites(ctx context.Context, client *searchconsole.Client) (*mcp.CallToolResult, any, error) {
	result, err := client.ListSites(ctx)
	if err != nil {
		errResult := map[string]string{"error": fmt.Sprintf("listing sites: %v", err)}
		b, _ := json.Marshal(errResult)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
	}
	b, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling result: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func listSitemaps(ctx context.Context, client *searchconsole.Client, siteURL string) (*mcp.CallToolResult, any, error) {
	result, err := client.ListSitemaps(ctx, siteURL)
	if err != nil {
		errResult := map[string]string{"error": fmt.Sprintf("listing sitemaps: %v", err)}
		b, _ := json.Marshal(errResult)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
	}
	b, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling result: %w", err)
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}
