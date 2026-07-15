using System.Text.Json;
using SearchConsoleMcp.SearchConsole;
using SearchConsoleMcp.Tools;
using Xunit;

namespace SearchConsoleMcp.Tests;

/// <summary>
/// Characterization tests for SearchConsoleTool's three MCP tool methods, written ahead
/// of the ModelContextProtocol SDK dependency modernization (issue #6). Before these
/// tests, all three methods were at 0% coverage -- nothing exercised the JSON
/// serialization or the try/catch-to-ErrorResult conversion that wraps every call.
/// </summary>
public sealed class SearchConsoleToolTests
{
    [Fact]
    public async Task QuerySearchAnalytics_Success_ReturnsSerializedResult()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.OkJson(new
        {
            rows = new[] { new { keys = new[] { "hello world" }, clicks = 5.0, impressions = 100.0, ctr = 0.05, position = 3.2 } }
        }));
        var client = new SearchConsoleClient(new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");
        var tool = new SearchConsoleTool(client);

        var result = await tool.QuerySearchAnalytics("devleader.ca", "2025-01-01", "2025-12-31", dimensions: [], row_limit: 10);

        Assert.Contains("hello world", result, StringComparison.Ordinal);
    }

    [Fact]
    public async Task QuerySearchAnalytics_ApiError_ReturnsErrorResult()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.ServerError());
        var client = new SearchConsoleClient(new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");
        var tool = new SearchConsoleTool(client);

        var result = await tool.QuerySearchAnalytics("devleader.ca", "2025-01-01", "2025-12-31", dimensions: []);

        Assert.Contains("GscApiException", result, StringComparison.Ordinal);
    }

    [Fact]
    public async Task QuerySearchAnalytics_InvalidSearchType_ReturnsErrorResult()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.OkJson(new { rows = Array.Empty<object>() }));
        var client = new SearchConsoleClient(new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");
        var tool = new SearchConsoleTool(client);

        var result = await tool.QuerySearchAnalytics(
            "devleader.ca", "2025-01-01", "2025-12-31", dimensions: [], search_type: "youtube");

        Assert.Contains("invalid search_type", result, StringComparison.Ordinal);
        Assert.Equal(0, handler.CallCount);
    }

    [Fact]
    public async Task ListSites_Success_ReturnsSerializedResult()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.OkJson(new
        {
            siteEntry = new[] { new { siteUrl = "sc-domain:devleader.ca", permissionLevel = "siteFullUser" } }
        }));
        var client = new SearchConsoleClient(new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");
        var tool = new SearchConsoleTool(client);

        var result = await tool.ListSites();

        Assert.Contains("sc-domain:devleader.ca", result, StringComparison.Ordinal);
    }

    [Fact]
    public async Task ListSites_ApiError_ReturnsErrorResult()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.ServerError());
        var client = new SearchConsoleClient(new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");
        var tool = new SearchConsoleTool(client);

        var result = await tool.ListSites();

        Assert.Contains("GscApiException", result, StringComparison.Ordinal);
    }

    [Fact]
    public async Task ListSitemaps_Success_ReturnsSerializedResult()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.OkJson(new
        {
            sitemap = new object[]
            {
                new { path = "https://devleader.ca/string.xml", isPending = false, isSitemapsIndex = false, type = "sitemap", warnings = "2", errors = "0" },
                new { path = "https://devleader.ca/null.xml", isPending = false, isSitemapsIndex = false, type = "sitemap", warnings = (object?)null, errors = (object?)null },
                new { path = "https://devleader.ca/missing.xml", isPending = false, isSitemapsIndex = false, type = "sitemap" },
                new { path = "https://devleader.ca/numeric.xml", isPending = false, isSitemapsIndex = false, type = "sitemap", warnings = 0, errors = 1 }
            }
        }));
        var client = new SearchConsoleClient(new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");
        var tool = new SearchConsoleTool(client);

        var result = await tool.ListSitemaps("devleader.ca");

        using var document = JsonDocument.Parse(result);
        var sitemaps = document.RootElement.GetProperty("sitemaps");
        Assert.Equal(4, sitemaps.GetArrayLength());
        Assert.Equal(2L, sitemaps[0].GetProperty("warnings").GetInt64());
        Assert.Equal(0L, sitemaps[0].GetProperty("errors").GetInt64());
        Assert.Equal(JsonValueKind.Null, sitemaps[1].GetProperty("warnings").ValueKind);
        Assert.Equal(JsonValueKind.Null, sitemaps[1].GetProperty("errors").ValueKind);
        Assert.Equal(JsonValueKind.Null, sitemaps[2].GetProperty("warnings").ValueKind);
        Assert.Equal(JsonValueKind.Null, sitemaps[2].GetProperty("errors").ValueKind);
        Assert.Equal(0L, sitemaps[3].GetProperty("warnings").GetInt64());
        Assert.Equal(1L, sitemaps[3].GetProperty("errors").GetInt64());
    }

    [Fact]
    public async Task ListSitemaps_InvalidCounter_ReturnsErrorResult()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.OkJson(new
        {
            sitemap = new[]
            {
                new { path = "https://devleader.ca/sitemap.xml", warnings = "invalid", errors = "0" }
            }
        }));
        var client = new SearchConsoleClient(new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");
        var tool = new SearchConsoleTool(client);

        var result = await tool.ListSitemaps("devleader.ca");

        using var document = JsonDocument.Parse(result);
        var error = document.RootElement.GetProperty("error").GetString();
        Assert.NotNull(error);
        Assert.Contains("JsonException", error, StringComparison.Ordinal);
        Assert.Contains("warnings", error, StringComparison.Ordinal);
    }

    [Fact]
    public async Task ListSitemaps_ApiError_ReturnsErrorResult()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.ServerError());
        var client = new SearchConsoleClient(new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");
        var tool = new SearchConsoleTool(client);

        var result = await tool.ListSitemaps("devleader.ca");

        Assert.Contains("GscApiException", result, StringComparison.Ordinal);
    }
}
