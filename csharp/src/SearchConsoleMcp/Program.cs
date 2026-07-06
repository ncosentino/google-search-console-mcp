using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using ModelContextProtocol.Server;
using SearchConsoleMcp;
using SearchConsoleMcp.Config;
using SearchConsoleMcp.SearchConsole;
using SearchConsoleMcp.Tools;

var serviceAccountFilePath = args.SkipWhile(a => a != "--service-account-file").Skip(1).FirstOrDefault();
var serviceAccountJson = ServiceAccountResolver.Resolve(serviceAccountFilePath);

if (serviceAccountJson is null || serviceAccountJson.Length == 0)
{
    await Console.Error.WriteLineAsync(
        "ERROR: No service account credentials provided. " +
        "Use --service-account-file <path>, set GOOGLE_SERVICE_ACCOUNT_FILE env var, " +
        "or set GOOGLE_SERVICE_ACCOUNT_JSON env var.")
        .ConfigureAwait(false);
    return 1;
}

var builder = Host.CreateApplicationBuilder(args);

// All logs must go to stderr to avoid corrupting the MCP STDIO stream.
builder.Logging.AddConsole(o => o.LogToStandardErrorThreshold = LogLevel.Trace);
builder.Logging.SetMinimumLevel(LogLevel.Warning);

builder.Services
    .AddHttpClient(nameof(SearchConsoleClient), http =>
    {
        http.Timeout = TimeSpan.FromSeconds(30);
    });

builder.Services.AddTransient<SearchConsoleClient>(sp =>
{
    var factory = sp.GetRequiredService<IHttpClientFactory>();
    return SearchConsoleClient.Create(serviceAccountJson, factory.CreateClient(nameof(SearchConsoleClient)));
});

builder.Services
    .AddMcpServer()
    .WithRequestFilters(requestFilters =>
    {
        // Repair a widespread MCP client bug where array-typed arguments arrive
        // JSON-encoded as a string instead of a genuine array (see
        // StringifiedArgsCoercion.cs).
        requestFilters.AddCallToolFilter(next => async (context, cancellationToken) =>
        {
            if (context.Params is not null)
                StringifiedArgsCoercion.CoerceStringifiedArrayArgs(context.Params, StringifiedArgsCoercion.ToolArrayFields);
            return await next(context, cancellationToken).ConfigureAwait(false);
        });
    })
    .WithStdioServerTransport()
    .WithTools<SearchConsoleTool>();

var host = builder.Build();
await host.RunAsync().ConfigureAwait(false);
return 0;
