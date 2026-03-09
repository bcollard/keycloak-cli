#!/usr/bin/env bash
set -euo pipefail

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

kcadm_exec get clients -r "$REALM_NAME" --fields id,clientId,name,description,enabled,publicClient,serviceAccountsEnabled </dev/null
