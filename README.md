# sentinelone-mcp-server

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/c0tton-fluff/sentinelone-mcp-server)](https://github.com/c0tton-fluff/sentinelone-mcp-server/releases)

MCP server for [SentinelOne](https://www.sentinelone.com/) integration. Enables AI assistants like Claude Code to manage threats, investigate endpoints, query alerts, and run Deep Visibility hunts.

## Features

- **Threat management** - List, inspect, and mitigate threats (kill, quarantine, remediate, rollback)
- **Agent operations** - List agents, get details, network isolate/reconnect endpoints
- **Unified alerts** - Query alerts via GraphQL with severity, verdict, and storyline filters
- **Hash intelligence** - SHA1/SHA256 reputation lookups
- **Deep Visibility** - Run threat hunting queries with automatic polling and pagination
- **Error sanitization** - API keys are redacted from all error messages
- **Single binary** - No runtime dependencies, just copy and run

## Installation

```bash
git clone https://github.com/c0tton-fluff/sentinelone-mcp-server.git
cd sentinelone-mcp-server
go build -o sentinelone-mcp-server .
```

## Quick Start

**1. Get your SentinelOne API token**

Log into your S1 console > Profile icon (top right) > **My Profile** > Actions > **API token operations** > **Regenerate API token**.

**2. Configure MCP client**

Add to `~/.mcp.json`:

```json
{
  "mcpServers": {
    "sentinelone": {
      "command": "/path/to/sentinelone-mcp-server",
      "env": {
        "SENTINELONE_API_KEY": "your_api_token_here",
        "SENTINELONE_API_BASE": "https://your-tenant.sentinelone.net"
      }
    }
  }
}
```

**3. Use with Claude Code**

```
"List all unmitigated threats"
"Show infected agents"
"Isolate agent 1234567890"
"What's the reputation of this SHA256?"
"Hunt for PowerShell processes in the last 24 hours"
"Show high severity alerts"
```

## Tools Reference

### Threats

| Tool | Description |
|------|-------------|
| `s1_list_threats` | List threats with classification, status, and computer filters |
| `s1_get_threat` | Get threat details including hashes, file path, and storyline |
| `s1_mitigate_threat` | Kill, quarantine, remediate, or rollback a threat |

### Agents

| Tool | Description |
|------|-------------|
| `s1_list_agents` | List agents with OS, status, and infection filters |
| `s1_get_agent` | Get agent details including version, site, and network info |
| `s1_isolate_agent` | Network isolate an endpoint (maintains S1 communication) |
| `s1_reconnect_agent` | Remove network isolation from an agent |

### Alerts

| Tool | Description |
|------|-------------|
| `s1_list_alerts` | Query unified alerts via GraphQL with severity and verdict filters |

### Intelligence

| Tool | Description |
|------|-------------|
| `s1_hash_reputation` | SHA1 or SHA256 hash reputation lookup |

### Deep Visibility

| Tool | Description |
|------|-------------|
| `s1_dv_query` | Run a Deep Visibility query with automatic status polling |
| `s1_dv_get_events` | Retrieve events from a completed DV query |

<details>
<summary>Full parameter reference</summary>

### s1_list_threats
| Parameter | Type | Description |
|-----------|------|-------------|
| `computerName` | string | Search by computer/endpoint name (partial match) |
| `threatName` | string | Search by threat name (partial match) |
| `limit` | number | Max results (default 10, max 50) |
| `mitigationStatuses` | string[] | Filter: not_mitigated, mitigated, marked_as_benign |
| `classifications` | string[] | Filter: Malware, PUA, Suspicious |

### s1_get_threat
| Parameter | Type | Description |
|-----------|------|-------------|
| `threatId` | string | The threat ID to retrieve |

### s1_mitigate_threat
| Parameter | Type | Description |
|-----------|------|-------------|
| `threatId` | string | The threat ID to mitigate |
| `action` | string | kill, quarantine, remediate, rollback-remediation |

### s1_list_agents
| Parameter | Type | Description |
|-----------|------|-------------|
| `computerName` | string | Search by computer name (partial match) |
| `limit` | number | Max results (default 10, max 50) |
| `osTypes` | string[] | Filter by OS: windows, macos, linux |
| `isActive` | boolean | Filter by active status |
| `isInfected` | boolean | Filter by infected status |
| `networkStatuses` | string[] | Filter: connected, disconnected |

### s1_get_agent
| Parameter | Type | Description |
|-----------|------|-------------|
| `agentId` | string | The agent ID to retrieve |

### s1_isolate_agent / s1_reconnect_agent
| Parameter | Type | Description |
|-----------|------|-------------|
| `agentId` | string | The agent ID |

### s1_list_alerts
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | number | Max results (default 20, max 50) |
| `cursor` | string | Pagination cursor from previous response |
| `severity` | string | Filter: Low, Medium, High, Critical |
| `analystVerdict` | string | Filter: TruePositive, FalsePositive, Suspicious, Undefined |
| `incidentStatus` | string | Filter: Unresolved, InProgress, Resolved |
| `siteIds` | string[] | Filter by site IDs |
| `storylineId` | string | Filter by storyline ID (correlate with threat) |

### s1_hash_reputation
| Parameter | Type | Description |
|-----------|------|-------------|
| `hash` | string | SHA1 (40 chars) or SHA256 (64 chars) hash |

### s1_dv_query
| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | string | Deep Visibility query (e.g., ProcessName Contains "python") |
| `fromDate` | string | Start date in ISO format |
| `toDate` | string | End date in ISO format |
| `siteIds` | string[] | Filter by site IDs |
| `groupIds` | string[] | Filter by group IDs |
| `accountIds` | string[] | Filter by account IDs |

### s1_dv_get_events
| Parameter | Type | Description |
|-----------|------|-------------|
| `queryId` | string | Query ID from s1_dv_query |
| `limit` | number | Max results (default 50, max 100) |
| `cursor` | string | Pagination cursor |

</details>

## Troubleshooting

| Error | Fix |
|-------|-----|
| Configuration error | Ensure SENTINELONE_API_KEY and SENTINELONE_API_BASE are set |
| HTTP 401 | API token expired or invalid - regenerate in S1 console |
| HTTP 403 | Token lacks permissions for this endpoint |
| Request timeout | S1 API took >30s - try narrowing your query filters |
| Tools not appearing | Verify binary path in `~/.mcp.json`, restart Claude Code |

Check MCP logs: `~/.cache/claude-cli-nodejs/*/mcp-logs-sentinelone/`

## License

MIT
