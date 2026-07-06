using System.Globalization;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using ModelContextProtocol.Server;
using SearchConsoleMcp;
using SearchConsoleMcp.Config;
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

// --transport has no environment-variable fallback, matching the Go implementation.
string transport = args.SkipWhile(a => a != "--transport").Skip(1).FirstOrDefault() ?? "stdio";

if (transport == "http")
{
    var port = Environment.GetEnvironmentVariable("PORT") is { Length: > 0 } portEnv ? portEnv : "8080";
    var app = Hosting.BuildHttpHost(args, serviceAccountJson, int.Parse(port, CultureInfo.InvariantCulture));
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
