#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/kcadm-common.sh"

read -r -p "Enter realm name: " REALM_NAME

kcadm_exec create clients-initial-access \
  -r "$REALM_NAME" \
  -s expiration=3600 \
  -s count=15 \
  -o | jq -r '.token' | tr -d '\r'
