---
description: Reference for the inspect_url MCP tool -- inspect Google's indexed status and available per-URL mobile usability, rich results, and AMP details.
---

# inspect_url

Inspect Google's indexed version of one known URL under a Search Console property.

The tool returns the complete `inspectionResult` supplied by Google's URL Inspection API, including index status and any available per-URL mobile usability, rich results, and AMP details.

---

## Parameters

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `site_url` | string | Yes | -- | The Search Console property. Accepts bare domain (`devleader.ca`), full URL (`https://www.devleader.ca`), or canonical GSC form. |
| `inspection_url` | string | Yes | -- | Fully qualified URL under the property to inspect. |
| `language_code` | string | No | `en-US` | BCP-47 language code used for translated issue messages. |

---

## Response

```json
{
  "siteUrl": "sc-domain:example.com",
  "inspectionUrl": "https://www.example.com/my-page",
  "languageCode": "en-US",
  "inspectionResult": {
    "inspectionResultLink": "https://search.google.com/search-console/inspect/...",
    "indexStatusResult": {
      "verdict": "PASS",
      "coverageState": "Submitted and indexed",
      "robotsTxtState": "ALLOWED",
      "indexingState": "INDEXING_ALLOWED",
      "pageFetchState": "SUCCESSFUL",
      "googleCanonical": "https://www.example.com/my-page",
      "userCanonical": "https://www.example.com/my-page"
    },
    "mobileUsabilityResult": {
      "verdict": "PASS"
    },
    "richResultsResult": {
      "verdict": "PASS",
      "detectedItems": []
    },
    "ampResult": {
      "verdict": "PASS"
    }
  },
  "queriedAt": "2026-07-14T20:00:00Z"
}
```

`mobileUsabilityResult`, `richResultsResult`, and `ampResult` are optional. Google omits sections that do not apply to the inspected URL.

---

## Example Prompts

> "Inspect https://www.example.com/my-page and explain whether Google considers it indexed."

> "Compare the Google-selected canonical and user-declared canonical for this URL."

> "Does this page have any rich result, mobile usability, or AMP issues?"

---

## Scope and Quotas

`inspect_url` checks Google's indexed information for one URL supplied by the caller.

It does **not**:

- test the live version of the page
- request indexing
- enumerate every indexed or excluded URL for a property
- reproduce the bulk Page Indexing report from the Search Console UI

Google applies per-site and per-project request quotas to URL Inspection. See the [official Search Console API usage limits](https://developers.google.com/webmaster-tools/limits) for current quota details.

Index status reflects Google's stored information and can lag behind changes to the live page.
