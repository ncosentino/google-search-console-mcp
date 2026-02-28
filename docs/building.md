---
description: Build the Google Search Console MCP server from source -- Go and C# build commands, test commands, and Native AOT publish instructions for all platforms.
---

# Building from Source

Both implementations can be built locally from the repository. Pre-built binaries are available on the [Releases page](https://github.com/ncosentino/google-search-console-mcp/releases/latest) if you don't need to build from source.

---

## Go

**Requirements:** Go 1.26+

```bash
cd go
go mod tidy
go build -ldflags="-s -w" -trimpath -o gsc-mcp-go .
```

Run tests:

```bash
go test ./...
```

Run linter (requires [golangci-lint](https://golangci-lint.run/)):

```bash
golangci-lint run
```

---

## C# (.NET 10)

**Requirements:** .NET 10 SDK

Build for development (no AOT):

```bash
cd csharp
dotnet restore SearchConsoleMcp.slnx
dotnet build SearchConsoleMcp.slnx -c Release --no-restore
```

Run tests:

```bash
dotnet test SearchConsoleMcp.slnx -c Release --no-build
```

Publish as a Native AOT self-contained binary:

=== "Linux x64"
    ```bash
    dotnet publish src/SearchConsoleMcp/SearchConsoleMcp.csproj \
      -r linux-x64 -c Release --self-contained true
    ```

=== "macOS arm64"
    ```bash
    dotnet publish src/SearchConsoleMcp/SearchConsoleMcp.csproj \
      -r osx-arm64 -c Release --self-contained true
    ```

=== "Windows x64"
    ```bash
    dotnet publish src/SearchConsoleMcp/SearchConsoleMcp.csproj ^
      -r win-x64 -c Release --self-contained true
    ```

!!! note "Native AOT requirements"
    Native AOT compilation on Linux requires `clang` and `zlib1g-dev`. Install with:
    ```bash
    sudo apt-get install -y clang zlib1g-dev
    ```

---

## Contributing

1. Open an issue describing the bug or feature before submitting a PR
2. Run `golangci-lint run` (Go) or `dotnet build` with zero warnings (C#) before submitting
3. Keep both implementations in sync -- a feature added to Go should also be added to C#
