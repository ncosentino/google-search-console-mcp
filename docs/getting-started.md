# Getting Started

Three steps: create a Google service account, download a binary, add it to your AI tool config.

---

## Step 1: Create a Google Service Account

1. Go to [Google Cloud Console](https://console.cloud.google.com/) and create or select a project.
2. Enable the **Google Search Console API**:
   - Navigate to `https://console.cloud.google.com/apis/library/searchconsole.googleapis.com`
   - Click **Enable**
3. Create a **Service Account** (IAM & Admin → Service Accounts → Create Service Account):
   - Give it a name (e.g. `gsc-mcp`) -- no project-level roles needed
4. Click on the service account → **Keys** tab → **Add Key → Create new key → JSON**
   - Download the JSON file -- the `client_email` field is the email you'll add to Search Console
5. Go to [Google Search Console](https://search.google.com/search-console) → select your property → **Settings → Users and permissions → Add user**:
   - Enter the `client_email` from the JSON file exactly
   - Set permission to **Full** → **Add**

!!! note "Service account email"
    Use the exact email from `client_email` in the downloaded JSON -- not a manually constructed one.

!!! note "Property URL format"
    Search Console has two property types. Domain properties use `sc-domain:example.com`; URL-prefix properties use `https://www.example.com/`. Use [`list_sites`](tools/list-sites.md) to discover the correct format for your property.

!!! info "No billing required"
    The Search Console API is free. No billing account or payment method is needed.

---

## Step 2: Download a Binary

Go to the [Releases page](https://github.com/ncosentino/google-search-console-mcp/releases/latest) and download the binary for your platform:

| Platform | Go binary | C# binary |
|----------|-----------|-----------|
| Linux x64 | `gsc-mcp-go-linux-amd64` | `gsc-mcp-csharp-linux-x64` |
| Linux arm64 | `gsc-mcp-go-linux-arm64` | `gsc-mcp-csharp-linux-arm64` |
| macOS x64 (Intel) | `gsc-mcp-go-darwin-amd64` | `gsc-mcp-csharp-osx-x64` |
| macOS arm64 (Apple Silicon) | `gsc-mcp-go-darwin-arm64` | `gsc-mcp-csharp-osx-arm64` |
| Windows x64 | `gsc-mcp-go-windows-amd64.exe` | `gsc-mcp-csharp-win-x64.exe` |
| Windows arm64 | `gsc-mcp-go-windows-arm64.exe` | `gsc-mcp-csharp-win-arm64.exe` |

Not sure which to pick? See [Go vs C#](implementations.md) for a comparison.

On Linux/macOS, make the binary executable after downloading:

```bash
chmod +x gsc-mcp-go-linux-amd64
```

---

## Step 3: Configure Your AI Tool

Add the server to your AI tool's MCP configuration. See [Setup by Tool](setup-by-tool.md) for tool-specific instructions.

The minimal config pattern (replace paths with your actual locations):

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

## Next Steps

- [MCP Tools Reference](tools/index.md) -- full parameter documentation for all three tools
- [Configuration](configuration.md) -- credential resolution order and all configuration options
- [Setup by Tool](setup-by-tool.md) -- exact config snippets for Claude, Cursor, VS Code, Visual Studio
