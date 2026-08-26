# SentinelOne MCP Server

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/c0tton-fluff/sentinelone-mcp-server)](https://github.com/c0tton-fluff/sentinelone-mcp-server/releases)

A [Model Context Protocol](https://modelcontextprotocol.io/) server that connects AI assistants to your SentinelOne tenant. Manage threats, investigate endpoints, hunt with Deep Visibility, and triage alerts -- all from natural language.

**Zero dependencies.** Stdlib-only Go binary. No runtime requirements. Just copy and run.

---

## Quick Start

### 1. Build

```bash
git clone https://github.com/c0tton-fluff/sentinelone-mcp-server.git
cd sentinelone-mcp-server
go build -o sentinelone-mcp-server .
```

### 2. Get your API token

**Recommended -- a Service User** (decoupled from your personal login, independently scoped, revocable without touching your account):

S1 Console > **Settings** > **Users** > **Service Users** > Actions > **Create New Service User**

Set an expiry, and scope it to the narrowest role that covers the tools you plan to use. Read-only is enough for everything except the mitigate/isolate/exclusion/STAR tools.

**Or a personal token** (dies when your account is deactivated or your role changes, and carries your full permissions):

S1 Console > Profile (top right) > **My Profile** > Actions > **API token operations** > **Regenerate API token**

> **The token is displayed once.** Copy it with the console's Copy button rather than selecting the text -- the field is scrollable, and a click-drag can silently grab only the visible portion.

#### Verify the token survived storage

Modern S1 tokens are **ES256 JWTs** several hundred characters long, of the form `header.payload.signature`. Anything that truncates a long string on the way into storage will produce a value that *looks* plausible but fails authentication.

[`examples/store-token.sh`](examples/store-token.sh) does these checks for you and refuses to overwrite a working token with a bad paste. To check by hand:

```bash
# expect: a length in the hundreds, and exactly 2 dot separators
printf '%s' "$SENTINELONE_API_KEY" | wc -c
printf '%s' "$SENTINELONE_API_KEY" | tr -cd '.' | wc -c
```

Known truncation trap: on macOS, `security add-generic-password -w` with **no value** prompts interactively and cuts the input off at 128 characters. Pass the value instead -- `-w "$(pbpaste | tr -d '\n\r')"` -- and note that Keychain Access.app was removed in macOS 26, so the CLI is the only way to write these items. A trailing newline in the stored value also fails auth, hence the `tr`.

### 3. Configure your MCP client

**Recommended: a wrapper script that resolves the token at startup.** Ready-made scripts are in [`examples/`](examples/).

```bash
# Store the token (reads it from your clipboard -- no arguments, so it never
# lands in shell history or `ps` output). Validates before writing.
./examples/store-token.sh

# Store your tenant URL
security add-generic-password -s sentinelone-mcp -a api-base \
  -w "https://your-tenant.sentinelone.net"
```

Then point your client at the wrapper -- no `env` block, no secret in the config file:

```json
{
  "mcpServers": {
    "sentinelone": {
      "command": "/path/to/sentinelone-mcp-server/examples/run.sh"
    }
  }
}
```

`run.sh` reads both values from the keychain, rejects a malformed token at startup rather than letting it surface as a confusing HTTP 401, and exec's the server. It's macOS/keychain by default; the footer comment shows the one-line swap for `secret-tool`, `pass`, `op`, or AWS Secrets Manager.

<details>
<summary><strong>Alternative: token inline in the config (simpler, less safe)</strong></summary>

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

This writes a live API token in cleartext to a config file that is easy to sync to a dotfiles repo, back up, or share while screen-sharing. It also has no truncation check, so a mangled paste shows up as an authentication failure with no indication of the real cause. Fine for a throwaway or a scoped read-only token you rotate often; the wrapper is better for anything else.

</details>

### 4. Go

```
"List all unmitigated threats"
"Investigate threat 1234567890"
"Show infected agents"
"Hunt for PowerShell processes in the last 24 hours"
"What's the reputation of this SHA256?"
"Create an exclusion for /opt/myapp on Linux"
"What applications are installed on Benedict's laptop?"
```

---

## Tools (21)

### Threats

| Tool | What it does |
|------|--------------|
| `s1_list_threats` | List threats with classification, status, and endpoint filters |
| `s1_get_threat` | Full threat details -- hashes, file path, storyline |
| `s1_mitigate_threat` | Kill, quarantine, un-quarantine, remediate, or rollback |
| `s1_investigate_threat` | One-call investigation: threat + correlated alerts + timeline |
| `s1_set_analyst_verdict` | Set verdict: true_positive, false_positive, suspicious, undefined |
| `s1_set_incident_status` | Set status (with optional verdict in the same call) |

### Agents

| Tool | What it does |
|------|--------------|
| `s1_list_agents` | List agents with OS, status, infection filters, and count-by grouping |
| `s1_get_agent` | Agent details -- version, site, network info, account ID |
| `s1_isolate_agent` | Network isolate an endpoint (maintains S1 comms) |
| `s1_reconnect_agent` | Remove network isolation |

### Alerts

| Tool | What it does |
|------|--------------|
| `s1_list_alerts` | Query unified alerts via GraphQL with severity, verdict, and status filters |
| `s1_set_alert_verdict` | Bulk set analyst verdict on matching alerts |
| `s1_set_alert_status` | Bulk set incident status (with optional verdict) |

### Deep Visibility

| Tool | What it does |
|------|--------------|
| `s1_dv_query` | Run a threat hunting query with automatic polling |
| `s1_dv_get_events` | Retrieve events from a completed query |

### Intelligence

| Tool | What it does |
|------|--------------|
| `s1_hash_reputation` | Hash verdict + fleet-wide hunt via Deep Visibility |

### Exclusions

| Tool | What it does |
|------|--------------|
| `s1_list_exclusions` | List exclusions (path, hash, certificate, browser, file type) |
| `s1_create_exclusion` | Create an exclusion to suppress false-positive detections |
| `s1_delete_exclusion` | Delete exclusions by ID |

### STAR Rules

| Tool | What it does |
|------|--------------|
| `s1_create_star_rule` | Create a custom detection rule from a Deep Visibility query |

### Applications

| Tool | What it does |
|------|--------------|
| `s1_list_applications` | List installed software on endpoints by name or computer |

---

<details>
<summary><strong>Full parameter reference</strong></summary>

### s1_list_threats
| Parameter | Type | Description |
|-----------|------|-------------|
| `computerName` | string | Search by endpoint name (partial match) |
| `threatName` | string | Search by threat name (partial match) |
| `limit` | number | Max results (default 50, max 200) |
| `mitigationStatuses` | string[] | not_mitigated, mitigated, marked_as_benign |
| `classifications` | string[] | Malware, PUA, Suspicious |

### s1_get_threat
| Parameter | Type | Description |
|-----------|------|-------------|
| `threatId` | string | **Required.** The threat ID |

### s1_mitigate_threat
| Parameter | Type | Description |
|-----------|------|-------------|
| `threatId` | string | **Required.** The threat ID |
| `action` | string | **Required.** kill, quarantine, un-quarantine, remediate, rollback-remediation |

### s1_investigate_threat
| Parameter | Type | Description |
|-----------|------|-------------|
| `threatId` | string | **Required.** The threat ID |

### s1_set_analyst_verdict
| Parameter | Type | Description |
|-----------|------|-------------|
| `threatId` | string | **Required.** The threat ID |
| `verdict` | string | **Required.** true_positive, false_positive, suspicious, undefined |

### s1_set_incident_status
| Parameter | Type | Description |
|-----------|------|-------------|
| `threatId` | string | **Required.** The threat ID |
| `status` | string | **Required.** unresolved, in_progress, resolved |
| `verdict` | string | Optional: set analyst verdict in the same call |

### s1_list_agents
| Parameter | Type | Description |
|-----------|------|-------------|
| `computerName` | string | Search by computer name (partial match) |
| `limit` | number | Max results (default 50, max 200) |
| `osTypes` | string[] | windows, macos, linux |
| `isActive` | boolean | Filter by active status |
| `isInfected` | boolean | Filter by infected status |
| `networkStatuses` | string[] | connected, disconnected, connecting, disconnecting |
| `countBy` | string | Group counts by: user, os, site, group |

### s1_get_agent
| Parameter | Type | Description |
|-----------|------|-------------|
| `agentId` | string | **Required.** The agent ID |

### s1_isolate_agent / s1_reconnect_agent
| Parameter | Type | Description |
|-----------|------|-------------|
| `agentId` | string | **Required.** The agent ID |

### s1_list_alerts
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | number | Max results (default 50, max 200) |
| `severity` | string | LOW, MEDIUM, HIGH, CRITICAL |
| `analystVerdict` | string | TRUE_POSITIVE, FALSE_POSITIVE, SUSPICIOUS, UNDEFINED |
| `incidentStatus` | string | NEW, IN_PROGRESS, RESOLVED (aliases: unresolved, open) |
| `siteIds` | string[] | Filter by site IDs |
| `storylineId` | string | Correlate with threat by storyline ID |

### s1_set_alert_verdict
| Parameter | Type | Description |
|-----------|------|-------------|
| `verdict` | string | **Required.** TRUE_POSITIVE, FALSE_POSITIVE, SUSPICIOUS, UNDEFINED |
| `alertIds` | string[] | Target specific alert IDs |
| `query` | string | Free-text search (use for username matching) |
| `ruleName` | string[] | Filter by rule name (partial match) |
| `agentName` | string[] | Filter by endpoint name, not username (partial match) |
| `incidentStatus` | string[] | Filter by current status |
| `siteIds` | string[] | Filter by site IDs |

### s1_set_alert_status
| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | **Required.** UNRESOLVED, IN_PROGRESS, RESOLVED |
| `verdict` | string | Optional: set analyst verdict in the same call |
| `alertIds` | string[] | Target specific alert IDs |
| `query` | string | Free-text search (use for username matching) |
| `ruleName` | string[] | Filter by rule name (partial match) |
| `agentName` | string[] | Filter by endpoint name, not username (partial match) |
| `incidentStatus` | string[] | Filter by current status |
| `siteIds` | string[] | Filter by site IDs |

### s1_hash_reputation
| Parameter | Type | Description |
|-----------|------|-------------|
| `hash` | string | **Required.** SHA1 (40 chars) or SHA256 (64 chars) |

### s1_dv_query
| Parameter | Type | Description |
|-----------|------|-------------|
| `query` | string | **Required.** Deep Visibility query (S1QL) |
| `fromDate` | string | **Required.** ISO format start date |
| `toDate` | string | **Required.** ISO format end date |
| `siteIds` | string[] | Filter by site IDs |
| `accountIds` | string[] | Filter by account IDs |

### s1_dv_get_events
| Parameter | Type | Description |
|-----------|------|-------------|
| `queryId` | string | **Required.** Query ID from s1_dv_query |
| `limit` | number | Max results (default 100, max 100) |

### s1_list_exclusions
| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | path, white_hash, certificate, browser, file_type |
| `value` | string | Search by exclusion value (partial match) |
| `osTypes` | string | windows, macos, linux (comma-separated) |
| `siteIds` | string[] | Filter by site IDs |
| `limit` | number | Max results (default 50, max 100) |

### s1_create_exclusion
| Parameter | Type | Description |
|-----------|------|-------------|
| `type` | string | **Required.** path, white_hash, certificate, browser, file_type |
| `value` | string | **Required.** The value to exclude |
| `osType` | string | **Required.** windows, macos, linux |
| `siteIds` | string[] | Scope to specific sites (use s1_get_agent to find siteId) |
| `description` | string | Reason for the exclusion |
| `mode` | string | suppress, suppress_dynamic_only, disable_in_process_monitor, disable_all_monitors |
| `pathExclusionType` | string | subfolders (default) or file |

### s1_delete_exclusion
| Parameter | Type | Description |
|-----------|------|-------------|
| `ids` | string[] | **Required.** Exclusion IDs to delete |

### s1_create_star_rule
| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | **Required.** Rule name |
| `s1ql` | string | **Required.** Deep Visibility query that triggers the rule |
| `severity` | string | **Required.** Low, Medium, High, Critical |
| `siteIds` | string[] | Scope to sites |
| `accountIds` | string[] | Scope to accounts |
| `tenant` | boolean | Tenant-wide scope |
| `description` | string | What the rule detects |
| `treatAsThreat` | string | UNDEFINED (alert only), Suspicious, Malicious |
| `networkQuarantine` | boolean | Auto-isolate endpoint on trigger |
| `expirationMode` | string | Permanent (default) or Temporary |
| `expiration` | string | ISO date (required if Temporary) |
| `status` | string | Active (default), Draft, Disabled |

### s1_list_applications
| Parameter | Type | Description |
|-----------|------|-------------|
| `name` | string | Filter by app name (partial match) |
| `agentName` | string | Filter by endpoint name (partial match) |
| `limit` | number | Max results (default 50, max 1000) |

At least one of `name` or `agentName` is required.

</details>

---

## Troubleshooting

| Error | Fix |
|-------|-----|
| Configuration error | Ensure `SENTINELONE_API_KEY` and `SENTINELONE_API_BASE` are set |
| API_BASE must be HTTPS | Use `https://` not `http://` for your tenant URL |
| HTTP 401 | Token expired, revoked, **or truncated/malformed** -- check the JWT shape first (see below), then regenerate |
| HTTP 403 | Token lacks permissions for this endpoint |
| HTTP 429 | Rate limited -- server retries automatically with backoff |
| Request timeout | S1 API took >30s -- narrow your query filters |
| Tools not appearing | Verify binary path in `~/.mcp.json`, restart your MCP client |

### Debugging HTTP 401

S1 returns the **same** error for an expired token and a malformed one:

```json
{"errors":[{"code":4010010,"detail":null,"title":"Authentication Failed"}]}
```

So "my token is valid until <date>" and a hard 401 can both be true at once. Rule out corruption before you regenerate:

```bash
# 1. Is it a well-formed JWT? Expect exactly 2.
printf '%s' "$SENTINELONE_API_KEY" | tr -cd '.' | wc -c

# 2. Does the header decode? Expect {"kid":"...","alg":"ES256"}
printf '%s' "$SENTINELONE_API_KEY" | cut -d. -f1 | base64 -d

# 3. Confirm against the API directly, bypassing this server
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: ApiToken $SENTINELONE_API_KEY" \
  "$SENTINELONE_API_BASE/web/api/v2.1/agents?limit=1"
```

Fewer than 2 dots means the stored value was truncated on write -- re-store it, don't regenerate. If `curl` also returns 401, the problem is the credential, not this server. A `403` instead means the token is valid but the Service User's role is too narrow for that endpoint.

MCP logs: `~/.cache/claude-cli-nodejs/*/mcp-logs-sentinelone/`

## License

MIT
