#!/usr/bin/env bash
set -euo pipefail

if ! command -v gum >/dev/null 2>&1; then
  echo "gum CLI is required. Install it first: brew install gum"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required. Install it first: brew install jq"
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

# Select client
CLIENTS_JSON="$(kcadm_exec get clients -r "$REALM_NAME" --fields id,clientId,name,enabled </dev/null)"

if printf '%s' "$CLIENTS_JSON" | jq -e 'length == 0' >/dev/null 2>&1; then
  echo "No clients found in realm '$REALM_NAME'."
  exit 0
fi

mapfile -t CLIENT_LINES < <(printf '%s' "$CLIENTS_JSON" | \
  jq -r '.[] | "\(.clientId)\(if .name != null and .name != "" and .name != .clientId then "  (\(.name))" else "" end)\(if .enabled then "" else "  [disabled]" end)  (\(.id))"')

SELECTED_CLIENT="$(printf '%s\n' "${CLIENT_LINES[@]}" | gum choose --header "Select a client:")"

if [[ -z "$SELECTED_CLIENT" ]]; then
  echo "No client selected."
  exit 0
fi

CLIENT_UUID="$(printf '%s' "$SELECTED_CLIENT" | grep -oE '[0-9a-f-]{36}' | tail -1)"
CLIENT_NAME="$(printf '%s' "$SELECTED_CLIENT" | awk '{print $1}')"

# Select scope type
SCOPE_TYPE="$(printf 'default\noptional' | gum choose --header "Scope type:")"

if [[ -z "$SCOPE_TYPE" ]]; then
  echo "No scope type selected."
  exit 0
fi

# List available client scopes
SCOPES_JSON="$(kcadm_exec get client-scopes -r "$REALM_NAME" --fields id,name,description,protocol </dev/null)"

if printf '%s' "$SCOPES_JSON" | jq -e 'length == 0' >/dev/null 2>&1; then
  echo "No client scopes found in realm '$REALM_NAME'."
  exit 0
fi

mapfile -t SCOPE_LINES < <(printf '%s' "$SCOPES_JSON" | \
  jq -r '.[] | "\(.name)\(if .description != null and .description != "" then "  (\(.description))" else "" end)  [\(.protocol)]  (\(.id))"')

SELECTED_SCOPE="$(printf '%s\n' "${SCOPE_LINES[@]}" | gum choose --header "Select a scope to add as $SCOPE_TYPE:")"

if [[ -z "$SELECTED_SCOPE" ]]; then
  echo "No scope selected."
  exit 0
fi

SCOPE_UUID="$(printf '%s' "$SELECTED_SCOPE" | grep -oE '[0-9a-f-]{36}' | tail -1)"
SCOPE_NAME="$(printf '%s' "$SELECTED_SCOPE" | awk '{print $1}')"

# Add scope to client
kcadm_exec update "clients/$CLIENT_UUID/${SCOPE_TYPE}-client-scopes/$SCOPE_UUID" -r "$REALM_NAME" </dev/null
echo "Added $SCOPE_TYPE scope '$SCOPE_NAME' to client '$CLIENT_NAME'."
