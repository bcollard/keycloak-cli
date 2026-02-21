#!/usr/bin/env bash
set -euo pipefail

KC_CONTAINER_NAME="${KC_CONTAINER_NAME:-keycloak-cli}"
KC_SERVER_HOSTNAME="${KC_SERVER_HOSTNAME:-keycloak.kong.runlocal.dev}"

kcadm_exec() {
  local args=("$@")
  if [[ -n "${KC_ADMIN_SECRET_HEADER:-}" ]]; then
    args+=( -h "keycloak-kong=${KC_ADMIN_SECRET_HEADER}" )
  fi
  docker exec -i "$KC_CONTAINER_NAME" ./kcadm.sh "${args[@]}"
}
