---
description: Run one Google Search Console MCP service shared by every agent session.
---

# Running One Shared Service

Streamable HTTP lets every local agent connect to one long-lived Search Console
process instead of launching one STDIO child per session.

## Start the service

Place `GOOGLE_SERVICE_ACCOUNT_FILE` in the service environment or in a `.env`
file beside the binary, then start the server:

```bash
./gsc-mcp-go-linux-amd64 \
  --transport http \
  --listen-address 127.0.0.1 \
  --port 8081
```

The C# Native AOT binary uses the same arguments.

## Lifecycle management

Use a platform service supervisor, or reuse the generic
[`manage-mcp-service.ps1`](https://github.com/ncosentino/google-psi-mcp/blob/main/scripts/manage-mcp-service.ps1)
maintained for the native NexusLabs MCP servers:

```powershell
.\manage-mcp-service.ps1 Start `
  -ServiceName google-search-console-mcp `
  -BinaryPath C:\path\to\gsc-mcp-go.exe `
  -Port 8081
```

The manager health-checks before starting, serializes concurrent start
attempts, records process identity, and uses a per-run authenticated loopback
shutdown before falling back to terminating the verified process.

## Configure Copilot CLI

```json
{
  "mcpServers": {
    "google-search-console": {
      "type": "http",
      "url": "http://127.0.0.1:8081/mcp",
      "tools": ["*"]
    }
  }
}
```

Remove `command`, `args`, and `env` from the HTTP entry because Copilot no
longer launches the process. Existing sessions retain their STDIO child until
they restart.

## Start automatically

Run the manager's `Start` action from a per-user service, scheduled task, or
Copilot `sessionStart` hook. Repeated calls reuse the healthy process.

## Network deployment

The included manager is deliberately intended for loopback services.
Non-loopback hosting requires a platform supervisor, TLS, authentication and
authorization on every request, trusted proxy handling, and ingress limits.
