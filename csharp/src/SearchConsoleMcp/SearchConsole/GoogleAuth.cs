using System.Net.Http.Json;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace SearchConsoleMcp.SearchConsole;

/// <summary>Manages Google OAuth2 access tokens obtained via service account JWT flow.</summary>
internal sealed class GoogleServiceAccountAuth
{
    private const string TokenUri = "https://oauth2.googleapis.com/token";

    private readonly string _clientEmail;
    private readonly RSA _privateKey;
    private readonly string _privateKeyId;
    private readonly string _scope;
    private readonly HttpClient _httpClient;

    private string? _cachedToken;
    private DateTimeOffset _tokenExpiry = DateTimeOffset.MinValue;

    internal GoogleServiceAccountAuth(
        string clientEmail,
        string privateKeyPem,
        string privateKeyId,
        string scope,
        HttpClient httpClient)
    {
        _clientEmail = clientEmail;
        _privateKeyId = privateKeyId;
        _scope = scope;
        _httpClient = httpClient;

        _privateKey = RSA.Create();
        _privateKey.ImportFromPem(privateKeyPem);
    }

    /// <summary>Returns a valid Bearer token, refreshing if necessary.</summary>
    internal async Task<string> GetAccessTokenAsync(CancellationToken cancellationToken = default)
    {
        if (_cachedToken is not null && DateTimeOffset.UtcNow < _tokenExpiry.AddMinutes(-1))
            return _cachedToken;

        var token = await FetchTokenAsync(cancellationToken).ConfigureAwait(false);
        _cachedToken = token.AccessToken;
        _tokenExpiry = DateTimeOffset.UtcNow.AddSeconds(token.ExpiresIn);
        return _cachedToken;
    }

    private async Task<TokenResponse> FetchTokenAsync(CancellationToken cancellationToken)
    {
        var jwt = BuildJwt();

        using var content = new FormUrlEncodedContent(new Dictionary<string, string>
        {
            ["grant_type"] = "urn:ietf:params:oauth:grant-type:jwt-bearer",
            ["assertion"] = jwt
        });

        using var response = await _httpClient.PostAsync(TokenUri, content, cancellationToken)
            .ConfigureAwait(false);

        if (!response.IsSuccessStatusCode)
        {
            var body = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
            throw new InvalidOperationException(
                $"Failed to obtain access token: HTTP {(int)response.StatusCode} -- {body[..Math.Min(body.Length, 300)]}");
        }

        var result = await response.Content
            .ReadFromJsonAsync(GscJsonContext.Default.TokenResponse, cancellationToken)
            .ConfigureAwait(false);

        return result ?? throw new InvalidOperationException("Empty token response from Google");
    }

    private string BuildJwt()
    {
        var now = DateTimeOffset.UtcNow.ToUnixTimeSeconds();

        var header = Base64UrlEncode(JsonSerializer.SerializeToUtf8Bytes(
            new JwtHeader { Alg = "RS256", Typ = "JWT", Kid = _privateKeyId },
            GscJsonContext.Default.JwtHeader));

        var payload = Base64UrlEncode(JsonSerializer.SerializeToUtf8Bytes(
            new JwtPayload
            {
                Iss = _clientEmail,
                Sub = _clientEmail,
                Scope = _scope,
                Aud = TokenUri,
                Iat = now,
                Exp = now + 3600
            },
            GscJsonContext.Default.JwtPayload));

        var signingInput = Encoding.ASCII.GetBytes($"{header}.{payload}");
        var signature = _privateKey.SignData(signingInput, HashAlgorithmName.SHA256, RSASignaturePadding.Pkcs1);

        return $"{header}.{payload}.{Base64UrlEncode(signature)}";
    }

    private static string Base64UrlEncode(byte[] data) =>
        Convert.ToBase64String(data)
            .TrimEnd('=')
            .Replace('+', '-')
            .Replace('/', '_');
}

internal sealed class TokenResponse
{
    [JsonPropertyName("access_token")]
    public string AccessToken { get; set; } = string.Empty;

    [JsonPropertyName("expires_in")]
    public int ExpiresIn { get; set; }
}

internal sealed class JwtHeader
{
    [JsonPropertyName("alg")]
    public string Alg { get; set; } = string.Empty;

    [JsonPropertyName("typ")]
    public string Typ { get; set; } = string.Empty;

    [JsonPropertyName("kid")]
    public string Kid { get; set; } = string.Empty;
}

internal sealed class JwtPayload
{
    [JsonPropertyName("iss")]
    public string Iss { get; set; } = string.Empty;

    [JsonPropertyName("sub")]
    public string Sub { get; set; } = string.Empty;

    [JsonPropertyName("scope")]
    public string Scope { get; set; } = string.Empty;

    [JsonPropertyName("aud")]
    public string Aud { get; set; } = string.Empty;

    [JsonPropertyName("iat")]
    public long Iat { get; set; }

    [JsonPropertyName("exp")]
    public long Exp { get; set; }
}
