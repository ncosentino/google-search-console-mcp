using System.Net;
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
