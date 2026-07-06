using System.Security.Cryptography;
using SearchConsoleMcp.SearchConsole;
using Xunit;

namespace SearchConsoleMcp.Tests;

/// <summary>
/// Characterization tests for GoogleServiceAccountAuth's JWT-bearer token fetch and
/// caching logic, written ahead of the ModelContextProtocol SDK dependency
/// modernization (issue #6). Before these tests, this class was at 0% coverage.
/// </summary>
public sealed class GoogleServiceAccountAuthTests
{
    private static string GenerateTestPrivateKeyPem()
    {
        using var rsa = RSA.Create(2048);
        return rsa.ExportRSAPrivateKeyPem();
    }

    [Fact]
    public async Task GetAccessTokenAsync_Success_ReturnsTokenFromResponse()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.OkJson(new
        {
            access_token = "real-looking-token",
            expires_in = 3600
        }));
        var auth = new GoogleServiceAccountAuth(
            "test@example.iam.gserviceaccount.com",
            GenerateTestPrivateKeyPem(),
            "key-id-123",
            "https://www.googleapis.com/auth/webmasters.readonly",
            new HttpClient(handler));

        var token = await auth.GetAccessTokenAsync();

        Assert.Equal("real-looking-token", token);
        Assert.Equal(1, handler.CallCount);
    }

    [Fact]
    public async Task GetAccessTokenAsync_CalledTwice_UsesCacheWithoutRefetching()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.OkJson(new
        {
            access_token = "cached-token",
            expires_in = 3600
        }));
        var auth = new GoogleServiceAccountAuth(
            "test@example.iam.gserviceaccount.com",
            GenerateTestPrivateKeyPem(),
            "key-id-123",
            "https://www.googleapis.com/auth/webmasters.readonly",
            new HttpClient(handler));

        var first = await auth.GetAccessTokenAsync();
        var second = await auth.GetAccessTokenAsync();

        Assert.Equal(first, second);
        Assert.Equal(1, handler.CallCount); // second call must be served from cache
    }

    [Fact]
    public async Task GetAccessTokenAsync_TokenEndpointFails_ThrowsWithStatusAndBody()
    {
        var handler = new FakeMessageHandler(_ => FakeResponses.ServerError(@"{""error"":""invalid_grant""}"));
        var auth = new GoogleServiceAccountAuth(
            "test@example.iam.gserviceaccount.com",
            GenerateTestPrivateKeyPem(),
            "key-id-123",
            "https://www.googleapis.com/auth/webmasters.readonly",
            new HttpClient(handler));

        var ex = await Assert.ThrowsAsync<InvalidOperationException>(() => auth.GetAccessTokenAsync());
        Assert.Contains("invalid_grant", ex.Message, StringComparison.Ordinal);
    }
}
