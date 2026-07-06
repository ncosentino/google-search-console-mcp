using System.Text.Json;
using ModelContextProtocol.Protocol;
using Xunit;

namespace SearchConsoleMcp.Tests;

/// <summary>
/// Tests for StringifiedArgsCoercion, written as part of #7 (dimensions arriving
/// JSON-encoded as a string instead of a genuine array). The end-to-end scenario
/// (a real MCP client session, before vs. after this filter was wired into
/// Program.cs) was verified manually via a real compiled binary over real stdio:
/// before the fix, a stringified "dimensions" value caused a generic SDK-level
/// parameter-binding failure (no real Google API call attempted); after the fix,
/// the same call proceeds past binding to the actual tool method body. That
/// end-to-end proof isn't automated here because reaching the method body
/// requires a real (if fake-credentialed and fast-failing) network call to
/// Google's OAuth2 endpoint, which would make this suite depend on live network
/// access. These tests instead cover the coercion logic itself, deterministically
/// and without any network dependency.
/// </summary>
public sealed class StringifiedArgsCoercionTests
{
    private static CallToolRequestParams RequestFor(string toolName, Dictionary<string, JsonElement> arguments)
        => new() { Name = toolName, Arguments = arguments };

    private static JsonElement JsonOf(string json) => JsonDocument.Parse(json).RootElement.Clone();

    [Fact]
    public void CoerceStringifiedArrayArgs_StringifiedArray_IsReplacedWithGenuineArray()
    {
        var requestParams = RequestFor("query_search_analytics", new Dictionary<string, JsonElement>
        {
            ["dimensions"] = JsonOf("\"[\\\"query\\\",\\\"page\\\"]\""),
        });

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(requestParams, StringifiedArgsCoercion.ToolArrayFields);

        var dimensions = requestParams.Arguments!["dimensions"];
        Assert.Equal(JsonValueKind.Array, dimensions.ValueKind);
        Assert.Equal(2, dimensions.GetArrayLength());
        Assert.Equal("query", dimensions[0].GetString());
        Assert.Equal("page", dimensions[1].GetString());
    }

    [Fact]
    public void CoerceStringifiedArrayArgs_GenuineArray_IsLeftUnchanged()
    {
        var requestParams = RequestFor("query_search_analytics", new Dictionary<string, JsonElement>
        {
            ["dimensions"] = JsonOf("[\"query\"]"),
        });

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(requestParams, StringifiedArgsCoercion.ToolArrayFields);

        var dimensions = requestParams.Arguments!["dimensions"];
        Assert.Equal(JsonValueKind.Array, dimensions.ValueKind);
        Assert.Equal("query", dimensions[0].GetString());
    }

    [Fact]
    public void CoerceStringifiedArrayArgs_StringThatIsNotJson_IsLeftUnchanged()
    {
        var requestParams = RequestFor("query_search_analytics", new Dictionary<string, JsonElement>
        {
            ["dimensions"] = JsonOf("\"not valid json at all\""),
        });

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(requestParams, StringifiedArgsCoercion.ToolArrayFields);

        Assert.Equal(JsonValueKind.String, requestParams.Arguments!["dimensions"].ValueKind);
    }

    [Fact]
    public void CoerceStringifiedArrayArgs_StringThatDecodesToNonArray_IsLeftUnchanged()
    {
        var requestParams = RequestFor("query_search_analytics", new Dictionary<string, JsonElement>
        {
            // A string that happens to be valid JSON, but decodes to an object, not an array.
            ["dimensions"] = JsonOf("\"{\\\"not\\\":\\\"an array\\\"}\""),
        });

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(requestParams, StringifiedArgsCoercion.ToolArrayFields);

        Assert.Equal(JsonValueKind.String, requestParams.Arguments!["dimensions"].ValueKind);
    }

    [Fact]
    public void CoerceStringifiedArrayArgs_MissingField_IsANoOp()
    {
        var requestParams = RequestFor("query_search_analytics", new Dictionary<string, JsonElement>
        {
            ["site_url"] = JsonOf("\"devleader.ca\""),
        });

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(requestParams, StringifiedArgsCoercion.ToolArrayFields);

        Assert.False(requestParams.Arguments!.ContainsKey("dimensions"));
    }

    [Fact]
    public void CoerceStringifiedArrayArgs_UnrelatedTool_IsANoOp()
    {
        var requestParams = RequestFor("list_sites", new Dictionary<string, JsonElement>
        {
            ["dimensions"] = JsonOf("\"[\\\"query\\\"]\""),
        });

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(requestParams, StringifiedArgsCoercion.ToolArrayFields);

        // list_sites isn't in ToolArrayFields, so even a field named "dimensions" must be left alone.
        Assert.Equal(JsonValueKind.String, requestParams.Arguments!["dimensions"].ValueKind);
    }

    [Fact]
    public void CoerceStringifiedArrayArgs_NullArguments_IsANoOp()
    {
        var requestParams = new CallToolRequestParams { Name = "query_search_analytics", Arguments = null };

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(requestParams, StringifiedArgsCoercion.ToolArrayFields);

        Assert.Null(requestParams.Arguments);
    }
}
