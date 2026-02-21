#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/kcadm-common.sh"

read -r -p "Enter new realm name: " NEW_REALM_NAME

kcadm_exec create realms -s "realm=${NEW_REALM_NAME}" -s enabled=true
