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

REALMS_JSON="$(kcadm_exec get realms --fields realm,enabled)"

if printf '%s' "$REALMS_JSON" | jq -e 'length == 0' >/dev/null 2>&1; then
  echo "No realms found."
  exit 0
fi

# Exclude the master realm from the list (Keycloak forbids deleting it)
mapfile -t REALM_LINES < <(printf '%s' "$REALMS_JSON" | \
  jq -r '.[] | select(.realm != "master") | "\(.realm)\(if .enabled then "" else "  [disabled]" end)"')

if [[ ${#REALM_LINES[@]} -eq 0 ]]; then
  echo "No deletable realms found (master realm cannot be deleted)."
  exit 0
fi

SELECTED="$(printf '%s\n' "${REALM_LINES[@]}" | gum choose --no-limit --header "Select realms to delete:")"

if [[ -z "$SELECTED" ]]; then
  echo "No realms selected."
  exit 0
fi

SELECTED_COUNT="$(printf '%s\n' "$SELECTED" | wc -l | tr -d '[:space:]')"
if ! gum confirm "Permanently delete $SELECTED_COUNT realm(s) and ALL their data? This cannot be undone."; then
  echo "Aborted."
  exit 0
fi

mapfile -t SELECTED_LINES <<< "$SELECTED"
for line in "${SELECTED_LINES[@]}"; do
  # Realm name is the first word (strip optional "[disabled]" suffix)
  REALM_NAME="$(printf '%s' "$line" | awk '{print $1}')"
  kcadm_exec delete "realms/$REALM_NAME" </dev/null
  echo "Deleted realm: $REALM_NAME"
done
