using System.Net.Http.Json;
using System.Text;
using System.Text.Json;

namespace SearchConsoleMcp.SearchConsole;

/// <summary>Provides OAuth2 Bearer tokens for authenticating Google API requests.</summary>
internal interface ITokenProvider
{
    Task<string> GetAccessTokenAsync(CancellationToken cancellationToken = default);
}

/// <summary>Thrown when the Search Console API returns a non-success status code.</summary>
internal sealed class GscApiException : Exception
{
    internal int StatusCode { get; }

    internal GscApiException(int statusCode, string message) : base(message)
        => StatusCode = statusCode;
}

/// <summary>Client for Google Search Console APIs.</summary>
internal sealed class SearchConsoleClient
{
    private const string DefaultBaseUrl = "https://www.googleapis.com/webmasters/v3";
    private const string DefaultUrlInspectionBaseUrl = "https://searchconsole.googleapis.com/v1";
    private const string GscScope = "https://www.googleapis.com/auth/webmasters.readonly";
    private const string DefaultSearchType = "web";
    private const string DefaultLanguageCode = "en-US";

    private static readonly HashSet<string> ValidSearchTypes = new(StringComparer.Ordinal)
    {
        "web", "image", "video", "news", "discover", "googleNews"
    };

    private readonly string _baseUrl;
    private readonly string _urlInspectionBaseUrl;
    private readonly HttpClient _httpClient;
    private readonly ITokenProvider _tokenProvider;

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

    /// <summary>Creates a client with custom authentication and optional API URL overrides.</summary>
    internal SearchConsoleClient(
        HttpClient httpClient,
        ITokenProvider tokenProvider,
        string? baseUrlOverride = null,
        string? urlInspectionBaseUrlOverride = null)
    {
        _httpClient = httpClient;
        _tokenProvider = tokenProvider;
        _baseUrl = baseUrlOverride ?? DefaultBaseUrl;
        _urlInspectionBaseUrl =
            urlInspectionBaseUrlOverride ?? baseUrlOverride ?? DefaultUrlInspectionBaseUrl;
    }

    /// <summary>Lists all Search Console properties accessible to the service account.</summary>
    internal async Task<SiteListResponse> ListSitesAsync(CancellationToken cancellationToken = default)
    {
        var request = await BuildRequestAsync(HttpMethod.Get, $"{_baseUrl}/sites", null, cancellationToken)
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
        return await ExecuteWithResolvedSiteUrlAsync(
            siteUrl,
            ListSitemapsInternalAsync,
            cancellationToken).ConfigureAwait(false);
    }

    private async Task<SitemapListResponse> ListSitemapsInternalAsync(
        string siteUrl,
        CancellationToken cancellationToken)
    {
        var encodedSite = Uri.EscapeDataString(siteUrl);
        var request = await BuildRequestAsync(
            HttpMethod.Get, $"{_baseUrl}/sites/{encodedSite}/sitemaps", null, cancellationToken)
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
        string? searchType = null,
        CancellationToken cancellationToken = default)
    {
        if (!string.IsNullOrEmpty(searchType) && !ValidSearchTypes.Contains(searchType))
        {
            throw new InvalidOperationException(
                $"invalid search_type '{searchType}': must be one of web, image, video, news, discover, googleNews");
        }

        return await ExecuteWithResolvedSiteUrlAsync(
            siteUrl,
            (resolved, token) => QuerySearchAnalyticsInternalAsync(
                resolved, startDate, endDate, dimensions, rowLimit, searchType, token),
            cancellationToken).ConfigureAwait(false);
    }

