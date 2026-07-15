# Search Console MCP Server -- Search Analytics and URL Inspection for AI Assistants

[![Latest Release](https://img.shields.io/github/v/release/ncosentino/google-search-console-mcp?style=flat-square)](https://github.com/ncosentino/google-search-console-mcp/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](go/go.mod)
[![.NET Version](https://img.shields.io/badge/.NET-10-512BD4?style=flat-square&logo=dotnet)](csharp/Directory.Build.props)
[![CI](https://img.shields.io/github/actions/workflow/status/ncosentino/google-search-console-mcp/ci.yml?label=CI&style=flat-square)](https://github.com/ncosentino/google-search-console-mcp/actions/workflows/ci.yml)

> **Zero-dependency MCP server for Google Search Console.**
> Pre-built native binaries for Linux, macOS, and Windows. No Node.js. No Python. No .NET runtime. No Go toolchain. Download one binary and configure your AI tool.

Expose real Google organic search data and per-URL index status directly to AI assistants like Claude, GitHub Copilot, and Cursor via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io). Ask your AI which queries are driving traffic, identify ranking opportunities, diagnose CTR issues, and inspect Google's indexed view of a known URL.

---

## Why This Exists

AI assistants are powerful at analyzing SEO strategy -- but they need real data. This MCP server bridges your AI tool to Google Search Console, giving it:

- **Real organic search performance** -- clicks, impressions, CTR, and average position for every query and page in your site
- **Up to 50,000 rows per query** (vs 1,000 in the UI), unlocking complete keyword datasets
- **16 months of history** for trend analysis
- **Dimension flexibility** -- group by query, page, country, device, date, or any combination
- **Per-URL index inspection** -- retrieve Google's index status and available mobile usability, rich results, and AMP details for a known URL

With this MCP server configured, you can ask your AI: _"Which queries am I ranking position 8-15 for with high impressions but low CTR? What pages could I optimize to break into the top 5?"_ and get a real data-backed answer.

---

## Quick Start

**Three steps: create a service account, download a binary, add it to your MCP config.**

### Step 1: Create a Google Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/) and create or select a project
2. Enable the **Google Search Console API** for that project:
   - Go to `https://console.cloud.google.com/apis/library/searchconsole.googleapis.com`
   - Click **Enable**
3. Create a **Service Account** (IAM & Admin → Service Accounts → Create Service Account)
   - Give it a name (e.g. `gsc-mcp`) -- no project-level roles needed
4. Click on the service account → **Keys** tab → **Add Key → Create new key → JSON**
   - Download the JSON file -- the `client_email` field inside is the email you'll use in step 5
5. Go to [Google Search Console](https://search.google.com/search-console) → select your property → **Settings → Users and permissions → Add user**
   - Enter the `client_email` from the JSON file exactly (e.g. `gsc-mcp@my-project.iam.gserviceaccount.com`)
   - Set permission to **Full** → **Add**

> **Note:** When adding the service account in Search Console, use the exact email from the `client_email` field in the downloaded JSON -- not a manually constructed one.

> **Note on property URL format:** Search Console has two property types. If your property was added as a domain property, the site URL is `sc-domain:example.com` (not `https://www.example.com/`). Use `list_sites` to discover the correct format for your property.

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
      "type": "stdio",
      "command": "/path/to/gsc-mcp-go-linux-amd64",
      "args": [],
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
| `search_type` | string | No | `web` | Which Google Search results the metrics come from: `web`, `image`, `video`, `news`, `discover`, `googleNews`. Note: `video` reports Google Video search performance, not the Video Indexing report. See [full docs](https://www.devleader.ca/projects/google-search-console-mcp/tools/query-search-analytics/#search-types) for details. |

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

### `inspect_url`

Inspect Google's indexed version of one known URL.

**Parameters:**

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `site_url` | string | Yes | -- | The Search Console property |
| `inspection_url` | string | Yes | -- | Fully qualified URL under the property to inspect |
| `language_code` | string | No | `en-US` | BCP-47 language code for translated issue messages |

The result includes index status and any available per-URL mobile usability, rich results, and AMP details.

> **Scope boundary:** This tool inspects Google's indexed version of one known URL. It is not a live URL test, an indexing request, or a bulk Page Indexing report. Google applies per-site and per-project URL Inspection API quotas.

**Example prompt:**

> "Inspect https://www.example.com/my-page and explain whether Google considers it indexed."

### Response Structure

`query_search_analytics` returns:

```json
{
  "siteUrl": "https://www.example.com/",
  "startDate": "2026-01-01",
  "endDate": "2026-01-31",
  "dimensions": ["query", "page"],
  "searchType": "web",
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

## Transports

Both binaries default to **stdio** and also support one shared Streamable HTTP
service for every agent session:

```bash
./gsc-mcp-go-linux-amd64 \
  --transport http \
  --listen-address 127.0.0.1 \
  --port 8081 \
  --service-account-file /path/to/key.json
```

The MCP endpoint is `/mcp`; supervisors can probe `/health`.

```json
{
  "mcpServers": {
    "search-console": {
      "type": "http",
      "url": "http://127.0.0.1:8081/mcp",
      "tools": ["*"]
    }
  }
}
```

Both implementations default to loopback, validate the Host and Origin
boundaries, limit request sizes and HTTP timeouts, and run statelessly without
session affinity.

The HTTP host does not authenticate ordinary MCP callers. Keep it on loopback
or place it behind TLS and an authenticated reverse proxy. See
[Shared Service](https://www.devleader.ca/projects/google-search-console-mcp/shared-service/)
and [Transports](https://www.devleader.ca/projects/google-search-console-mcp/transports/).

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
| Transports | stdio, HTTP | stdio, HTTP |

**Recommendation:** Both work great. Pick Go for smaller binary size and faster startup; pick C# if you prefer the .NET ecosystem. Both support the same [HTTP transport](#transports).

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

This MCP server was built by **[Nick Cosentino](https://www.devleader.ca)**, a software engineer and content creator known as **Dev Leader**. Nick creates practical .NET, C#, ASP.NET Core, Blazor, and software engineering content for intermediate to advanced developers -- covering everything from performance optimization and clean architecture to real-world career advice.

This tool was born out of real work analyzing search performance for [devleader.ca](https://www.devleader.ca) and the desire to use AI assistants effectively during that process. It serves as a practical example of building Native AOT C# and idiomatic Go MCP servers with zero runtime dependencies.

**Find Nick online:**

- Blog: [https://www.devleader.ca](https://www.devleader.ca)
- YouTube: [https://www.youtube.com/@devleaderca](https://www.youtube.com/@devleaderca)
- Newsletter: [https://weekly.devleader.ca](https://weekly.devleader.ca)
- LinkedIn: [https://linkedin.com/in/nickcosentino](https://linkedin.com/in/nickcosentino)
- All My Links: [https://links.devleader.ca](https://links.devleader.ca)

### BrandGhost

[BrandGhost](https://www.brandghost.ai) is a social media automation platform built by Nick that lets content creators cross-post and schedule content across all social platforms in one click. If you create content and want to spend less time on distribution and more time creating, check it out.

---

## Contributing

Contributions are welcome! Please:

1. Open an issue describing the bug or feature request before submitting a PR
2. Run `golangci-lint run` (Go) or `dotnet build` with zero warnings (C#) before submitting
3. Keep both implementations in sync -- a feature added to Go should also be added to C#, and vice versa

---

## License

MIT License -- see [LICENSE](LICENSE) for details.
