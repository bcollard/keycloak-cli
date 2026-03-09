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

GROUPS_JSON="$(kcadm_exec get groups -r "$REALM_NAME" --fields id,name,path,subGroupCount)"

if printf '%s' "$GROUPS_JSON" | jq -e 'length == 0' >/dev/null 2>&1; then
  echo "No groups found in realm '$REALM_NAME'."
  exit 0
fi

# Enrich each group with its members.
# </dev/null on the inner kcadm_exec calls prevents docker exec -i from
# consuming lines from the while-loop's process substitution pipe.
ENRICHED_GROUPS=()
while IFS= read -r group; do
  GROUP_ID="$(printf '%s' "$group" | jq -r '.id')"
  MEMBERS_JSON="$(kcadm_exec get "groups/$GROUP_ID/members" -r "$REALM_NAME" --fields username </dev/null 2>/dev/null || echo '[]')"
  ENRICHED="$(printf '%s' "$group" | jq \
    --argjson members "$MEMBERS_JSON" \
    'del(.id) + {members: [$members[].username]}')"
  ENRICHED_GROUPS+=("$ENRICHED")
done < <(printf '%s' "$GROUPS_JSON" | jq -c '.[]')

printf '%s\n' "${ENRICHED_GROUPS[@]}" | jq -s '.'
