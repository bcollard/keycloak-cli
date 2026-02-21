#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/container-common.sh"

KC_CONTAINER_NAME="${KC_CONTAINER_NAME:-keycloak-cli}"
KC_SERVER_HOSTNAME="${KC_SERVER_HOSTNAME:-keycloak.kong.runlocal.dev}"

kcadm_exec() {
  ensure_container_running "$KC_CONTAINER_NAME"
  local args=("$@")
  if [[ -n "${KC_ADMIN_SECRET_HEADER:-}" ]]; then
    args+=( -h "keycloak-kong=${KC_ADMIN_SECRET_HEADER}" )
  fi
  docker exec -i "$KC_CONTAINER_NAME" ./kcadm.sh "${args[@]}"
}