    private async Task<SearchAnalyticsResponse> QuerySearchAnalyticsInternalAsync(
        string siteUrl,
        string startDate,
        string endDate,
        IReadOnlyList<string>? dimensions,
        int rowLimit,
        string? searchType,
        CancellationToken cancellationToken)
    {
        if (rowLimit <= 0) rowLimit = 1000;

        var body = new SearchAnalyticsRequest
        {
            StartDate = startDate,
            EndDate = endDate,
            Dimensions = dimensions?.Count > 0 ? dimensions : null,
            Type = string.IsNullOrEmpty(searchType) ? null : searchType,
            RowLimit = rowLimit
        };

        var json = JsonSerializer.Serialize(body, GscJsonContext.Default.SearchAnalyticsRequest);
        var encodedSite = Uri.EscapeDataString(siteUrl);
        var request = await BuildRequestAsync(
            HttpMethod.Post,
            $"{_baseUrl}/sites/{encodedSite}/searchAnalytics/query",
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

        var effectiveSearchType = string.IsNullOrEmpty(searchType) ? DefaultSearchType : searchType;

        return new SearchAnalyticsResponse(
            siteUrl, startDate, endDate, dimensions, effectiveSearchType, rows.Count, rows, DateTimeOffset.UtcNow);
    }

    /// <summary>Inspects Google's indexed version of one URL under a Search Console property.</summary>
    internal async Task<UrlInspectionResponse> InspectUrlAsync(
        string siteUrl,
        string inspectionUrl,
        string? languageCode = null,
        CancellationToken cancellationToken = default)
    {
        inspectionUrl = inspectionUrl.Trim();
        if (inspectionUrl.Length == 0)
            throw new ArgumentException("inspection_url is required", nameof(inspectionUrl));

        var effectiveLanguageCode = string.IsNullOrWhiteSpace(languageCode)
            ? DefaultLanguageCode
            : languageCode.Trim();

        return await ExecuteWithResolvedSiteUrlAsync(
            siteUrl,
            (resolved, token) => InspectUrlInternalAsync(
                resolved, inspectionUrl, effectiveLanguageCode, token),
            cancellationToken).ConfigureAwait(false);
    }

    private async Task<UrlInspectionResponse> InspectUrlInternalAsync(
        string siteUrl,
        string inspectionUrl,
        string languageCode,
        CancellationToken cancellationToken)
    {
        var body = new UrlInspectionRequest
        {
            SiteUrl = siteUrl,
            InspectionUrl = inspectionUrl,
            LanguageCode = languageCode
        };
        var json = JsonSerializer.Serialize(body, GscJsonContext.Default.UrlInspectionRequest);
        var request = await BuildRequestAsync(
            HttpMethod.Post,
            $"{_urlInspectionBaseUrl}/urlInspection/index:inspect",
            json,
            cancellationToken).ConfigureAwait(false);

        using var response = await _httpClient.SendAsync(request, cancellationToken).ConfigureAwait(false);
        await EnsureSuccessAsync(response, cancellationToken).ConfigureAwait(false);

        var raw = await response.Content
            .ReadFromJsonAsync(GscJsonContext.Default.ApiUrlInspectionResponse, cancellationToken)
            .ConfigureAwait(false);
        if (raw is null ||
            raw.InspectionResult.ValueKind is JsonValueKind.Undefined or JsonValueKind.Null)
        {
            throw new InvalidOperationException(
                "Search Console URL Inspection response did not include inspectionResult.");
        }

        return new UrlInspectionResponse(
            siteUrl,
            inspectionUrl,
            languageCode,
            raw.InspectionResult.Clone(),
            DateTimeOffset.UtcNow);
    }

    private async Task<T> ExecuteWithResolvedSiteUrlAsync<T>(
        string input,
        Func<string, CancellationToken, Task<T>> operation,
        CancellationToken cancellationToken)
    {
        var resolved = SiteUrlResolver.Normalize(input);
        try
        {
            return await operation(resolved, cancellationToken).ConfigureAwait(false);
        }
        catch (GscApiException ex) when (ex.StatusCode == 403)
        {
            var resolvedUrl = await ResolveSiteUrlAsync(input, cancellationToken).ConfigureAwait(false);
            return await operation(resolvedUrl, cancellationToken).ConfigureAwait(false);
        }
    }

    private async Task<string> ResolveSiteUrlAsync(string input, CancellationToken cancellationToken)
    {
        var sites = await ListSitesAsync(cancellationToken).ConfigureAwait(false);
        return SiteUrlResolver.FindBestMatch(input, sites.Sites)
            ?? throw new InvalidOperationException(
                $"No accessible GSC property found for '{input}'. " +
                $"Accessible properties: {string.Join(", ", sites.Sites.Select(s => s.SiteUrl))}");
    }

    private async Task<HttpRequestMessage> BuildRequestAsync(
        HttpMethod method,
        string url,
        string? jsonBody,
        CancellationToken cancellationToken)
    {
        var token = await _tokenProvider.GetAccessTokenAsync(cancellationToken).ConfigureAwait(false);
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
        throw new GscApiException(
            (int)response.StatusCode,
            $"Search Console API returned HTTP {(int)response.StatusCode} {response.StatusCode}: {snippet}");
    }
}
