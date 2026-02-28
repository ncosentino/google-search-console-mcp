---
description: Reference for the list_sitemaps MCP tool -- parameters, response format, and example prompts for listing sitemaps submitted to Google Search Console for a property.
---

# list_sitemaps

List sitemaps submitted to Google Search Console for a specific property, including submission status and error counts.

---

## Parameters

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `site_url` | string | Yes | The Search Console property. Accepts bare domain (`devleader.ca`), full URL (`https://www.devleader.ca`), or canonical GSC form. |

---

## Response

```json
{
  "siteUrl": "https://www.example.com/",
  "sitemaps": [
    {
      "path": "https://www.example.com/sitemap.xml",
      "lastSubmitted": "2026-01-15T10:30:00Z",
      "isPending": false,
      "isSitemapsIndex": false,
      "type": "sitemap",
      "lastDownloaded": "2026-01-15T11:00:00Z",
      "warnings": 0,
      "errors": 0
    }
  ]
}
```

**Field notes:**

- `isPending` -- true if Google hasn't processed this sitemap yet
- `isSitemapsIndex` -- true if this is a sitemap index file referencing other sitemaps
- `type` -- typically `sitemap` or `atomFeed` or `rssFeed`
- `warnings` and `errors` -- count of issues Google found when parsing the sitemap

---

## Example Prompts

> "Show me all submitted sitemaps for my site and whether any have errors."

> "When was my sitemap last submitted and downloaded?"

> "Do I have any sitemap errors I should fix?"

---

## Notes

- Sitemap data reflects what Google has indexed, not real-time status.
- A sitemap with `errors > 0` means some URLs couldn't be parsed -- this can cause pages to be excluded from the index.
