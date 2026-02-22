using System.Net;
using System.Text.Json;
using SearchConsoleMcp.SearchConsole;
using Xunit;

namespace SearchConsoleMcp.Tests;

/// <summary>Tests for SearchConsoleClient site URL normalization and 403 retry logic.</summary>
public sealed class SearchConsoleClientSiteUrlTests
{
    private sealed class FakeTokenProvider : ITokenProvider
    {
        public Task<string> GetAccessTokenAsync(CancellationToken cancellationToken = default)
            => Task.FromResult("fake-token");
    }

    private sealed class FakeMessageHandler : HttpMessageHandler
    {
        private readonly Func<HttpRequestMessage, HttpResponseMessage> _handler;
        public int CallCount { get; private set; }

        internal FakeMessageHandler(Func<HttpRequestMessage, HttpResponseMessage> handler)
            => _handler = handler;

        protected override Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken cancellationToken)
        {
            CallCount++;
            return Task.FromResult(_handler(request));
        }
    }

    private static HttpResponseMessage OkJson(object body)
    {
        var json = JsonSerializer.Serialize(body);
        return new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new StringContent(json, System.Text.Encoding.UTF8, "application/json")
        };
    }

    private static HttpResponseMessage Forbidden()
        => new HttpResponseMessage(HttpStatusCode.Forbidden)
        {
            Content = new StringContent(@"{""error"":{""code"":403}}", System.Text.Encoding.UTF8, "application/json")
        };

    [Fact]
    public async Task QuerySearchAnalytics_NormalizesBareInput_ToSCDomain()
    {
        string? requestedUrl = null;
        var handler = new FakeMessageHandler(req =>
        {
            requestedUrl = req.RequestUri?.AbsoluteUri;
            return OkJson(new { rows = Array.Empty<object>() });
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
        var callCount = 0;
        var handler = new FakeMessageHandler(req =>
        {
            callCount++;
            var encodedPrefix = Uri.EscapeDataString("https://www.devleader.ca/");
            var encodedScDomain = Uri.EscapeDataString("sc-domain:devleader.ca");

            if (req.Method == HttpMethod.Post && req.RequestUri!.AbsoluteUri.Contains(encodedPrefix))
                return Forbidden();

            if (req.Method == HttpMethod.Get && req.RequestUri!.AbsolutePath.TrimEnd('/').EndsWith("/sites"))
                return OkJson(new
                {
                    siteEntry = new[]
                    {
                        new { siteUrl = "sc-domain:devleader.ca", permissionLevel = "siteFullUser" }
                    }
                });

            if (req.Method == HttpMethod.Post && req.RequestUri!.AbsoluteUri.Contains(encodedScDomain))
                return OkJson(new { rows = Array.Empty<object>() });

            return new HttpResponseMessage(HttpStatusCode.NotFound);
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler),
            new FakeTokenProvider(),
            baseUrlOverride: "http://localhost/gsc");

        var result = await client.QuerySearchAnalyticsAsync(
            "https://www.devleader.ca/", "2025-01-01", "2025-12-31", null, 10);

        Assert.NotNull(result);
        Assert.Equal(3, callCount); // 403 + ListSites + retry
    }

    [Fact]
    public async Task ListSitemaps_On403_RetriesWithResolvedUrl()
    {
        var callCount = 0;
        var handler = new FakeMessageHandler(req =>
        {
            callCount++;
            var encodedPrefix = Uri.EscapeDataString("https://www.devleader.ca/");
            var encodedScDomain = Uri.EscapeDataString("sc-domain:devleader.ca");

            if (req.Method == HttpMethod.Get && req.RequestUri!.AbsoluteUri.Contains(encodedPrefix + "/sitemaps"))
                return Forbidden();

            if (req.Method == HttpMethod.Get && req.RequestUri!.AbsolutePath.TrimEnd('/').EndsWith("/sites"))
                return OkJson(new
                {
                    siteEntry = new[]
                    {
                        new { siteUrl = "sc-domain:devleader.ca", permissionLevel = "siteFullUser" }
                    }
                });

            if (req.Method == HttpMethod.Get && req.RequestUri!.AbsoluteUri.Contains(encodedScDomain + "/sitemaps"))
                return OkJson(new { sitemap = Array.Empty<object>() });

            return new HttpResponseMessage(HttpStatusCode.NotFound);
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler),
            new FakeTokenProvider(),
            baseUrlOverride: "http://localhost/gsc");

        var result = await client.ListSitemapsAsync("https://www.devleader.ca/");

        Assert.NotNull(result);
        Assert.Equal(3, callCount);
    }
}
