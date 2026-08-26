#!/bin/bash
#
# Stores a SentinelOne API token from the clipboard into the macOS login keychain,
# where run.sh can read it at startup.
#
# Usage:
#   1. Copy the token from the S1 console (use its Copy button -- the field is
#      scrollable, and a click-drag can grab only the visible portion).
#   2. Run this script. It takes no arguments, so the token never appears in your
#      shell history, in `ps` output, or on disk.
#
#      ./store-token.sh
#
# Why not `security add-generic-password -w` with no value? That prompts
# interactively and truncates input at 128 characters, which silently mangles the
# several-hundred-character ES256 JWT that S1 issues. The result authenticates
# with HTTP 401 and looks exactly like an expired token.
#
# Note: Keychain Access.app was removed in macOS 26, so the CLI is the only way
# to write these items.

set -uo pipefail

KEYCHAIN_SERVICE="sentinelone-mcp"
ACCOUNT="api-key"

TOK=$(pbpaste | tr -d '\n\r')

if [[ -z "$TOK" ]]; then
  echo "Clipboard is empty. Copy the token from the S1 console first." >&2
  exit 1
fi

dots=$(printf '%s' "$TOK" | tr -cd '.' | wc -c | tr -d ' ')
echo "Clipboard holds ${#TOK} chars with $dots '.' separators."

# Validate BEFORE writing, so a bad paste can never clobber a working token.
if [[ "$dots" != "2" ]]; then
  cat >&2 <<EOF

REFUSING TO STORE: an ES256 JWT has exactly 2 '.' separators; found $dots.
The clipboard holds a truncated or non-JWT value. Nothing was changed.

Re-copy from the S1 console using its Copy button rather than selecting the text.
EOF
  exit 1
fi

if [[ "$TOK" != eyJ* ]]; then
  cat >&2 <<EOF

REFUSING TO STORE: value does not begin with 'eyJ' (a base64url JWT header).
Nothing was changed.
EOF
  exit 1
fi

security delete-generic-password -s "$KEYCHAIN_SERVICE" -a "$ACCOUNT" >/dev/null 2>&1 || true
security add-generic-password -s "$KEYCHAIN_SERVICE" -a "$ACCOUNT" -w "$TOK"

# Round-trip: prove the write path did not truncate what we handed it.
STORED=$(security find-generic-password -s "$KEYCHAIN_SERVICE" -a "$ACCOUNT" -w 2>/dev/null | tr -d '\n\r')

if [[ "$STORED" != "$TOK" ]]; then
  echo "MISMATCH: wrote ${#TOK} chars, read back ${#STORED}. The write path truncated." >&2
  exit 1
fi

echo "Stored OK: ${#STORED} chars, round-trip verified."
echo
echo "If you have not set the tenant URL yet:"
echo "  security add-generic-password -s $KEYCHAIN_SERVICE -a api-base -w \"https://your-tenant.sentinelone.net\""
echo
echo "Then restart your MCP client to pick up the new credentials."
