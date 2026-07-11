using SearchConsoleMcp.SearchConsole;
using Xunit;

namespace SearchConsoleMcp.Tests;

/// <summary>Tests for the search_type parameter added to QuerySearchAnalyticsAsync (issue #22).</summary>
public sealed class SearchConsoleClientSearchTypeTests
{
    /// <summary>
    /// RED-first test for issue #22: search_type "video" must be forwarded to the
    /// upstream Search Console API as the request body's "type" field.
    /// </summary>
    [Fact]
    public async Task QuerySearchAnalytics_SearchTypeVideo_SendsTypeVideoUpstream()
    {
        HttpRequestMessage? capturedRequest = null;
        var handler = new FakeMessageHandler(req =>
        {
            capturedRequest = req;
            return FakeResponses.OkJson(new { rows = Array.Empty<object>() });
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");

        await client.QuerySearchAnalyticsAsync(
            "devleader.ca", "2025-01-01", "2025-12-31", null, 10, searchType: "video");

        var body = await capturedRequest!.Content!.ReadAsStringAsync();
        Assert.Contains("\"type\":\"video\"", body, StringComparison.Ordinal);
    }

    /// <summary>
    /// Confirms that omitting search_type sends no "type" field at all, preserving
    /// the exact request shape that existed before #22 rather than silently
    /// rewriting every existing caller's request to add "type":"web". The response,
    /// however, still reports the effective default explicitly.
    /// </summary>
    [Fact]
    public async Task QuerySearchAnalytics_SearchTypeOmitted_PreservesExistingRequest()
    {
        HttpRequestMessage? capturedRequest = null;
        var handler = new FakeMessageHandler(req =>
        {
            capturedRequest = req;
            return FakeResponses.OkJson(new { rows = Array.Empty<object>() });
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");

        var result = await client.QuerySearchAnalyticsAsync(
            "devleader.ca", "2025-01-01", "2025-12-31", null, 10);

        var body = await capturedRequest!.Content!.ReadAsStringAsync();
        Assert.DoesNotContain("\"type\"", body, StringComparison.Ordinal);
        Assert.Equal("web", result.SearchType);
    }

    /// <summary>
    /// Confirms an explicit "web" is accepted and forwarded as-is -- distinct from
    /// omission, which sends no "type" field at all (see the preceding test).
    /// </summary>
    [Fact]
    public async Task QuerySearchAnalytics_SearchTypeWeb_SendsTypeWebExplicitly()
    {
        HttpRequestMessage? capturedRequest = null;
        var handler = new FakeMessageHandler(req =>
        {
            capturedRequest = req;
            return FakeResponses.OkJson(new { rows = Array.Empty<object>() });
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");

        var result = await client.QuerySearchAnalyticsAsync(
            "devleader.ca", "2025-01-01", "2025-12-31", null, 10, searchType: "web");

        var body = await capturedRequest!.Content!.ReadAsStringAsync();
        Assert.Contains("\"type\":\"web\"", body, StringComparison.Ordinal);
        Assert.Equal("web", result.SearchType);
    }

    /// <summary>
    /// Table-drives every upstream-supported value (per
    /// https://developers.google.com/webmaster-tools/v1/searchanalytics/query) to
    /// guard against silently narrowing the accepted set in a future refactor.
    /// </summary>
    [Theory]
    [InlineData("web")]
    [InlineData("image")]
    [InlineData("video")]
    [InlineData("news")]
    [InlineData("discover")]
    [InlineData("googleNews")]
    public async Task QuerySearchAnalytics_AllValidSearchTypes_AreAccepted(string searchType)
    {
        HttpRequestMessage? capturedRequest = null;
        var handler = new FakeMessageHandler(req =>
        {
            capturedRequest = req;
            return FakeResponses.OkJson(new { rows = Array.Empty<object>() });
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");

        var result = await client.QuerySearchAnalyticsAsync(
            "devleader.ca", "2025-01-01", "2025-12-31", null, 10, searchType);

        var body = await capturedRequest!.Content!.ReadAsStringAsync();
        Assert.Contains($"\"type\":\"{searchType}\"", body, StringComparison.Ordinal);
        Assert.Equal(searchType, result.SearchType);
    }

    /// <summary>
    /// Covers this issue's own motivating example (video-search performance grouped
    /// by page): search_type and dimensions are independent request fields, but this
    /// proves they compose correctly on the wire together rather than one silently
    /// overwriting or suppressing the other.
    /// </summary>
    [Fact]
    public async Task QuerySearchAnalytics_SearchTypeVideo_ComposesWithDimensions()
    {
        HttpRequestMessage? capturedRequest = null;
        var handler = new FakeMessageHandler(req =>
        {
            capturedRequest = req;
            return FakeResponses.OkJson(new { rows = Array.Empty<object>() });
        });

        var client = new SearchConsoleClient(
            new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");

        var result = await client.QuerySearchAnalyticsAsync(
            "devleader.ca", "2025-01-01", "2025-12-31", ["page"], 10, searchType: "video");

        var body = await capturedRequest!.Content!.ReadAsStringAsync();
        Assert.Contains("\"type\":\"video\"", body, StringComparison.Ordinal);
        Assert.Contains("\"dimensions\":[\"page\"]", body, StringComparison.Ordinal);
        Assert.Equal("video", result.SearchType);
    }

    /// <summary>
    /// Confirms an unsupported search_type value is rejected before any network
    /// call is made, rather than forwarded upstream to fail with a confusing API
    /// error.
    /// </summary>
    [Fact]
    public async Task QuerySearchAnalytics_InvalidSearchType_RejectedWithoutHTTPCall()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.OkJson(new { rows = Array.Empty<object>() }));
        var client = new SearchConsoleClient(
            new HttpClient(handler), new FakeTokenProvider(), baseUrlOverride: "http://localhost/gsc");

        await Assert.ThrowsAsync<InvalidOperationException>(() => client.QuerySearchAnalyticsAsync(
            "devleader.ca", "2025-01-01", "2025-12-31", null, 10, searchType: "youtube"));

        Assert.Equal(0, handler.CallCount);
    }
}
