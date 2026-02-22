namespace SearchConsoleMcp.SearchConsole;

/// <summary>Pure functions for normalizing and resolving Google Search Console site URLs.</summary>
internal static class SiteUrlResolver
{
    /// <summary>
    /// Normalizes a user-supplied site URL to the canonical GSC property format.
    /// </summary>
    /// <remarks>
    /// Rules:
    /// <list type="bullet">
    ///   <item>Already <c>sc-domain:*</c> → returned unchanged.</item>
    ///   <item>Bare domain (no scheme, no <c>sc-domain:</c>) → <c>sc-domain:&lt;apex&gt;</c>.</item>
    ///   <item>URL with scheme and no trailing slash path → <c>sc-domain:&lt;apex&gt;</c>.</item>
    ///   <item>URL with trailing slash at root, or a non-trivial path → returned as URL-prefix with trailing slash.</item>
    /// </list>
    /// </remarks>
    internal static string Normalize(string input)
    {
        var trimmed = input.Trim();

        if (trimmed.StartsWith("sc-domain:", StringComparison.OrdinalIgnoreCase))
            return trimmed;

        if (!Uri.TryCreate(trimmed, UriKind.Absolute, out var uri))
        {
            // Bare domain: strip www. and return sc-domain form.
            return "sc-domain:" + ExtractApexDomain(trimmed);
        }

        // Has a scheme. Check path.
        var path = uri.AbsolutePath;

        // Explicit trailing slash at root = caller meant a URL-prefix property.
        if (path == "/" && trimmed.EndsWith("/", StringComparison.Ordinal))
            return trimmed;

        // Non-trivial path: ensure trailing slash, keep as URL-prefix.
        if (path.Length > 1)
        {
            var withSlash = trimmed.TrimEnd('/') + "/";
            return withSlash;
        }

        // Scheme but no trailing slash or path: treat as domain property.
        return "sc-domain:" + ExtractApexDomain(uri.Host);
    }

    /// <summary>
    /// Finds the best matching GSC property from a list of accessible sites for the given input.
    /// Prefers <c>sc-domain:&lt;apex&gt;</c> over URL-prefix matches.
    /// Returns <c>null</c> if no match is found.
    /// </summary>
    internal static string? FindBestMatch(string input, IEnumerable<SiteEntry> sites)
    {
        var apex = ExtractApexDomain(input);
        string? urlPrefixMatch = null;

        foreach (var site in sites)
        {
            if (site.SiteUrl.StartsWith("sc-domain:", StringComparison.OrdinalIgnoreCase))
            {
                var siteApex = site.SiteUrl["sc-domain:".Length..];
                if (string.Equals(siteApex, apex, StringComparison.OrdinalIgnoreCase))
                    return site.SiteUrl; // domain property is always preferred
            }
            else if (urlPrefixMatch is null)
            {
                // URL-prefix: match if the apex domain appears in the host.
                if (Uri.TryCreate(site.SiteUrl, UriKind.Absolute, out var siteUri))
                {
                    var siteApex = ExtractApexDomain(siteUri.Host);
                    if (string.Equals(siteApex, apex, StringComparison.OrdinalIgnoreCase))
                        urlPrefixMatch = site.SiteUrl;
                }
            }
        }

        return urlPrefixMatch;
    }

    private static string ExtractApexDomain(string input)
    {
        var trimmed = input.Trim();

        // If it looks like a URL, parse the host.
        if (Uri.TryCreate(trimmed, UriKind.Absolute, out var uri))
            trimmed = uri.Host;

        // Strip port if present.
        var colonIdx = trimmed.IndexOf(':');
        if (colonIdx >= 0)
            trimmed = trimmed[..colonIdx];

        // Strip www. prefix.
        if (trimmed.StartsWith("www.", StringComparison.OrdinalIgnoreCase))
            trimmed = trimmed[4..];

        return trimmed.ToLowerInvariant();
    }
}
