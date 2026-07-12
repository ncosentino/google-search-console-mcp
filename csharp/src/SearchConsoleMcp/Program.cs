using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using ModelContextProtocol.Server;
using SearchConsoleMcp;
using SearchConsoleMcp.Config;
using SearchConsoleMcp.Tools;

if (ServerOptions.IsHelpRequested(args))
{
    await Console.Out.WriteLineAsync(ServerOptions.Usage).ConfigureAwait(false);
    return 0;
}

ServerOptions options;
try
{
    options = ServerOptions.Parse(args);
}
catch (ArgumentException exception)
{
    await Console.Error.WriteLineAsync($"ERROR: {exception.Message}").ConfigureAwait(false);
    return 1;
}

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

if (options.Transport == "http")
{
    var app = Hosting.BuildHttpHost(
        args,
        serviceAccountJson,
        options.Port,
        listenAddress: options.ListenAddress,
        shutdownToken: options.ShutdownToken);
    await app.RunAsync().ConfigureAwait(false);
    return 0;
}

var builder = Host.CreateApplicationBuilder(args);

Hosting.ConfigureCommonServices(builder, serviceAccountJson);

builder.Services
    .AddMcpServer()
    .WithStringifiedArgsCoercion()
    .WithStdioServerTransport()
    .WithTools<SearchConsoleTool>();

var host = builder.Build();
await host.RunAsync().ConfigureAwait(false);
return 0;
