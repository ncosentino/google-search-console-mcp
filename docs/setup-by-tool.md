# Setup by Tool

Replace `/path/to/binary` with the actual path to your downloaded binary. Replace `/path/to/service-account.json` with the path to your downloaded service account JSON key file.

See [Getting Started](getting-started.md) for how to obtain the service account JSON file.

---

## Claude Code / GitHub Copilot CLI

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

---

## Claude Desktop

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

---

## Cursor

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

---

## VS Code with GitHub Copilot

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

---

## Visual Studio

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

---

## Using a CLI Argument Instead of an Environment Variable

You can also pass the service account file path as a CLI argument instead of an environment variable:

```json
{
  "command": "/path/to/binary",
  "args": ["--service-account-file", "/path/to/service-account.json"]
}
```

---

## Troubleshooting

**403 errors:** The server automatically retries with property discovery on 403 responses to handle property URL format mismatches. If you're still getting errors, use [`list_sites`](tools/list-sites.md) to confirm the exact property URL your service account can access.

**"No properties found":** Make sure you added the `client_email` from your service account JSON to the correct Search Console property with at least **Full** permission.
