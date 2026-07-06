using ModelContextProtocol.Server;

namespace SearchConsoleMcp;

/// <summary>Shared IMcpServerBuilder configuration used by both the stdio and HTTP hosts.</summary>
internal static class McpServerBuilderExtensions
{
    /// <summary>
    /// Registers the call-tool filter that repairs a widespread MCP client bug where
    /// array-typed arguments arrive JSON-encoded as a string instead of a genuine
    /// array (see StringifiedArgsCoercion.cs). Extracted so both the stdio host
    /// (Program.cs) and the HTTP host (Hosting.BuildHttpHost) apply the identical
    /// filter instead of duplicating the registration in both places.
    /// </summary>
    internal static IMcpServerBuilder WithStringifiedArgsCoercion(this IMcpServerBuilder builder) =>
        builder.WithRequestFilters(requestFilters =>
        {
            requestFilters.AddCallToolFilter(next => async (context, cancellationToken) =>
            {
                if (context.Params is not null)
                {
                    StringifiedArgsCoercion.CoerceStringifiedArrayArgs(
                        context.Params, StringifiedArgsCoercion.ToolArrayFields);
                }
                return await next(context, cancellationToken).ConfigureAwait(false);
            });
        });
}
