using System.Runtime.CompilerServices;

namespace SearchConsoleMcp.Config;

/// <summary>Resolves Google service account credentials from multiple sources.</summary>
/// <remarks>
/// Priority order: CLI argument (file path) > GOOGLE_SERVICE_ACCOUNT_FILE env var >
/// GOOGLE_SERVICE_ACCOUNT_JSON env var > .env file.
/// </remarks>
internal static class ServiceAccountResolver
{
    private const string EnvVarFile = "GOOGLE_SERVICE_ACCOUNT_FILE";
    private const string EnvVarJson = "GOOGLE_SERVICE_ACCOUNT_JSON";
    private const string DotEnvFile = ".env";

    /// <summary>Returns the service account JSON bytes from the highest-priority available source.</summary>
    internal static byte[]? Resolve(string? flagFilePath)
    {
        if (!string.IsNullOrWhiteSpace(flagFilePath))
        {
            if (File.Exists(flagFilePath))
                return File.ReadAllBytes(flagFilePath);
        }

        var filePathFromEnv = Environment.GetEnvironmentVariable(EnvVarFile);
        if (!string.IsNullOrWhiteSpace(filePathFromEnv) && File.Exists(filePathFromEnv))
            return File.ReadAllBytes(filePathFromEnv);

        var jsonFromEnv = Environment.GetEnvironmentVariable(EnvVarJson);
        if (!string.IsNullOrWhiteSpace(jsonFromEnv))
            return System.Text.Encoding.UTF8.GetBytes(jsonFromEnv);

        return ReadFromDotEnv();
    }

    [MethodImpl(MethodImplOptions.NoInlining)]
    private static byte[]? ReadFromDotEnv()
    {
        if (!File.Exists(DotEnvFile))
            return null;

        foreach (var line in File.ReadLines(DotEnvFile))
        {
            var trimmed = line.Trim();
            if (trimmed.StartsWith('#') || trimmed.Length == 0)
                continue;

            var filePrefix = EnvVarFile + "=";
            if (trimmed.StartsWith(filePrefix, StringComparison.Ordinal))
            {
                var path = trimmed[filePrefix.Length..].Trim('"', '\'');
                if (File.Exists(path))
                    return File.ReadAllBytes(path);
            }

            var jsonPrefix = EnvVarJson + "=";
            if (trimmed.StartsWith(jsonPrefix, StringComparison.Ordinal))
            {
                var value = trimmed[jsonPrefix.Length..].Trim('"', '\'');
                return string.IsNullOrWhiteSpace(value) ? null : System.Text.Encoding.UTF8.GetBytes(value);
            }
        }

        return null;
    }
}
