using System.Net;
using System.Reflection;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Primitives;
using ModelContextProtocol.Server;
using SearchConsoleMcp.SearchConsole;
using SearchConsoleMcp.Tools;

namespace SearchConsoleMcp;

/// <summary>Builds the MCP server's hosts for both the stdio and HTTP transports.</summary>
internal static class Hosting
{
    /// <summary>Default Host header allow-list when none is configured via AllowedHosts.</summary>
    internal const string DefaultAllowedHosts = "localhost;127.0.0.1;[::1]";

    /// <summary>Health-check endpoint for service supervisors.</summary>
    internal const string HealthPath = "/health";

    /// <summary>Streamable HTTP MCP endpoint.</summary>
    internal const string McpPath = "/mcp";

    /// <summary>Authenticated local service shutdown endpoint.</summary>
    internal const string ShutdownPath = "/shutdown";

    private const long MaxMcpRequestBytes = 1 << 20;

    /// <summary>Builds an HTTP host without starting it.</summary>
    internal static WebApplication BuildHttpHost(
        string[] args,
        byte[] serviceAccountJson,
        int port,
        HttpMessageHandler? httpMessageHandler = null,
        string listenAddress = ServerOptions.DefaultListenAddress,
        string? shutdownToken = null)
    {
        var builder = WebApplication.CreateBuilder(args);
        if (string.IsNullOrWhiteSpace(builder.Configuration["AllowedHosts"]))
        {
            builder.Configuration["AllowedHosts"] = DefaultAllowedHosts;
        }

        builder.WebHost.UseUrls($"http://{FormatListenAddress(listenAddress)}:{port}");
        builder.WebHost.ConfigureKestrel(options =>
        {
            options.Limits.MaxRequestBodySize = MaxMcpRequestBytes;
            options.Limits.RequestHeadersTimeout = TimeSpan.FromSeconds(5);
            options.Limits.KeepAliveTimeout = TimeSpan.FromMinutes(2);
        });

        ConfigureCommonServices(builder, serviceAccountJson, httpMessageHandler);
        builder.Services.AddMcpServer()
            .WithStringifiedArgsCoercion()
            .WithHttpTransport(options => options.Stateless = true)
            .WithTools<SearchConsoleTool>();

        var app = builder.Build();
        app.Use(async (context, next) =>
        {
            if (context.Request.Path.StartsWithSegments(McpPath) &&
                !IsAllowedOrigin(context.Request, context.Request.Headers.Origin))
            {
                context.Response.StatusCode = StatusCodes.Status403Forbidden;
                return;
            }
            await next(context).ConfigureAwait(false);
        });
        app.MapGet(HealthPath, () =>
        {
            var response = new ServiceHealth(
                "ok",
                "google-search-console-mcp",
                GetServiceVersion(),
                "http");
            return Results.Text(
                JsonSerializer.Serialize(response, HostingJsonContext.Default.ServiceHealth),
                "application/json");
        });
        app.MapMcp(McpPath);
        if (!string.IsNullOrEmpty(shutdownToken))
        {
            app.MapPost(ShutdownPath, async context =>
            {
                if (context.Connection.RemoteIpAddress is null ||
                    !IPAddress.IsLoopback(context.Connection.RemoteIpAddress))
                {
                    context.Response.StatusCode = StatusCodes.Status403Forbidden;
                    return;
                }
                if (!HasBearerToken(context.Request, shutdownToken))
                {
                    context.Response.StatusCode = StatusCodes.Status401Unauthorized;
                    return;
                }

                context.Response.StatusCode = StatusCodes.Status202Accepted;
                context.Response.ContentType = "application/json";
                context.Response.Headers.CacheControl = "no-store";
                await context.Response.WriteAsync("""{"stopping":true}""")
                    .ConfigureAwait(false);
                app.Lifetime.StopApplication();
            });
        }
        return app;
    }

    /// <summary>Registers services shared by both transports.</summary>
    internal static void ConfigureCommonServices(
        IHostApplicationBuilder builder,
        byte[] serviceAccountJson,
        HttpMessageHandler? httpMessageHandler = null)
    {
        builder.Logging.AddConsole(options =>
            options.LogToStandardErrorThreshold = LogLevel.Trace);
        builder.Logging.SetMinimumLevel(LogLevel.Warning);

        var apiClient = builder.Services.AddHttpClient(nameof(SearchConsoleClient), http =>
        {
            http.Timeout = TimeSpan.FromSeconds(30);
        });
        if (httpMessageHandler is not null)
        {
            apiClient.ConfigurePrimaryHttpMessageHandler(() => httpMessageHandler);
        }

        builder.Services.AddTransient<SearchConsoleClient>(services =>
        {
            var factory = services.GetRequiredService<IHttpClientFactory>();
            return SearchConsoleClient.Create(
                serviceAccountJson,
                factory.CreateClient(nameof(SearchConsoleClient)));
        });
    }

    internal static bool IsAllowedOrigin(HttpRequest request, StringValues origins)
    {
        if (StringValues.IsNullOrEmpty(origins))
        {
            return true;
        }
        if (origins.Count != 1 ||
            !Uri.TryCreate(origins[0], UriKind.Absolute, out var origin) ||
            origin.UserInfo.Length != 0)
        {
            return false;
        }

        var requestPort = request.Host.Port ?? DefaultPort(request.Scheme);
        var originPort = origin.IsDefaultPort ? DefaultPort(origin.Scheme) : origin.Port;
        return string.Equals(request.Scheme, origin.Scheme, StringComparison.OrdinalIgnoreCase) &&
            string.Equals(request.Host.Host, origin.Host, StringComparison.OrdinalIgnoreCase) &&
            requestPort == originPort;
    }

    private static bool HasBearerToken(HttpRequest request, string expectedToken)
    {
        const string prefix = "Bearer ";
        var authorization = request.Headers.Authorization.ToString();
        if (!authorization.StartsWith(prefix, StringComparison.Ordinal))
        {
            return false;
        }

        return CryptographicOperations.FixedTimeEquals(
            Encoding.UTF8.GetBytes(authorization[prefix.Length..]),
            Encoding.UTF8.GetBytes(expectedToken));
    }

    private static string FormatListenAddress(string listenAddress) =>
        IPAddress.TryParse(listenAddress, out var address) &&
        address.AddressFamily == System.Net.Sockets.AddressFamily.InterNetworkV6
            ? $"[{listenAddress}]"
            : listenAddress;

    private static int DefaultPort(string scheme) =>
        string.Equals(scheme, Uri.UriSchemeHttps, StringComparison.OrdinalIgnoreCase) ? 443 : 80;

    private static string GetServiceVersion()
    {
        var assembly = typeof(Hosting).Assembly;
        return assembly.GetCustomAttribute<AssemblyInformationalVersionAttribute>()?
            .InformationalVersion
            ?? assembly.GetName().Version?.ToString()
            ?? "dev";
    }
}

internal sealed record ServiceHealth(
    string Status,
    string Service,
    string Version,
    string Transport);

/// <summary>System.Text.Json source generation context for service metadata.</summary>
[JsonSerializable(typeof(ServiceHealth))]
[JsonSourceGenerationOptions(PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase)]
internal partial class HostingJsonContext : JsonSerializerContext;
