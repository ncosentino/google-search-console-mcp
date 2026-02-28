---
description: Reference for the list_sites MCP tool -- response format, permission levels, and example prompts for listing Google Search Console properties accessible to the service account.
---

# list_sites

List all Google Search Console properties (sites) the service account has access to, along with permission levels.

---

## Parameters

No parameters required.

---

## Response

```json
{
  "sites": [
    {
      "siteUrl": "sc-domain:devleader.ca",
      "permissionLevel": "siteFullUser"
    },
    {
      "siteUrl": "https://www.devleader.ca/",
      "permissionLevel": "siteOwner"
    }
  ]
}
```

**Permission levels:**

| Level | Description |
|-------|-------------|
| `siteOwner` | Full owner access |
| `siteFullUser` | Full read/write access |
| `siteRestrictedUser` | Read-only access to aggregated data |
| `siteUnverifiedUser` | No verified access |

---

## Example Prompts

> "What Search Console properties do I have access to?"

> "What's the correct site URL to use for my domain?"

---

## Notes

- Use this tool first if you're getting 403 errors on other tools -- it confirms the exact property URL format your service account has access to.
- A domain may appear twice: once as a domain property (`sc-domain:...`) and once as a URL-prefix property (`https://...`). These are separate properties with separate data.
