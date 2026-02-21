#!/usr/bin/env bash
set -euo pipefail

KC_CONTAINER_NAME="${KC_CONTAINER_NAME:-keycloak-cli}"
KC_SERVER_HOSTNAME="${KC_SERVER_HOSTNAME:-keycloak.kong.runlocal.dev}"

docker exec -it "$KC_CONTAINER_NAME" ./kcadm.sh config credentials \
  --server "https://${KC_SERVER_HOSTNAME}" \
  --realm master \
  --user admin \
  --password "${KC_ADMIN_PASSWORD}" \
  --client admin-cli
