#!/bin/bash
#
# Wrapper that loads SentinelOne credentials from the macOS login keychain and
# exec's the MCP server. Point your MCP client at this script instead of putting
# the API token in plaintext in ~/.mcp.json.
#
# Setup:
#   1. Store the credentials (see store-token.sh for the token):
#        ./store-token.sh
#        security add-generic-password -s sentinelone-mcp -a api-base \
#          -w "https://your-tenant.sentinelone.net"
#   2. Point your MCP client at this script:
#        { "mcpServers": { "sentinelone": { "command": "/path/to/run.sh" } } }
#
# Why the JWT check below: S1 tokens are long ES256 JWTs, and anything that
# truncates them on the way into storage yields a value that authenticates with
# HTTP 401 -- the same response an expired token gives. Failing loudly here turns
# a silent, misleading auth error into a named cause at startup.

set -euo pipefail

KEYCHAIN_SERVICE="sentinelone-mcp"

# MCP servers do not inherit your interactive shell's PATH.
export PATH="/usr/bin:/bin:/usr/sbin:/sbin"

# Defaults to the binary in the repo root (this script lives in examples/).
# Override with SENTINELONE_MCP_BIN if you keep the binary elsewhere.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="${SENTINELONE_MCP_BIN:-${SCRIPT_DIR}/../sentinelone-mcp-server}"

die() { echo "run.sh: $*" >&2; exit 1; }

[[ -x "$BIN" ]] || die "server binary not found or not executable at $BIN (run 'go build .', or set SENTINELONE_MCP_BIN)"

# Read a field from the keychain, stripping any trailing newline. A stray \n in
# the stored value is itself enough to fail authentication.
kc() {
  security find-generic-password -s "$KEYCHAIN_SERVICE" -a "$1" -w 2>/dev/null | tr -d '\n\r'
}

SENTINELONE_API_KEY=$(kc api-key) || die "no 'api-key' in keychain service '$KEYCHAIN_SERVICE' -- run store-token.sh"
SENTINELONE_API_BASE=$(kc api-base) || die "no 'api-base' in keychain service '$KEYCHAIN_SERVICE'"

[[ -n "$SENTINELONE_API_KEY" ]] || die "api-key is present but empty"
[[ -n "$SENTINELONE_API_BASE" ]] || die "api-base is present but empty"

# ES256 JWT: header.payload.signature -- exactly two separators.
dots=$(printf '%s' "$SENTINELONE_API_KEY" | tr -cd '.' | wc -c | tr -d ' ')
[[ "$dots" == "2" ]] || die "api-key is not a well-formed JWT (found $dots '.' separators, expected 2) -- the stored value is truncated; re-store it with store-token.sh rather than regenerating the token"

export SENTINELONE_API_KEY SENTINELONE_API_BASE
exec "$BIN"

# ---------------------------------------------------------------------------
# Not on macOS? The pattern is the same -- resolve the secret at startup from
# whatever store you already trust, then exec. Swap the kc() body for one of:
#
#   secret-tool lookup service sentinelone-mcp key api-key        # Linux/libsecret
#   pass show sentinelone/api-key                                 # pass
#   op read "op://vault/SentinelOne/credential"                    # 1Password CLI
#   aws secretsmanager get-secret-value --secret-id s1 \
#     --query SecretString --output text                           # AWS
#
# Keep the JWT check regardless of the backend -- it is the store-agnostic part.
# ---------------------------------------------------------------------------
