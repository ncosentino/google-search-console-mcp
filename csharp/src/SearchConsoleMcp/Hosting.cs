using Microsoft.AspNetCore.Builder;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using ModelContextProtocol.Server;
using SearchConsoleMcp.SearchConsole;
using SearchConsoleMcp.Tools;

namespace SearchConsoleMcp;

/// <summary>Builds the MCP server's hosts for both the stdio and HTTP transports.</summary>
internal static class Hosting
{
    /// <summary>Default Host header allow-list when none is configured via AllowedHosts.</summary>
    internal const string DefaultAllowedHosts = "localhost;127.0.0.1;[::1]";

    /// <summary>
    /// Builds a WebApplication configured for the MCP Streamable HTTP transport,
    /// listening on the given port. Does not call Run/RunAsync -- the caller owns the
    /// application's lifetime, so tests can bind an ephemeral port and stop the host
    /// directly instead of going through Program's own command-line parsing.
    /// </summary>
    internal static WebApplication BuildHttpHost(
        string[] args,
        byte[] serviceAccountJson,
        int port,
        HttpMessageHandler? httpMessageHandler = null)
    {
        var builder = WebApplication.CreateBuilder(args);

        // Host Filtering Middleware (added automatically by WebApplication.CreateBuilder)
        // is disabled until AllowedHosts is set. Default to loopback-only unless the
        // caller already configured it via appsettings.json, an environment variable,
        // or a --AllowedHosts command-line argument -- all standard .NET configuration
        // sources, already merged into builder.Configuration by this point.
        if (string.IsNullOrWhiteSpace(builder.Configuration["AllowedHosts"]))
        {
            builder.Configuration["AllowedHosts"] = DefaultAllowedHosts;
        }

        builder.WebHost.UseUrls($"http://0.0.0.0:{port}");

        ConfigureCommonServices(builder, serviceAccountJson, httpMessageHandler);

        builder.Services.AddMcpServer()
            .WithStringifiedArgsCoercion()
            .WithHttpTransport(options =>
            {
                // This server has no need for server-to-client requests, so stateless
                // mode is the documented recommendation: no session-affinity
                // requirements, and no in-memory session state to leak across requests
                // or restarts.
                options.Stateless = true;
            })
            .WithTools<SearchConsoleTool>();

        var app = builder.Build();
        app.MapMcp();
        return app;
    }

    /// <summary>Registers the DI services shared between the stdio and HTTP hosts.</summary>
    internal static void ConfigureCommonServices(
        IHostApplicationBuilder builder,
        byte[] serviceAccountJson,
        HttpMessageHandler? httpMessageHandler = null)
    {
        builder.Logging.AddConsole(o => o.LogToStandardErrorThreshold = LogLevel.Trace);
        builder.Logging.SetMinimumLevel(LogLevel.Warning);

        var apiClient = builder.Services.AddHttpClient(nameof(SearchConsoleClient), http =>
        {
            http.Timeout = TimeSpan.FromSeconds(30);
        });

        // Production (Program.cs) never passes a handler, so this is a no-op there;
        // tests use it to substitute a fake Google Search Console / OAuth2 endpoint
        // instead of making real network calls.
        if (httpMessageHandler is not null)
        {
            apiClient.ConfigurePrimaryHttpMessageHandler(() => httpMessageHandler);
        }

        builder.Services.AddTransient<SearchConsoleClient>(sp =>
        {
            var factory = sp.GetRequiredService<IHttpClientFactory>();
            return SearchConsoleClient.Create(serviceAccountJson, factory.CreateClient(nameof(SearchConsoleClient)));
        });
    }
}
