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

# --- Select users ---
USERS_JSON="$(kcadm_exec get users -r "$REALM_NAME" --fields id,username,email </dev/null)"

if printf '%s' "$USERS_JSON" | jq -e 'length == 0' >/dev/null 2>&1; then
  echo "No users found in realm '$REALM_NAME'."
  exit 0
fi

mapfile -t USER_LINES < <(printf '%s' "$USERS_JSON" | \
  jq -r '.[] | "\(.username)\(if .email != null and .email != "" then "  <\(.email)>" else "" end)  (\(.id))"')

SELECTED_USERS="$(printf '%s\n' "${USER_LINES[@]}" | gum choose --no-limit --header "Select users to add to a group:")"

if [[ -z "$SELECTED_USERS" ]]; then
  echo "No users selected."
  exit 0
fi

# --- Select groups ---
GROUPS_JSON="$(kcadm_exec get groups -r "$REALM_NAME" --fields id,name,path </dev/null)"

if printf '%s' "$GROUPS_JSON" | jq -e 'length == 0' >/dev/null 2>&1; then
  echo "No groups found in realm '$REALM_NAME'."
  exit 0
fi

mapfile -t GROUP_LINES < <(printf '%s' "$GROUPS_JSON" | \
  jq -r '.[] | "\(.name)  (\(.path))  (\(.id))"')

SELECTED_GROUPS="$(printf '%s\n' "${GROUP_LINES[@]}" | gum choose --no-limit --header "Select groups to add the users to:")"

if [[ -z "$SELECTED_GROUPS" ]]; then
  echo "No groups selected."
  exit 0
fi

# --- Assign: each user × each group ---
# Use mapfile + for loops to avoid docker exec -i consuming herestring stdin
mapfile -t SELECTED_USER_LINES <<< "$SELECTED_USERS"
mapfile -t SELECTED_GROUP_LINES <<< "$SELECTED_GROUPS"

for user_line in "${SELECTED_USER_LINES[@]}"; do
  USER_ID="$(printf '%s' "$user_line" | grep -oE '[0-9a-f-]{36}' | tail -1)"
  USERNAME="$(printf '%s' "$user_line" | awk '{print $1}')"
  for group_line in "${SELECTED_GROUP_LINES[@]}"; do
    GROUP_ID="$(printf '%s' "$group_line" | grep -oE '[0-9a-f-]{36}' | tail -1)"
    GROUP_NAME="$(printf '%s' "$group_line" | sed 's/  (.*//')"
    kcadm_exec update "users/$USER_ID/groups/$GROUP_ID" -r "$REALM_NAME" </dev/null
    echo "Added $USERNAME to group: $GROUP_NAME"
  done
done
