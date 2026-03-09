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

CLIENTS_JSON="$(kcadm_exec get clients -r "$REALM_NAME" --fields id,clientId,name,enabled </dev/null)"

if printf '%s' "$CLIENTS_JSON" | jq -e 'length == 0' >/dev/null 2>&1; then
  echo "No clients found in realm '$REALM_NAME'."
  exit 0
fi

mapfile -t CLIENT_LINES < <(printf '%s' "$CLIENTS_JSON" | \
  jq -r '.[] | "\(.clientId)\(if .name != null and .name != "" and .name != .clientId then "  (\(.name))" else "" end)\(if .enabled then "" else "  [disabled]" end)  (\(.id))"')

SELECTED="$(printf '%s\n' "${CLIENT_LINES[@]}" | gum choose --no-limit --header "Select clients to delete:")"

if [[ -z "$SELECTED" ]]; then
  echo "No clients selected."
  exit 0
fi

SELECTED_COUNT="$(printf '%s\n' "$SELECTED" | wc -l | tr -d '[:space:]')"
if ! gum confirm "Permanently delete $SELECTED_COUNT client(s)? This cannot be undone."; then
  echo "Aborted."
  exit 0
fi

mapfile -t SELECTED_LINES <<< "$SELECTED"
for line in "${SELECTED_LINES[@]}"; do
  CLIENT_ID="$(printf '%s' "$line" | grep -oE '[0-9a-f-]{36}' | tail -1)"
  CLIENT_NAME="$(printf '%s' "$line" | awk '{print $1}')"
  kcadm_exec delete "clients/$CLIENT_ID" -r "$REALM_NAME" </dev/null
  echo "Deleted client: $CLIENT_NAME"
done
