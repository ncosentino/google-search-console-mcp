using System.Text.Json;
using ModelContextProtocol.Protocol;

namespace SearchConsoleMcp;

/// <summary>
/// Repairs a widespread MCP client bug where array-typed tool arguments arrive
/// JSON-encoded as a string (e.g. <c>"[\"query\"]"</c> instead of <c>["query"]</c>)
/// instead of a genuine JSON array. That failure mode has been reproduced against
/// multiple, unrelated MCP servers and confirmed independent of this server's schema
/// shape (see google-keyword-planner-mcp#2/#4, and this repo's #7), so it is repaired
/// defensively here rather than left to reject valid calls. Mirrors the Go server's
/// coerceStringifiedArrayArgs (go/stringified_args.go) for parity between the two
/// implementations.
/// </summary>
internal static class StringifiedArgsCoercion
{
    /// <summary>
    /// Declares, for each tool name, which top-level argument fields are array-typed.
    /// <see cref="CoerceStringifiedArrayArgs"/> uses this to know which fields to
    /// repair; it is intentionally a plain data map (not per-tool duplicated logic),
    /// so every tool with an array-typed parameter is covered by one code path.
    /// </summary>
    internal static readonly IReadOnlyDictionary<string, string[]> ToolArrayFields =
        new Dictionary<string, string[]>
        {
            ["query_search_analytics"] = ["dimensions"],
        };

    /// <summary>
    /// Mutates <paramref name="requestParams"/>'s <see cref="CallToolRequestParams.Arguments"/>
    /// in place, replacing any declared array field whose value is a JSON string that
    /// itself decodes to a JSON array with the decoded array. Fields that are missing,
    /// already an array, or a string that doesn't decode to an array are left
    /// untouched, so normal schema/parameter-binding validation can handle them.
    /// </summary>
    /// <remarks>
    /// This must run as a call-tool filter (see <c>Program.cs</c>'s
    /// <c>WithRequestFilters</c> registration), not inside a tool method: the SDK
    /// binds <see cref="CallToolRequestParams.Arguments"/> to the tool method's typed
    /// parameters before the method body ever runs, so a malformed argument never
    /// reaches the method to be fixed there. The filter runs earlier, while arguments
    /// are still raw <see cref="JsonElement"/> values.
    /// </remarks>
    internal static void CoerceStringifiedArrayArgs(
        CallToolRequestParams requestParams,
        IReadOnlyDictionary<string, string[]> arrayFieldsByTool)
    {
        if (requestParams.Arguments is null)
            return;
        if (!arrayFieldsByTool.TryGetValue(requestParams.Name, out var fields))
            return;

        foreach (var field in fields)
        {
            if (!requestParams.Arguments.TryGetValue(field, out var value))
                continue;

            if (TryCoerceStringifiedArray(value, out var coerced))
                requestParams.Arguments[field] = coerced;
        }
    }

    /// <summary>
    /// Reports whether <paramref name="value"/> is a JSON string that itself decodes
    /// to a JSON array, returning that array as a <see cref="JsonElement"/> if so.
    /// Returns <see langword="false"/> for a value that is missing, already an array,
    /// or a string that doesn't decode to an array -- in all of those cases the
    /// caller should leave the value untouched and let normal validation handle it.
    /// </summary>
    private static bool TryCoerceStringifiedArray(JsonElement value, out JsonElement coerced)
    {
        coerced = default;

        if (value.ValueKind != JsonValueKind.String)
            return false; // Already an array (or some other type); leave as-is.

        var asString = value.GetString();
        if (string.IsNullOrEmpty(asString))
            return false;

        JsonDocument probe;
        try
        {
            probe = JsonDocument.Parse(asString);
        }
        catch (JsonException)
        {
            return false; // Not valid JSON at all; leave for validation to reject.
        }

        using (probe)
        {
            if (probe.RootElement.ValueKind != JsonValueKind.Array)
                return false; // Decodes, but not to an array; leave for validation to reject.

            coerced = probe.RootElement.Clone();
            return true;
        }
    }
}
