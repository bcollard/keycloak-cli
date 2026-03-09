#!/usr/bin/env bash
set -euo pipefail

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required. Install it first: brew install jq"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/kcadm-common.sh"

REALM_NAME="${REALM_NAME:-}"
if [[ -z "$REALM_NAME" ]]; then
  read -r -p "Enter realm name: " REALM_NAME
fi

if [[ -z "$REALM_NAME" ]]; then
  echo "Realm name is required."
  exit 1
fi

USERS_JSON="$(kcadm_exec get users -r "$REALM_NAME" --fields id,username,email,firstName,lastName,enabled)"

if printf '%s' "$USERS_JSON" | jq -e 'length == 0' >/dev/null 2>&1; then
  echo "No users found in realm '$REALM_NAME'."
  exit 0
fi

# Enrich each user with their group memberships.
# </dev/null on the inner kcadm_exec calls prevents docker exec -i from
# consuming lines from the while-loop's process substitution pipe.
ENRICHED_USERS=()
while IFS= read -r user; do
  USER_ID="$(printf '%s' "$user" | jq -r '.id')"
  GROUPS_JSON="$(kcadm_exec get "users/$USER_ID/groups" -r "$REALM_NAME" --fields name </dev/null 2>/dev/null || echo '[]')"
  ENRICHED="$(printf '%s' "$user" | jq \
    --argjson groups "$GROUPS_JSON" \
    'del(.id) + {groups: [$groups[].name]}')"
  ENRICHED_USERS+=("$ENRICHED")
done < <(printf '%s' "$USERS_JSON" | jq -c '.[]')

printf '%s\n' "${ENRICHED_USERS[@]}" | jq -s '.'
