---
description: Google Search Console MCP server credential and property configuration reference -- resolution order, environment variables, .env file, and site URL normalization rules.
---

# Configuration

Credential resolution uses this priority order (highest to lowest):

---

## 1. CLI Argument (Highest Priority)

Pass the service account file path directly on the command line:

```bash
/path/to/gsc-mcp-go-linux-amd64 --service-account-file /path/to/key.json
```

---

## 2. Environment Variable -- File Path

Set `GOOGLE_SERVICE_ACCOUNT_FILE` to the path of your service account JSON key file:

```bash
export GOOGLE_SERVICE_ACCOUNT_FILE=/path/to/service-account.json
```

---

## 3. Environment Variable -- JSON Content

Set `GOOGLE_SERVICE_ACCOUNT_JSON` to the raw JSON content of the service account key:

```bash
export GOOGLE_SERVICE_ACCOUNT_JSON='{"type":"service_account","client_email":"..."}'
```

Useful in environments where secrets are stored as string values rather than files (e.g. containers, CI/CD secrets).

---

## 4. `.env` File (Lowest Priority -- Dev Convenience)

Create a `.env` file in the working directory:

```
GOOGLE_SERVICE_ACCOUNT_FILE=/path/to/service-account.json
```

Or with inline JSON:

```
GOOGLE_SERVICE_ACCOUNT_JSON={"type":"service_account",...}
```

---

## Property URL Format

Search Console has two property types. The server accepts flexible input and normalizes automatically, but the underlying API requires a specific canonical form:

| Property type | Canonical form | Example |
|---|---|---|
| Domain property | `sc-domain:example.com` | `sc-domain:devleader.ca` |
| URL-prefix property | `https://www.example.com/` | `https://www.devleader.ca/` |

The server accepts bare domains, full URLs, or canonical forms as input to `site_url` parameters and normalizes them. On 403 errors, it automatically retries with property discovery to handle format mismatches.

Use [`list_sites`](tools/list-sites.md) to see the exact canonical URLs your service account has access to.
