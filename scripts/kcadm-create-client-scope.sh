#!/usr/bin/env bash
set -euo pipefail

if ! command -v gum >/dev/null 2>&1; then
  echo "gum CLI is required. Install it first: brew install gum"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/kcadm-common.sh"

REALM_NAME="${REALM_NAME:-}"
if [[ -z "$REALM_NAME" ]]; then
  REALM_NAME="$(gum input --prompt "Realm name: " --placeholder "master")"
fi

if [[ -z "$REALM_NAME" ]]; then
  echo "Realm name is required."
  exit 1
fi

SCOPE_NAME="$(gum input --prompt "Scope name: " --placeholder "my-scope")"
if [[ -z "$SCOPE_NAME" ]]; then
  echo "Scope name is required."
  exit 1
fi

SCOPE_DESCRIPTION="$(gum input --prompt "Description (optional): " --placeholder "")"

SCOPE_ID="$(kcadm_exec create client-scopes -r "$REALM_NAME" \
  -s "name=$SCOPE_NAME" \
  -s "description=$SCOPE_DESCRIPTION" \
  -s "protocol=openid-connect" \
  -s "type=default" \
  -s 'attributes={"include.in.token.scope":"true"}' \
  -i </dev/null)"

SCOPE_ID="$(printf '%s' "$SCOPE_ID" | tr -d '[:space:]')"
echo "Client scope '$SCOPE_NAME' created with ID: $SCOPE_ID"
