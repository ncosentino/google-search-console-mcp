using System.Net;
using System.Security.Cryptography;
using System.Text.Json;

namespace SearchConsoleMcp.Tests;

/// <summary>Always returns a fixed fake access token; bypasses real Google OAuth2.</summary>
internal sealed class FakeTokenProvider : SearchConsoleMcp.SearchConsole.ITokenProvider
{
    public Task<string> GetAccessTokenAsync(CancellationToken cancellationToken = default)
        => Task.FromResult("fake-token");
}

/// <summary>An <see cref="HttpMessageHandler"/> that answers every request via a caller-supplied
/// function instead of making a real network call, and counts how many requests it received.</summary>
internal sealed class FakeMessageHandler : HttpMessageHandler
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

/// <summary>Shared response-building helpers for tests using <see cref="FakeMessageHandler"/>.</summary>
internal static class FakeResponses
{
    internal static HttpResponseMessage OkJson(object body)
    {
        var json = JsonSerializer.Serialize(body);
        return new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new StringContent(json, System.Text.Encoding.UTF8, "application/json")
        };
    }

    internal static HttpResponseMessage Forbidden()
        => new HttpResponseMessage(HttpStatusCode.Forbidden)
        {
            Content = new StringContent(@"{""error"":{""code"":403}}", System.Text.Encoding.UTF8, "application/json")
        };

    internal static HttpResponseMessage ServerError(string body = @"{""error"":""boom""}")
        => new HttpResponseMessage(HttpStatusCode.InternalServerError)
        {
            Content = new StringContent(body, System.Text.Encoding.UTF8, "application/json")
        };
}

/// <summary>
/// Builds a throwaway (not a real Google credential) service account JSON payload
/// with a cryptographically valid, freshly generated RSA keypair -- enough for
/// GoogleServiceAccountAuth's RSA.ImportFromPem call to succeed at construction
/// time, without needing a real Google-issued key. The JWT this key ultimately
/// signs is never itself validated by a real Google endpoint in these tests: either
/// no network call is reached at all (tool listing only), or a fake
/// HttpMessageHandler intercepts the OAuth2 token exchange and returns a canned
/// response regardless of the JWT's actual signature.
/// </summary>
internal static class FakeServiceAccount
{
    internal static byte[] JsonBytes(
        string clientEmail = "test@test-project.iam.gserviceaccount.com",
        string projectId = "test-project",
        string privateKeyId = "test-key-id")
    {
        using var rsa = RSA.Create(2048);
        var json = JsonSerializer.Serialize(new
        {
            type = "service_account",
            project_id = projectId,
            private_key_id = privateKeyId,
            private_key = rsa.ExportRSAPrivateKeyPem(),
            client_email = clientEmail,
        });
        return System.Text.Encoding.UTF8.GetBytes(json);
    }
}
