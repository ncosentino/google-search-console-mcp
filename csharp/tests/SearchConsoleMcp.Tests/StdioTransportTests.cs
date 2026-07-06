using ModelContextProtocol.Client;
using SearchConsoleMcp.SearchConsole;
using Xunit;

namespace SearchConsoleMcp.Tests;

/// <summary>
/// Characterization test for the default (and, as of this writing, only) transport
/// path, written ahead of the ModelContextProtocol SDK dependency modernization
/// (issue #6).
/// </summary>
/// <remarks>
/// Program.cs's stdio branch (AddMcpServer().WithStdioServerTransport()) binds
/// directly to the real, process-wide Console.In/Console.Out, and the SDK offers no
/// overload that accepts a substitute stream. So the only faithful way to exercise
/// this path is to spawn the real compiled server as a subprocess and connect a real
/// MCP client over its actual stdin/stdout -- the same way Claude Desktop, Claude
/// Code, and other stdio-based MCP clients launch this server in production.
///
/// Unlike google-keyword-planner-mcp's equivalent test, this repo's credentials are a
/// full service-account JSON requiring a cryptographically valid RSA private key
/// (GoogleServiceAccountAuth calls RSA.ImportFromPem on it at construction time), and
/// every tool call goes straight to a live Google endpoint with no local
/// validation short-circuit to exploit for a network-free error path. So this test
/// generates a throwaway (not a real Google credential) but syntactically and
/// cryptographically valid RSA keypair -- enough for the server to start and list its
/// tools without touching the network -- and deliberately does not attempt a
/// CallTool, since every tool here would need a real, reachable Google endpoint to
/// succeed or fail meaningfully.
/// </remarks>
public sealed class StdioTransportTests
{
    [Fact]
    public async Task StdioTransport_ServesRealSession_ListsTools()
    {
        var serverDllPath = typeof(SearchConsoleClient).Assembly.Location;

        var serviceAccountJson = System.Text.Encoding.UTF8.GetString(FakeServiceAccount.JsonBytes());

        await using var client = await McpClient.CreateAsync(new StdioClientTransport(new StdioClientTransportOptions
        {
            Name = "gsc-mcp-stdio-test",
            Command = "dotnet",
            Arguments = [serverDllPath],
            EnvironmentVariables = new Dictionary<string, string?>
            {
                ["GOOGLE_SERVICE_ACCOUNT_JSON"] = serviceAccountJson,
                ["GOOGLE_SERVICE_ACCOUNT_FILE"] = "",
            },
        }));

        var tools = await client.ListToolsAsync();

        Assert.Equal(3, tools.Count);
        Assert.Contains(tools, t => t.Name == "query_search_analytics");
        Assert.Contains(tools, t => t.Name == "list_sites");
        Assert.Contains(tools, t => t.Name == "list_sitemaps");
    }
}
