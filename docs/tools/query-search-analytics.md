# query_search_analytics

Query Google Search Console search analytics. Returns clicks, impressions, CTR, and average position for queries, pages, and other dimensions.

---

## Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `site_url` | string | Yes | -- | The Search Console property. Accepts bare domain (`devleader.ca`), full URL (`https://www.devleader.ca`), or canonical GSC form (`sc-domain:devleader.ca`). |
| `start_date` | string | Yes | -- | Start date in `YYYY-MM-DD` format |
| `end_date` | string | Yes | -- | End date in `YYYY-MM-DD` format |
| `dimensions` | string[] | No | `[]` | Dimensions to group by. Valid values: `query`, `page`, `country`, `device`, `date`. Pass an empty array to get aggregate totals. |
| `row_limit` | int | No | `1000` | Maximum rows to return (1--25000) |

---

## Response

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

**Field notes:**

- `keys` -- the dimension values for this row, in the same order as the `dimensions` parameter
- `ctr` -- click-through rate as a decimal (0.0372 = 3.72%)
- `position` -- average position (1.0 = first result; lower is better)
- Empty `dimensions` array returns a single aggregate row with no `keys`

---

## Example Prompts

> "Which queries are driving the most impressions to my site this month but have a CTR below 2%?"

> "Show me my top 50 pages by clicks over the last 90 days."

> "Which queries am I ranking position 8-15 for? These are my best ranking improvement opportunities."

> "What's my total click and impression count for the last 30 days?"

> "Break down my traffic by country -- which countries are sending the most organic visitors?"

---

## Dimension Combinations

| Goal | `dimensions` value |
|------|--------------------|
| Top queries | `["query"]` |
| Top pages | `["page"]` |
| Queries per page | `["query", "page"]` |
| Traffic by country | `["country"]` |
| Traffic by device | `["device"]` |
| Traffic over time | `["date"]` |
| Aggregate totals | `[]` |

---

## Notes

- Search Console data has a **2--4 day delay** -- recent dates may return incomplete data.
- The API returns up to **25,000 rows per request** (vs 1,000 rows in the Search Console UI).
- Position values are averages across all impressions for that dimension group.
- CTR is calculated as `clicks / impressions`.
