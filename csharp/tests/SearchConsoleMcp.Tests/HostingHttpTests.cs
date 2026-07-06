using System.Net;
using System.Text;
using Microsoft.AspNetCore.Builder;
using ModelContextProtocol.Client;
using Xunit;

namespace SearchConsoleMcp.Tests;

/// <summary>
/// Tests for the HTTP transport host (Hosting.BuildHttpHost), exercising a real
/// bound Kestrel instance and a real MCP client connecting over HTTP -- the same
/// wiring Program.cs uses for --transport http, minus argument parsing.
/// </summary>
public sealed class HostingHttpTests
{
    private static readonly byte[] FakeServiceAccountJson = FakeServiceAccount.JsonBytes();

    /// <summary>
    /// Fake handler that satisfies the OAuth2 token exchange with a canned success
    /// response, then returns a configurable body for the Search Console API call
    /// itself. Used to exercise query_search_analytics/list_sites/list_sitemaps
    /// end-to-end through a real MCP session without making any real network call.
    /// </summary>
    private sealed class FakeSearchConsoleHandler(string apiResponseBody) : HttpMessageHandler
    {
        protected override Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request, CancellationToken cancellationToken)
        {
            if (request.RequestUri?.Host == "oauth2.googleapis.com")
            {
                const string tokenJson = """{"access_token":"fake-token","expires_in":3600}""";
                return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
                {
                    Content = new StringContent(tokenJson, Encoding.UTF8, "application/json"),
                });
            }

            return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(apiResponseBody, Encoding.UTF8, "application/json"),
            });
        }
    }

    /// <summary>
    /// Hosting.BuildHttpHost binds "0.0.0.0" (all interfaces), which is correct for
    /// production but isn't itself a connectable target address -- app.Urls reports
    /// back the bind address verbatim. Tests connecting from the same machine need
    /// to target loopback explicitly, on whatever port Kestrel actually chose.
    /// </summary>
    private static Uri ConnectableUri(WebApplication app)
    {
        var bound = new Uri(app.Urls.First());
        return new UriBuilder(bound) { Host = "127.0.0.1" }.Uri;
    }

    [Fact]
    public async Task BuildHttpHost_ServesRealSession_ListsAllTools()
    {
        await using var app = Hosting.BuildHttpHost([], FakeServiceAccountJson, port: 0);
        await app.StartAsync();
        try
        {
            await using var client = await McpClient.CreateAsync(new HttpClientTransport(new HttpClientTransportOptions
            {
                Endpoint = ConnectableUri(app),
            }));

            var tools = await client.ListToolsAsync();

            Assert.Equal(3, tools.Count);
            Assert.Contains(tools, t => t.Name == "query_search_analytics");
            Assert.Contains(tools, t => t.Name == "list_sites");
            Assert.Contains(tools, t => t.Name == "list_sitemaps");
        }
        finally
        {
            await app.StopAsync();
        }
    }

    [Fact]
    public async Task BuildHttpHost_NoAllowedHostsConfigured_DefaultsToLoopback()
    {
        await using var app = Hosting.BuildHttpHost([], FakeServiceAccountJson, port: 0);

        Assert.Equal(Hosting.DefaultAllowedHosts, app.Configuration["AllowedHosts"]);
    }

    [Fact]
    public async Task BuildHttpHost_AllowedHostsPassedOnCommandLine_OverridesDefault()
    {
        await using var app = Hosting.BuildHttpHost(
            ["--AllowedHosts", "example.com"], FakeServiceAccountJson, port: 0);

        Assert.Equal("example.com", app.Configuration["AllowedHosts"]);
    }

    /// <summary>
    /// Confirms list_sites -- previously exercised only via direct C# method calls in
    /// SearchConsoleToolTests, never through a real MCP session -- works end-to-end:
    /// [McpServerTool] reflection-based dispatch and the DI-registered
    /// SearchConsoleClient all have to agree for this to pass.
    /// </summary>
    [Fact]
    public async Task BuildHttpHost_CallListSitesTool_ViaRealSession_ReturnsSuccessResult()
    {
        var handler = new FakeSearchConsoleHandler("""{"siteEntry":[]}""");
        await using var app = Hosting.BuildHttpHost([], FakeServiceAccountJson, port: 0, handler);
        await app.StartAsync();
        try
        {
            await using var client = await McpClient.CreateAsync(new HttpClientTransport(new HttpClientTransportOptions
            {
                Endpoint = ConnectableUri(app),
            }));

            var result = await client.CallToolAsync("list_sites", new Dictionary<string, object?>());

            // IsError is nullable: null (unset) on success, true on error.
            Assert.NotEqual(true, result.IsError);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>
    /// Confirms query_search_analytics works end-to-end via a real MCP session with
    /// dimensions omitted entirely -- exercising both the HTTP transport wiring and
    /// the dimensions-is-optional parity fix from #7 together.
    /// </summary>
    [Fact]
    public async Task BuildHttpHost_CallQuerySearchAnalyticsTool_ViaRealSession_DimensionsOmitted_ReturnsSuccessResult()
    {
        var handler = new FakeSearchConsoleHandler("""{"rows":[]}""");
        await using var app = Hosting.BuildHttpHost([], FakeServiceAccountJson, port: 0, handler);
        await app.StartAsync();
        try
        {
            await using var client = await McpClient.CreateAsync(new HttpClientTransport(new HttpClientTransportOptions
            {
                Endpoint = ConnectableUri(app),
            }));

            var result = await client.CallToolAsync("query_search_analytics", new Dictionary<string, object?>
            {
                ["site_url"] = "devleader.ca",
                ["start_date"] = "2025-01-01",
                ["end_date"] = "2025-12-31",
            });

            Assert.NotEqual(true, result.IsError);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>
    /// Confirms query_search_analytics works end-to-end via a real MCP session when
    /// dimensions arrives JSON-encoded as a string instead of a genuine array --
    /// exercising both the HTTP transport wiring and the StringifiedArgsCoercion
    /// filter from #7 together, through a real request filter pipeline rather than a
    /// unit test of the coercion function alone.
    /// </summary>
    [Fact]
    public async Task BuildHttpHost_CallQuerySearchAnalyticsTool_ViaRealSession_StringifiedDimensions_ReturnsSuccessResult()
    {
        var handler = new FakeSearchConsoleHandler("""{"rows":[]}""");
        await using var app = Hosting.BuildHttpHost([], FakeServiceAccountJson, port: 0, handler);
        await app.StartAsync();
        try
        {
            await using var client = await McpClient.CreateAsync(new HttpClientTransport(new HttpClientTransportOptions
            {
                Endpoint = ConnectableUri(app),
            }));

            var result = await client.CallToolAsync("query_search_analytics", new Dictionary<string, object?>
            {
                ["site_url"] = "devleader.ca",
                ["start_date"] = "2025-01-01",
                ["end_date"] = "2025-12-31",
                ["dimensions"] = """["query"]""",
            });

            Assert.NotEqual(true, result.IsError);
        }
        finally
        {
            await app.StopAsync();
        }
    }
}
