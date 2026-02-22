using System.Text.Json.Serialization;

namespace SearchConsoleMcp.SearchConsole;

// --- Tool result models ---

/// <summary>A single row from a search analytics query.</summary>
internal sealed record SearchAnalyticsRow(
    IReadOnlyList<string>? Keys,
    double Clicks,
    double Impressions,
    double Ctr,
    double Position);

/// <summary>The result of a search analytics query.</summary>
internal sealed record SearchAnalyticsResponse(
    string SiteUrl,
    string StartDate,
    string EndDate,
    IReadOnlyList<string>? Dimensions,
    int RowCount,
    IReadOnlyList<SearchAnalyticsRow> Rows,
    DateTimeOffset QueriedAt);

/// <summary>A single Search Console property.</summary>
internal sealed record SiteEntry(string SiteUrl, string PermissionLevel);

/// <summary>The result of listing Search Console properties.</summary>
internal sealed record SiteListResponse(IReadOnlyList<SiteEntry> Sites, DateTimeOffset QueriedAt);

/// <summary>A single submitted sitemap.</summary>
internal sealed record SitemapEntry(
    string Path,
    DateTimeOffset? LastSubmitted,
    bool IsPending,
    bool IsSitemapsIndex,
    string Type,
    DateTimeOffset? LastDownloaded,
    long Warnings,
    long Errors);

/// <summary>The result of listing sitemaps for a property.</summary>
internal sealed record SitemapListResponse(
    string SiteUrl,
    IReadOnlyList<SitemapEntry> Sitemaps,
    DateTimeOffset QueriedAt);

// --- Raw Google API response models ---

internal sealed class ApiSiteEntry
{
    [JsonPropertyName("siteUrl")]
    public string SiteUrl { get; set; } = string.Empty;

    [JsonPropertyName("permissionLevel")]
    public string PermissionLevel { get; set; } = string.Empty;
}

internal sealed class ApiSiteListResponse
{
    [JsonPropertyName("siteEntry")]
    public List<ApiSiteEntry>? SiteEntry { get; set; }
}

internal sealed class ApiSitemapEntry
{
    [JsonPropertyName("path")]
    public string Path { get; set; } = string.Empty;

    [JsonPropertyName("lastSubmitted")]
    public string? LastSubmitted { get; set; }

    [JsonPropertyName("isPending")]
    public bool IsPending { get; set; }

    [JsonPropertyName("isSitemapsIndex")]
    public bool IsSitemapsIndex { get; set; }

    [JsonPropertyName("type")]
    public string Type { get; set; } = string.Empty;

    [JsonPropertyName("lastDownloaded")]
    public string? LastDownloaded { get; set; }

    [JsonPropertyName("warnings")]
    public long Warnings { get; set; }

    [JsonPropertyName("errors")]
    public long Errors { get; set; }
}

internal sealed class ApiSitemapListResponse
{
    [JsonPropertyName("sitemap")]
    public List<ApiSitemapEntry>? Sitemap { get; set; }
}

internal sealed class ApiSearchAnalyticsRow
{
    [JsonPropertyName("keys")]
    public List<string>? Keys { get; set; }

    [JsonPropertyName("clicks")]
    public double Clicks { get; set; }

    [JsonPropertyName("impressions")]
    public double Impressions { get; set; }

    [JsonPropertyName("ctr")]
    public double Ctr { get; set; }

    [JsonPropertyName("position")]
    public double Position { get; set; }
}

internal sealed class ApiSearchAnalyticsResponse
{
    [JsonPropertyName("rows")]
    public List<ApiSearchAnalyticsRow>? Rows { get; set; }
}

/// <summary>An error result returned when an API call fails.</summary>
internal sealed record ErrorResult(string Error);

/// <summary>Request body for search analytics queries.</summary>
internal sealed class SearchAnalyticsRequest
{
    [JsonPropertyName("startDate")]
    public string StartDate { get; set; } = string.Empty;

    [JsonPropertyName("endDate")]
    public string EndDate { get; set; } = string.Empty;

    [JsonPropertyName("dimensions")]
    public IReadOnlyList<string>? Dimensions { get; set; }

    [JsonPropertyName("rowLimit")]
    public int RowLimit { get; set; }
}

internal sealed class ServiceAccountJson
{
    [JsonPropertyName("type")]
    public string Type { get; set; } = string.Empty;

    [JsonPropertyName("project_id")]
    public string ProjectId { get; set; } = string.Empty;

    [JsonPropertyName("private_key_id")]
    public string PrivateKeyId { get; set; } = string.Empty;

    [JsonPropertyName("private_key")]
    public string PrivateKey { get; set; } = string.Empty;

    [JsonPropertyName("client_email")]
    public string ClientEmail { get; set; } = string.Empty;

    [JsonPropertyName("token_uri")]
    public string TokenUri { get; set; } = string.Empty;
}

/// <summary>System.Text.Json source generation context for AOT-safe serialization.</summary>
[JsonSerializable(typeof(ServiceAccountJson))]
[JsonSerializable(typeof(ApiSiteListResponse))]
[JsonSerializable(typeof(ApiSitemapListResponse))]
[JsonSerializable(typeof(ApiSearchAnalyticsResponse))]
[JsonSerializable(typeof(SearchAnalyticsRequest))]
[JsonSerializable(typeof(TokenResponse))]
[JsonSerializable(typeof(JwtHeader))]
[JsonSerializable(typeof(JwtPayload))]
[JsonSerializable(typeof(SearchAnalyticsResponse))]
[JsonSerializable(typeof(SiteListResponse))]
[JsonSerializable(typeof(SitemapListResponse))]
[JsonSerializable(typeof(ErrorResult))]
[JsonSourceGenerationOptions(
    PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
    WriteIndented = false,
    DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull)]
internal partial class GscJsonContext : JsonSerializerContext;
