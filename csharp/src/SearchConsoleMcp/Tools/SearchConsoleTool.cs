using System.ComponentModel;
using System.Text.Json;
using ModelContextProtocol.Server;
using SearchConsoleMcp.SearchConsole;

namespace SearchConsoleMcp.Tools;

/// <summary>MCP tools for Google Search Console.</summary>
[McpServerToolType]
internal sealed class SearchConsoleTool(SearchConsoleClient client)
{
    [McpServerTool(Name = "query_search_analytics")]
    [Description("Query Google Search Console search analytics. Returns clicks, impressions, CTR, and average position. Dimensions can be any combination of: query, page, country, device, date.")]
    internal async Task<string> QuerySearchAnalytics(
        [Description("The Search Console property URL (e.g. 'https://www.example.com/' or 'sc-domain:example.com').")] string site_url,
        [Description("Start date in YYYY-MM-DD format.")] string start_date,
        [Description("End date in YYYY-MM-DD format.")] string end_date,
        [Description("Dimensions to group by. Valid values: query, page, country, device, date. Pass an empty array to get aggregate totals.")] string[] dimensions,
        [Description("Maximum number of rows to return (1-25000). Defaults to 1000.")] int row_limit = 1000,
        CancellationToken cancellationToken = default)
    {
        try
        {
            var result = await client
                .QuerySearchAnalyticsAsync(site_url, start_date, end_date, dimensions, row_limit, cancellationToken)
                .ConfigureAwait(false);
            return JsonSerializer.Serialize(result, GscJsonContext.Default.SearchAnalyticsResponse);
        }
        catch (Exception ex)
        {
            return JsonSerializer.Serialize(
                new ErrorResult($"{ex.GetType().Name}: {ex.Message}"),
                GscJsonContext.Default.ErrorResult);
        }
    }

    [McpServerTool(Name = "list_sites")]
    [Description("List all Google Search Console properties (sites) the service account has access to, along with permission levels.")]
    internal async Task<string> ListSites(CancellationToken cancellationToken = default)
    {
        try
        {
            var result = await client.ListSitesAsync(cancellationToken).ConfigureAwait(false);
            return JsonSerializer.Serialize(result, GscJsonContext.Default.SiteListResponse);
        }
        catch (Exception ex)
        {
            return JsonSerializer.Serialize(
                new ErrorResult($"{ex.GetType().Name}: {ex.Message}"),
                GscJsonContext.Default.ErrorResult);
        }
    }

    [McpServerTool(Name = "list_sitemaps")]
    [Description("List sitemaps submitted to Google Search Console for a specific property, including submission status and error counts.")]
    internal async Task<string> ListSitemaps(
        [Description("The Search Console property URL (e.g. 'https://www.example.com/').")] string site_url,
        CancellationToken cancellationToken = default)
    {
        try
        {
            var result = await client.ListSitemapsAsync(site_url, cancellationToken).ConfigureAwait(false);
            return JsonSerializer.Serialize(result, GscJsonContext.Default.SitemapListResponse);
        }
        catch (Exception ex)
        {
            return JsonSerializer.Serialize(
                new ErrorResult($"{ex.GetType().Name}: {ex.Message}"),
                GscJsonContext.Default.ErrorResult);
        }
    }
}
