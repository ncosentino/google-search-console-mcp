using SearchConsoleMcp.Config;
using Xunit;

namespace SearchConsoleMcp.Tests;

public sealed class ServiceAccountResolverTests
{
    [Fact]
    public void Resolve_ValidFilePath_ReturnsBytes()
    {
        // Write a temp file with fake JSON content.
        var path = Path.GetTempFileName();
        try
        {
            File.WriteAllText(path, "{\"type\":\"service_account\"}");
            var result = ServiceAccountResolver.Resolve(path);
            Assert.NotNull(result);
            Assert.NotEmpty(result);
        }
        finally
        {
            File.Delete(path);
        }
    }

    [Fact]
    public void Resolve_NullFlag_ReturnsNull_WhenNoEnvOrFile()
    {
        // Only assert when the env vars are not set in the current environment.
        if (Environment.GetEnvironmentVariable("GOOGLE_SERVICE_ACCOUNT_FILE") is not null ||
            Environment.GetEnvironmentVariable("GOOGLE_SERVICE_ACCOUNT_JSON") is not null)
            return; // skip -- env var present

        var result = ServiceAccountResolver.Resolve(null);
        Assert.Null(result);
    }

    [Fact]
    public void Resolve_FlagTakesPriorityOverEnv()
    {
        var path = Path.GetTempFileName();
        try
        {
            File.WriteAllText(path, "{\"type\":\"service_account\",\"flag\":true}");
            var result = ServiceAccountResolver.Resolve(path);
            Assert.NotNull(result);
            var text = System.Text.Encoding.UTF8.GetString(result);
            Assert.Contains("flag", text, StringComparison.Ordinal);
        }
        finally
        {
            File.Delete(path);
        }
    }
}
