---
description: Zero-dependency MCP server for Google Search Console. Pre-built native binaries expose organic search analytics and per-URL index inspection directly to AI assistants like Claude, GitHub Copilot, and Cursor.
---

# Google Search Console MCP

> **Zero-dependency MCP server for Google Search Console.**
> Pre-built native binaries for Linux, macOS, and Windows. No Node.js. No Python. No .NET runtime. No Go toolchain. Download one binary and configure your AI tool.

Expose real Google organic search data and per-URL index status directly to AI assistants like Claude, GitHub Copilot, and Cursor via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io). Ask your AI which queries are driving traffic, identify ranking opportunities, diagnose CTR issues, and inspect Google's indexed view of a known URL.

---

## Why This Exists

AI assistants are powerful at analyzing search performance -- but they need real data. This MCP server bridges your AI tool to Google Search Console, giving it:

- **Real organic search performance** -- clicks, impressions, CTR, and average position for every query and page in your site
- **Up to 25,000 rows per query** (vs 1,000 in the UI), unlocking complete keyword datasets
- **16 months of history** for trend analysis
- **Dimension flexibility** -- group by query, page, country, device, date, or any combination
- **Per-URL index inspection** -- retrieve Google's index status and available mobile usability, rich results, and AMP details for a known URL

With this MCP server configured, you can ask your AI: *"Which queries am I ranking position 8-15 for with high impressions but low CTR? What pages could I optimize to break into the top 5?"* and get a real data-backed answer.

---

## Quick Overview

Four MCP tools are exposed:

| Tool | What it does |
|------|-------------|
| [`query_search_analytics`](tools/query-search-analytics.md) | Query clicks, impressions, CTR, position -- grouped by any dimension combination |
| [`list_sites`](tools/list-sites.md) | List all Search Console properties the service account can access |
| [`list_sitemaps`](tools/list-sitemaps.md) | List submitted sitemaps and their status for a property |
| [`inspect_url`](tools/inspect-url.md) | Inspect Google's indexed status and available per-URL enhancement details |

---

## Get Started

**[→ Getting Started](getting-started.md)** -- three steps: service account, binary, config.

---

## About

Built by **[Nick Cosentino](https://www.devleader.ca)** (Dev Leader) -- a software engineer and content creator covering .NET, C#, and software architecture. Available in both Go and C# (Native AOT) with zero runtime dependencies.

- Blog: [devleader.ca](https://www.devleader.ca)
- GitHub: [ncosentino/google-search-console-mcp](https://github.com/ncosentino/google-search-console-mcp)
