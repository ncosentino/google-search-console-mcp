using System.Net;
using System.Text.Json;
using SearchConsoleMcp.SearchConsole;
using Xunit;

namespace SearchConsoleMcp.Tests;

/// <summary>Tests for SearchConsoleClient site URL normalization and 403 retry logic.</summary>
public sealed class SearchConsoleClientSiteUrlTests
{
    [Fact]
    public async Task QuerySearchAnalytics_NormalizesBareInput_ToSCDomain()
    {
        string? requestedUrl = null;
        var handler = new FakeMessageHandler(req =>
        {
            requestedUrl = req.RequestUri?.AbsoluteUri;
            return FakeResponses.OkJson(new { rows = Array.Empty<object>() });
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler),
            new FakeTokenProvider(),
            baseUrlOverride: "http://localhost/gsc");

        await client.QuerySearchAnalyticsAsync(
            "devleader.ca", "2025-01-01", "2025-12-31", null, 10);

        Assert.NotNull(requestedUrl);
        Assert.Contains(Uri.EscapeDataString("sc-domain:devleader.ca"), requestedUrl);
    }

    [Fact]
    public async Task QuerySearchAnalytics_On403_RetriesWithResolvedUrl()
    {
        var handler = new FakeMessageHandler(req =>
        {
            var encodedPrefix = Uri.EscapeDataString("https://www.devleader.ca/");
            var encodedScDomain = Uri.EscapeDataString("sc-domain:devleader.ca");

            if (req.Method == HttpMethod.Post && req.RequestUri!.AbsoluteUri.Contains(encodedPrefix))
                return FakeResponses.Forbidden();

            if (req.Method == HttpMethod.Get && req.RequestUri!.AbsolutePath.TrimEnd('/').EndsWith("/sites"))
                return FakeResponses.OkJson(new
                {
                    siteEntry = new[]
                    {
                        new { siteUrl = "sc-domain:devleader.ca", permissionLevel = "siteFullUser" }
                    }
                });

            if (req.Method == HttpMethod.Post && req.RequestUri!.AbsoluteUri.Contains(encodedScDomain))
                return FakeResponses.OkJson(new { rows = Array.Empty<object>() });

            return new HttpResponseMessage(HttpStatusCode.NotFound);
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler),
            new FakeTokenProvider(),
            baseUrlOverride: "http://localhost/gsc");

        var result = await client.QuerySearchAnalyticsAsync(
            "https://www.devleader.ca/", "2025-01-01", "2025-12-31", null, 10);

        Assert.NotNull(result);
        Assert.Equal(3, handler.CallCount); // 403 + ListSites + retry
    }

    [Fact]
    public async Task ListSitemaps_On403_RetriesWithResolvedUrl()
    {
        var handler = new FakeMessageHandler(req =>
        {
            var encodedPrefix = Uri.EscapeDataString("https://www.devleader.ca/");
            var encodedScDomain = Uri.EscapeDataString("sc-domain:devleader.ca");

            if (req.Method == HttpMethod.Get && req.RequestUri!.AbsoluteUri.Contains(encodedPrefix + "/sitemaps"))
                return FakeResponses.Forbidden();

            if (req.Method == HttpMethod.Get && req.RequestUri!.AbsolutePath.TrimEnd('/').EndsWith("/sites"))
                return FakeResponses.OkJson(new
                {
                    siteEntry = new[]
                    {
                        new { siteUrl = "sc-domain:devleader.ca", permissionLevel = "siteFullUser" }
                    }
                });

            if (req.Method == HttpMethod.Get && req.RequestUri!.AbsoluteUri.Contains(encodedScDomain + "/sitemaps"))
                return FakeResponses.OkJson(new { sitemap = Array.Empty<object>() });

            return new HttpResponseMessage(HttpStatusCode.NotFound);
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler),
            new FakeTokenProvider(),
            baseUrlOverride: "http://localhost/gsc");

        var result = await client.ListSitemapsAsync("https://www.devleader.ca/");

        Assert.NotNull(result);
        Assert.Equal(3, handler.CallCount);
    }

    [Fact]
    public async Task InspectUrl_On403_RetriesWithResolvedUrl()
    {
        var handler = new FakeMessageHandler(async (req, cancellationToken) =>
        {
            if (req.Method == HttpMethod.Get &&
                req.RequestUri!.AbsolutePath.TrimEnd('/').EndsWith("/sites", StringComparison.Ordinal))
            {
                return FakeResponses.OkJson(new
                {
                    siteEntry = new[]
                    {
                        new { siteUrl = "sc-domain:devleader.ca", permissionLevel = "siteFullUser" }
                    }
                });
            }

            if (req.Method == HttpMethod.Post &&
                req.RequestUri!.AbsolutePath.EndsWith(
                    "/urlInspection/index:inspect",
                    StringComparison.Ordinal))
            {
                var json = await req.Content!.ReadAsStringAsync(cancellationToken);
                using var document = JsonDocument.Parse(json);
                var siteUrl = document.RootElement.GetProperty("siteUrl").GetString();

                if (siteUrl == "https://www.devleader.ca/")
                    return FakeResponses.Forbidden();

                if (siteUrl == "sc-domain:devleader.ca")
                {
                    return FakeResponses.OkJson(new
                    {
                        inspectionResult = new
                        {
                            indexStatusResult = new { verdict = "PASS" }
                        }
                    });
                }
            }

            return new HttpResponseMessage(HttpStatusCode.NotFound);
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler),
            new FakeTokenProvider(),
            baseUrlOverride: "http://localhost/gsc");

        var result = await client.InspectUrlAsync(
            "https://www.devleader.ca/",
            "https://www.devleader.ca/example");

        Assert.NotNull(result);
        Assert.Equal("sc-domain:devleader.ca", result.SiteUrl);
        Assert.Equal(3, handler.CallCount);
    }
}
