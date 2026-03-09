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

GROUP_NAME="$(gum input --prompt "Group name: " --placeholder "my-group")"
if [[ -z "$GROUP_NAME" ]]; then
  echo "Group name is required."
  exit 1
fi

PARENT_GROUP_ID=""
if gum confirm "Create as a subgroup of an existing group?"; then
  PARENT_GROUP_ID="$(gum input --prompt "Parent group ID: " --placeholder "<UUID>")"
fi

if [[ -n "$PARENT_GROUP_ID" ]]; then
  GROUP_ID="$(kcadm_exec create "groups/$PARENT_GROUP_ID/children" -r "$REALM_NAME" -s "name=$GROUP_NAME" -i)"
else
  GROUP_ID="$(kcadm_exec create groups -r "$REALM_NAME" -s "name=$GROUP_NAME" -i)"
fi

GROUP_ID="$(printf '%s' "$GROUP_ID" | tr -d '[:space:]')"
echo "Group created with ID: $GROUP_ID"
