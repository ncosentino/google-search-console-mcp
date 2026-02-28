---
description: Overview of the three MCP tools exposed by the Google Search Console MCP server -- query_search_analytics, list_sites, and list_sitemaps.
---

# MCP Tools

Three tools are exposed by this MCP server. All tools work identically across the Go and C# implementations.

| Tool | Description |
|------|-------------|
| [`query_search_analytics`](query-search-analytics.md) | Query clicks, impressions, CTR, and average position |
| [`list_sites`](list-sites.md) | List all Search Console properties the service account can access |
| [`list_sitemaps`](list-sitemaps.md) | List submitted sitemaps and their status for a property |

---

## Common Notes

### Property URL (`site_url`)

All tools that accept a `site_url` parameter support flexible input:

- Bare domain: `devleader.ca`
- Full URL: `https://www.devleader.ca`
- Canonical GSC form: `sc-domain:devleader.ca` or `https://www.devleader.ca/`

The server normalizes the input and automatically retries with property discovery on 403 errors.

### Error Responses

All tools return a JSON error object when an exception occurs:

```json
{
  "error": "ExceptionType: message"
}
```
