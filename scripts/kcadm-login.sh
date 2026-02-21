#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/kcadm-common.sh"

ensure_container_running "$KC_CONTAINER_NAME"

docker exec -it "$KC_CONTAINER_NAME" ./kcadm.sh config credentials \
  --server "https://${KC_SERVER_HOSTNAME}" \
  --realm master \
  --user admin \
  --password "${KC_ADMIN_PASSWORD}" \
  --client admin-cli
