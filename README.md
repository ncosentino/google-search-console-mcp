# Search Console MCP Server -- Google Search Analytics for AI Assistants

[![Latest Release](https://img.shields.io/github/v/release/ncosentino/google-search-console-mcp?style=flat-square)](https://github.com/ncosentino/google-search-console-mcp/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](go/go.mod)
[![.NET Version](https://img.shields.io/badge/.NET-10-512BD4?style=flat-square&logo=dotnet)](csharp/Directory.Build.props)
[![CI](https://img.shields.io/github/actions/workflow/status/ncosentino/google-search-console-mcp/ci.yml?label=CI&style=flat-square)](https://github.com/ncosentino/google-search-console-mcp/actions/workflows/ci.yml)

> **Zero-dependency MCP server for Google Search Console.**
> Pre-built native binaries for Linux, macOS, and Windows. No Node.js. No Python. No .NET runtime. No Go toolchain. Download one binary and configure your AI tool.

Expose real Google organic search data directly to AI assistants like Claude, GitHub Copilot, and Cursor via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io). Ask your AI which queries are driving traffic to which pages, identify ranking opportunities, and diagnose CTR issues -- all grounded in real Search Console data for your specific property.

---

## Why This Exists

AI assistants are powerful at analyzing SEO strategy -- but they need real data. This MCP server bridges your AI tool to Google Search Console, giving it:

- **Real organic search performance** -- clicks, impressions, CTR, and average position for every query and page in your site
- **Up to 50,000 rows per query** (vs 1,000 in the UI), unlocking complete keyword datasets
- **16 months of history** for trend analysis
- **Dimension flexibility** -- group by query, page, country, device, date, or any combination

With this MCP server configured, you can ask your AI: _"Which queries am I ranking position 8-15 for with high impressions but low CTR? What pages could I optimize to break into the top 5?"_ and get a real data-backed answer.

---

## Quick Start

**Three steps: create a service account, download a binary, add it to your MCP config.**

### Step 1: Create a Google Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/) and create or select a project
2. Enable the **Google Search Console API** for that project
3. Create a **Service Account** (IAM & Admin → Service Accounts → Create Service Account)
4. Create a JSON key for that service account and download it
5. Go to [Google Search Console](https://search.google.com/search-console) → Settings → Users and permissions
6. Add the service account email (e.g. `my-sa@my-project.iam.gserviceaccount.com`) as a **Full user** or **Restricted user** on your property

> The Search Console API is free. No billing account is required.

### Step 2: Download a Binary

Go to the [Releases page](https://github.com/ncosentino/google-search-console-mcp/releases/latest) and download the binary for your platform:

| Platform | Go binary | C# binary |
|----------|-----------|-----------|
| Linux x64 | `gsc-mcp-go-linux-amd64` | `gsc-mcp-csharp-linux-x64` |
| Linux arm64 | `gsc-mcp-go-linux-arm64` | `gsc-mcp-csharp-linux-arm64` |
| macOS x64 (Intel) | `gsc-mcp-go-darwin-amd64` | `gsc-mcp-csharp-osx-x64` |
| macOS arm64 (Apple Silicon) | `gsc-mcp-go-darwin-arm64` | `gsc-mcp-csharp-osx-arm64` |
| Windows x64 | `gsc-mcp-go-windows-amd64.exe` | `gsc-mcp-csharp-win-x64.exe` |
| Windows arm64 | `gsc-mcp-go-windows-arm64.exe` | `gsc-mcp-csharp-win-arm64.exe` |

On Linux/macOS, make the binary executable after downloading:

```bash
chmod +x gsc-mcp-go-linux-amd64
```

### Step 3: Add to Your AI Tool Config

See the [Setup by Tool](#setup-by-tool) section below for your specific client.

---

## Setup by Tool

Replace `/path/to/binary` with the actual path to your downloaded binary. Replace `/path/to/service-account.json` with the path to your downloaded service account JSON key file.

### Claude Code / GitHub Copilot CLI

```json
{
  "mcpServers": {
    "search-console": {
      "command": "/path/to/gsc-mcp-go-linux-amd64",
      "env": {
        "GOOGLE_SERVICE_ACCOUNT_FILE": "/path/to/service-account.json"
      }
    }
  }
}
```

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "search-console": {
      "command": "/path/to/gsc-mcp-go-darwin-arm64",
      "env": {
        "GOOGLE_SERVICE_ACCOUNT_FILE": "/path/to/service-account.json"
      }
    }
  }
}
```

### Cursor

```json
{
  "mcpServers": {
    "search-console": {
      "command": "/path/to/gsc-mcp-go-linux-amd64",
      "env": {
        "GOOGLE_SERVICE_ACCOUNT_FILE": "/path/to/service-account.json"
      }
    }
  }
}
```

### VS Code with GitHub Copilot

```json
{
  "mcp": {
    "servers": {
      "search-console": {
        "type": "stdio",
        "command": "/path/to/gsc-mcp-go-linux-amd64",
        "env": {
          "GOOGLE_SERVICE_ACCOUNT_FILE": "/path/to/service-account.json"
        }
      }
    }
  }
}
```

### Visual Studio

```json
{
  "search-console": {
    "command": "C:\\path\\to\\gsc-mcp-csharp-win-x64.exe",
    "env": {
      "GOOGLE_SERVICE_ACCOUNT_FILE": "C:\\path\\to\\service-account.json"
    }
  }
}
```

### Using CLI Argument

You can also pass the service account file path as a CLI argument:

```json
{
  "command": "/path/to/binary",
  "args": ["--service-account-file", "/path/to/service-account.json"]
}
```

---

## Available Tools

### `query_search_analytics`

Query clicks, impressions, CTR, and average position from Search Console.

**Parameters:**

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `site_url` | string | Yes | -- | The Search Console property URL (e.g. `https://www.example.com/` or `sc-domain:example.com`) |
| `start_date` | string | Yes | -- | Start date in `YYYY-MM-DD` format |
| `end_date` | string | Yes | -- | End date in `YYYY-MM-DD` format |
| `dimensions` | string[] | No | `[]` | Group by: `query`, `page`, `country`, `device`, `date`. Empty array returns aggregate totals. |
| `row_limit` | int | No | `1000` | Maximum rows to return (1-25000) |

**Example prompts:**

> "Which queries are driving the most impressions to my site this month but have a CTR below 2%?"

> "Show me my top 50 pages by clicks over the last 90 days."

> "Which queries am I ranking position 8-15 for? These are my best ranking improvement opportunities."

### `list_sites`

List all Search Console properties the service account has access to.

No parameters required.

**Example prompt:**

> "What Search Console properties do I have access to?"

### `list_sitemaps`

List submitted sitemaps for a property and their status.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `site_url` | string | Yes | The Search Console property URL |

**Example prompt:**

> "Show me all submitted sitemaps for my site and whether any have errors."

### Response Structure

`query_search_analytics` returns:

```json
{
  "siteUrl": "https://www.example.com/",
  "startDate": "2026-01-01",
  "endDate": "2026-01-31",
  "dimensions": ["query", "page"],
  "rowCount": 1234,
  "rows": [
    {
      "keys": ["blazor dependency injection", "https://www.example.com/blazor-di"],
      "clicks": 142,
      "impressions": 3820,
      "ctr": 0.0372,
      "position": 4.3
    }
  ],
  "queriedAt": "2026-02-21T19:00:00Z"
}
```

---

## Configuration Reference

Credential resolution uses this priority order (highest to lowest):

### 1. CLI Argument (Highest Priority)

```bash
/path/to/gsc-mcp-go-linux-amd64 --service-account-file /path/to/key.json
```

### 2. Environment Variable (File Path)

```bash
export GOOGLE_SERVICE_ACCOUNT_FILE=/path/to/service-account.json
```

### 3. Environment Variable (JSON Content)

```bash
export GOOGLE_SERVICE_ACCOUNT_JSON='{"type":"service_account","client_email":"..."}'
```

Useful in environments where you store secrets as string values rather than files (e.g. containers, CI/CD).

### 4. `.env` File (Lowest Priority -- Dev Convenience)

Create a `.env` file in the working directory:

```
GOOGLE_SERVICE_ACCOUNT_FILE=/path/to/service-account.json
```

Or with inline JSON:

```
GOOGLE_SERVICE_ACCOUNT_JSON={"type":"service_account",...}
```

---

## Go vs C# -- Which Binary?

Both implementations expose identical tools with identical behavior.

| Aspect | Go | C# Native AOT |
|--------|----|----|
| Binary size | ~10-15 MB | ~25-40 MB |
| Startup time | ~10-50ms | ~50-100ms |
| Runtime dependency | None | None |
| Language | Go 1.26 | C# / .NET 10 |
| MCP SDK | Official `go-sdk` | Official `ModelContextProtocol` |
| Auth | `golang.org/x/oauth2/google` | Native RSA + HttpClient |

**Recommendation:** Both work great. Pick Go for smaller binary size, C# if you prefer the .NET ecosystem.

---

## Building from Source

### Go

Requires Go 1.26+:

```bash
cd go
go mod tidy
go build -ldflags="-s -w" -trimpath -o gsc-mcp-go .
```

Run tests:

```bash
go test ./...
```

### C# (.NET 10 SDK Required)

```bash
cd csharp

# Build (non-AOT, for development)
dotnet build SearchConsoleMcp.slnx

# Publish Native AOT
dotnet publish src/SearchConsoleMcp/SearchConsoleMcp.csproj -r linux-x64 -c Release --self-contained true

# Run tests
dotnet test SearchConsoleMcp.slnx
```

---

## Related Projects

- [google-psi-mcp](https://github.com/ncosentino/google-psi-mcp) -- Zero-dependency MCP server for Google PageSpeed Insights Core Web Vitals
- [google-keyword-planner-mcp](https://github.com/ncosentino/google-keyword-planner-mcp) -- Zero-dependency MCP server for Google Ads Keyword Planner (keyword ideas, search volume, CPC)

---

## About

### Nick Cosentino -- Dev Leader

This MCP server was built by **[Nick Cosentino](https://www.devleader.ca)**, a software engineer and content creator known as **Dev Leader**. Nick creates practical .NET, C#, ASP.NET Core, Blazor, and software engineering content for intermediate to advanced developers.

**Find Nick online:**

- Blog: [https://www.devleader.ca](https://www.devleader.ca)
- YouTube: [https://www.youtube.com/@devleaderca](https://www.youtube.com/@devleaderca)
- Newsletter: [https://weekly.devleader.ca](https://weekly.devleader.ca)
- LinkedIn: [https://linkedin.com/in/nickcosentino](https://linkedin.com/in/nickcosentino)
- Linktree: [https://www.linktr.ee/devleader](https://www.linktr.ee/devleader)

### BrandGhost

[BrandGhost](https://www.brandghost.ai) is a social media automation platform built by Nick that lets content creators cross-post and schedule content across all social platforms in one click.

---

## Contributing

Contributions are welcome! Please:

1. Open an issue describing the bug or feature request before submitting a PR
2. Run `golangci-lint run` (Go) or `dotnet build` with zero warnings (C#) before submitting
3. Keep both implementations in sync -- a feature added to Go should also be added to C#, and vice versa

---

## License

MIT License -- see [LICENSE](LICENSE) for details.
