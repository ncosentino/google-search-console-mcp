using System.Net.Http.Json;
using System.Text;
using System.Text.Json;

namespace SearchConsoleMcp.SearchConsole;

/// <summary>Client for the Google Search Console API v3.</summary>
internal sealed class SearchConsoleClient
{
    private const string BaseUrl = "https://www.googleapis.com/webmasters/v3";
    private const string GscScope = "https://www.googleapis.com/auth/webmasters.readonly";

    private readonly HttpClient _httpClient;
    private readonly GoogleServiceAccountAuth _auth;

    /// <summary>Creates a <see cref="SearchConsoleClient"/> from raw service account JSON bytes.</summary>
    internal static SearchConsoleClient Create(byte[] serviceAccountJson, HttpClient httpClient)
    {
        var sa = JsonSerializer.Deserialize(serviceAccountJson, GscJsonContext.Default.ServiceAccountJson)
            ?? throw new InvalidOperationException("Service account JSON could not be parsed.");

        if (string.IsNullOrWhiteSpace(sa.ClientEmail))
            throw new InvalidOperationException("Service account JSON is missing 'client_email'.");
        if (string.IsNullOrWhiteSpace(sa.PrivateKey))
            throw new InvalidOperationException("Service account JSON is missing 'private_key'.");

        var auth = new GoogleServiceAccountAuth(
            sa.ClientEmail, sa.PrivateKey, sa.PrivateKeyId, GscScope, httpClient);

        return new SearchConsoleClient(httpClient, auth);
    }

    private SearchConsoleClient(HttpClient httpClient, GoogleServiceAccountAuth auth)
    {
        _httpClient = httpClient;
        _auth = auth;
    }

    /// <summary>Lists all Search Console properties accessible to the service account.</summary>
    internal async Task<SiteListResponse> ListSitesAsync(CancellationToken cancellationToken = default)
    {
        var request = await BuildRequestAsync(HttpMethod.Get, $"{BaseUrl}/sites", null, cancellationToken)
            .ConfigureAwait(false);

        using var response = await _httpClient.SendAsync(request, cancellationToken).ConfigureAwait(false);
        await EnsureSuccessAsync(response, cancellationToken).ConfigureAwait(false);

        var raw = await response.Content
            .ReadFromJsonAsync(GscJsonContext.Default.ApiSiteListResponse, cancellationToken)
            .ConfigureAwait(false);

        var sites = (raw?.SiteEntry ?? [])
            .Select(s => new SiteEntry(s.SiteUrl, s.PermissionLevel))
            .ToList();

        return new SiteListResponse(sites, DateTimeOffset.UtcNow);
    }

    /// <summary>Lists sitemaps submitted to Search Console for the given property.</summary>
    internal async Task<SitemapListResponse> ListSitemapsAsync(
        string siteUrl,
        CancellationToken cancellationToken = default)
    {
        var encodedSite = Uri.EscapeDataString(siteUrl);
        var request = await BuildRequestAsync(
            HttpMethod.Get, $"{BaseUrl}/sites/{encodedSite}/sitemaps", null, cancellationToken)
            .ConfigureAwait(false);

        using var response = await _httpClient.SendAsync(request, cancellationToken).ConfigureAwait(false);
        await EnsureSuccessAsync(response, cancellationToken).ConfigureAwait(false);

        var raw = await response.Content
            .ReadFromJsonAsync(GscJsonContext.Default.ApiSitemapListResponse, cancellationToken)
            .ConfigureAwait(false);

        var sitemaps = (raw?.Sitemap ?? []).Select(s => new SitemapEntry(
            s.Path,
            DateTimeOffset.TryParse(s.LastSubmitted, out var ls) ? ls : null,
            s.IsPending,
            s.IsSitemapsIndex,
            s.Type,
            DateTimeOffset.TryParse(s.LastDownloaded, out var ld) ? ld : null,
            s.Warnings,
            s.Errors)).ToList();

        return new SitemapListResponse(siteUrl, sitemaps, DateTimeOffset.UtcNow);
    }

    /// <summary>Queries search analytics data for the given property.</summary>
    internal async Task<SearchAnalyticsResponse> QuerySearchAnalyticsAsync(
        string siteUrl,
        string startDate,
        string endDate,
        IReadOnlyList<string>? dimensions,
        int rowLimit,
        CancellationToken cancellationToken = default)
    {
        if (rowLimit <= 0) rowLimit = 1000;

        var body = new SearchAnalyticsRequest
        {
            StartDate = startDate,
            EndDate = endDate,
            Dimensions = dimensions?.Count > 0 ? dimensions : null,
            RowLimit = rowLimit
        };

        var json = JsonSerializer.Serialize(body, GscJsonContext.Default.SearchAnalyticsRequest);
        var encodedSite = Uri.EscapeDataString(siteUrl);
        var request = await BuildRequestAsync(
            HttpMethod.Post,
            $"{BaseUrl}/sites/{encodedSite}/searchAnalytics/query",
            json,
            cancellationToken).ConfigureAwait(false);

        using var response = await _httpClient.SendAsync(request, cancellationToken).ConfigureAwait(false);
        await EnsureSuccessAsync(response, cancellationToken).ConfigureAwait(false);

        var raw = await response.Content
            .ReadFromJsonAsync(GscJsonContext.Default.ApiSearchAnalyticsResponse, cancellationToken)
            .ConfigureAwait(false);

        var rows = (raw?.Rows ?? [])
            .Select(r => new SearchAnalyticsRow(r.Keys, r.Clicks, r.Impressions, r.Ctr, r.Position))
            .ToList();

        return new SearchAnalyticsResponse(siteUrl, startDate, endDate, dimensions, rows.Count, rows, DateTimeOffset.UtcNow);
    }

    private async Task<HttpRequestMessage> BuildRequestAsync(
        HttpMethod method,
        string url,
        string? jsonBody,
        CancellationToken cancellationToken)
    {
        var token = await _auth.GetAccessTokenAsync(cancellationToken).ConfigureAwait(false);
        var request = new HttpRequestMessage(method, url);
        request.Headers.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", token);

        if (jsonBody is not null)
            request.Content = new StringContent(jsonBody, Encoding.UTF8, "application/json");

        return request;
    }

    private static async Task EnsureSuccessAsync(HttpResponseMessage response, CancellationToken cancellationToken)
    {
        if (response.IsSuccessStatusCode)
            return;

        var body = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
        var snippet = body.Length > 300 ? body[..300] + "..." : body;
        throw new InvalidOperationException(
            $"Search Console API returned HTTP {(int)response.StatusCode} {response.StatusCode}: {snippet}");
    }
}
