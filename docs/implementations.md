# Go vs C#

Both the Go and C# implementations expose identical MCP tools with identical behavior and response formats. Pick based on your preferences.

---

## Comparison

| Aspect | Go | C# Native AOT |
|--------|----|----|
| Binary size | ~10--15 MB | ~25--40 MB |
| Startup time | ~10--50ms | ~50--100ms |
| Runtime dependency | None | None |
| Language | Go 1.26 | C# / .NET 10 |
| MCP SDK | Official `modelcontextprotocol/go-sdk` | Official `ModelContextProtocol` (.NET) |
| Auth implementation | `golang.org/x/oauth2/google` | Native RSA + HttpClient |
| AOT compiled | Yes | Yes (Native AOT) |

Both binaries are fully self-contained. No Go toolchain, .NET runtime, Node.js, or Python is needed to run them.

---

## Recommendation

**Both work great.** Pick Go for a smaller binary size. Pick C# if you're more comfortable with the .NET ecosystem or want to inspect/modify the source.

---

## Naming Convention

Binary names follow this pattern:

```
gsc-mcp-{go|csharp}-{os}-{arch}[.exe]
```

Examples:

- `gsc-mcp-go-linux-amd64` -- Go, Linux x64
- `gsc-mcp-csharp-win-x64.exe` -- C#, Windows x64
- `gsc-mcp-go-darwin-arm64` -- Go, macOS Apple Silicon

See the [Getting Started](getting-started.md#step-2-download-a-binary) page for the full download table.

---

## Tool Parity

If a feature is added to one implementation, it's added to both. The Go and C# versions are kept in sync. If you discover a behavioral difference between the two, please [open an issue](https://github.com/ncosentino/google-search-console-mcp/issues).
