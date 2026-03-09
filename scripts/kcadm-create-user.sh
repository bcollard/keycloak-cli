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

USERNAME="$(gum input --prompt "Username: " --placeholder "jdoe")"
if [[ -z "$USERNAME" ]]; then
  echo "Username is required."
  exit 1
fi

EMAIL="$(gum input --prompt "Email (optional): " --placeholder "jdoe@example.com")"
FIRST_NAME="$(gum input --prompt "First name (optional): " --placeholder "Jane")"
LAST_NAME="$(gum input --prompt "Last name (optional): " --placeholder "Doe")"

PASSWORD="$(gum input --password --prompt "Password (leave blank to skip): ")"
TEMPORARY_PASSWORD=false
if [[ -n "$PASSWORD" ]]; then
  if gum confirm "Mark password as temporary (user must change on first login)?"; then
    TEMPORARY_PASSWORD=true
  fi
fi

# Build the user JSON payload with jq (credentials included when a password is set)
CREATE_PAYLOAD="$(jq -n \
  --arg username "$USERNAME" \
  --arg email "$EMAIL" \
  --arg firstName "$FIRST_NAME" \
  --arg lastName "$LAST_NAME" \
  --arg password "$PASSWORD" \
  --argjson temporary "$TEMPORARY_PASSWORD" \
  '{
    username: $username,
    enabled: true,
    email: (if $email != "" then $email else null end),
    firstName: (if $firstName != "" then $firstName else null end),
    lastName: (if $lastName != "" then $lastName else null end),
    credentials: (if $password != "" then [{"type":"password","value":$password,"temporary":$temporary}] else null end)
  } | with_entries(select(.value != null))')"

# Create the user (-i prints the created resource id to stdout)
USER_ID="$(printf '%s' "$CREATE_PAYLOAD" | kcadm_exec create users -r "$REALM_NAME" -f - -i)"
USER_ID="$(printf '%s' "$USER_ID" | tr -d '[:space:]')"

echo "User created with ID: $USER_ID"
if [[ -n "$PASSWORD" ]]; then
  echo "Password set (temporary: $TEMPORARY_PASSWORD)"
fi

if gum confirm "Add user to one or more groups?"; then
  GROUPS_JSON="$(kcadm_exec get groups -r "$REALM_NAME" --fields id,name 2>/dev/null)"
  if [[ -z "$GROUPS_JSON" ]] || [[ "$GROUPS_JSON" == "[ ]" ]] || [[ "$GROUPS_JSON" == "[]" ]]; then
    echo "No groups found in realm '$REALM_NAME'."
  else
    # Build a display list "name  (id)" and let the user pick (multi-select)
    mapfile -t GROUP_LINES < <(printf '%s' "$GROUPS_JSON" | jq -r '.[] | "\(.name)  (\(.id))"')
    SELECTED="$(printf '%s\n' "${GROUP_LINES[@]}" | gum choose --no-limit --header "Select groups to add the user to:")"
    if [[ -n "$SELECTED" ]]; then
      mapfile -t SELECTED_LINES <<< "$SELECTED"
      for line in "${SELECTED_LINES[@]}"; do
        GROUP_ID="$(printf '%s' "$line" | grep -oE '[0-9a-f-]{36}' | tail -1)"
        GROUP_NAME="$(printf '%s' "$line" | sed 's/  (.*//')"
        kcadm_exec update "users/$USER_ID/groups/$GROUP_ID" -r "$REALM_NAME" </dev/null
        echo "Added to group: $GROUP_NAME"
      done
    fi
  fi
fi
