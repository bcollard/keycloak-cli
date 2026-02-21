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

KC_SERVER_HOSTNAME="${KC_SERVER_HOSTNAME:-}"
KC_CONTAINER_NAME="${KC_CONTAINER_NAME:-keycloak-cli}"
INITIAL_TOKEN="${INITIAL_TOKEN:-}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/container-common.sh"

ensure_container_running "$KC_CONTAINER_NAME"

REALM_NAME="${REALM_NAME:-}"
if [[ -z "$REALM_NAME" ]]; then
  REALM_NAME="$(gum input --prompt "Realm name: " --placeholder "master")"
fi

if [[ -z "$INITIAL_TOKEN" ]]; then
  INITIAL_TOKEN_RAW="$(gum write --placeholder "Paste initial token (JWT) then press Ctrl+D")"
  INITIAL_TOKEN="$(printf '%s' "$INITIAL_TOKEN_RAW" | tr -d '[:space:]')"
fi

if [[ -z "$INITIAL_TOKEN" ]]; then
  echo "Initial token is required."
  echo "Generate one with: make new-client-initial-token"
  exit 1
fi

CLIENT_ID="$(gum input --prompt "Client ID: " --placeholder "my_client")"
CLIENT_NAME="$(gum input --prompt "Client name: " --value "$CLIENT_ID")"

FLOW_STANDARD="Authorization Code Flow (standardFlowEnabled)"
FLOW_IMPLICIT="Implicit Flow (implicitFlowEnabled)"
FLOW_DIRECT_ACCESS="Resource Owner Password Credentials Grant (directAccessGrantsEnabled)"
FLOW_SERVICE_ACCOUNTS="Client Credentials Grant (serviceAccountsEnabled)"
FLOW_DEVICE_AUTH="Device Authorization Grant (oauth2.device.authorization.grant.enabled)"
FLOW_JWT_AUTH="JWT Authorization Grant (oauth2.jwt.authorization.grant.enabled)"
FLOW_CIBA="CIBA Grant (oidc.ciba.grant.enabled)"
FLOW_TOKEN_EXCHANGE="Token Exchange (standard.token.exchange.enabled)"

SELECTED_FLOWS="$(gum choose --no-limit \
  "$FLOW_STANDARD" \
  "$FLOW_IMPLICIT" \
  "$FLOW_DIRECT_ACCESS" \
  "$FLOW_SERVICE_ACCOUNTS" \
  "$FLOW_DEVICE_AUTH" \
  "$FLOW_JWT_AUTH" \
  "$FLOW_CIBA" \
  "$FLOW_TOKEN_EXCHANGE")"

has_selection() {
  local selection_name="$1"
  if printf '%s\n' "$SELECTED_FLOWS" | grep -Fxq "$selection_name"; then
    echo true
  else
    echo false
  fi
}

STANDARD_FLOW_ENABLED="$(has_selection "$FLOW_STANDARD")"
IMPLICIT_FLOW_ENABLED="$(has_selection "$FLOW_IMPLICIT")"
DIRECT_ACCESS_GRANTS_ENABLED="$(has_selection "$FLOW_DIRECT_ACCESS")"
SERVICE_ACCOUNTS_ENABLED="$(has_selection "$FLOW_SERVICE_ACCOUNTS")"

OAUTH2_DEVICE_AUTHORIZATION_GRANT_ENABLED="$(has_selection "$FLOW_DEVICE_AUTH")"
OAUTH2_JWT_AUTHORIZATION_GRANT_ENABLED="$(has_selection "$FLOW_JWT_AUTH")"
OIDC_CIBA_GRANT_ENABLED="$(has_selection "$FLOW_CIBA")"
STANDARD_TOKEN_EXCHANGE_ENABLED="$(has_selection "$FLOW_TOKEN_EXCHANGE")"

REDIRECT_URIS_JSON='[]'
if [[ "$STANDARD_FLOW_ENABLED" == "true" ]]; then
  REDIRECT_URIS_RAW="$(gum write --value "https://example.com/*" --placeholder "Enter redirect URIs (one per line or comma-separated). Press Ctrl+D when done.")"
  REDIRECT_URIS_JSON="$(printf '%s' "$REDIRECT_URIS_RAW" \
    | tr ',' '\n' \
    | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' \
    | awk 'NF' \
    | jq -Rsc 'split("\n") | map(select(length > 0))')"
fi

OAUTH2_JWT_AUTHORIZATION_GRANT_IDP=""
if [[ "$OAUTH2_JWT_AUTHORIZATION_GRANT_ENABLED" == "true" ]]; then
  OAUTH2_JWT_AUTHORIZATION_GRANT_IDP="$(gum input --prompt "JWT Authorization Grant IdP (optional): " --value "")"
fi

CREATE_PAYLOAD="$(jq -n \
  --arg clientId "$CLIENT_ID" \
  --arg clientName "$CLIENT_NAME" \
  --argjson redirectUris "$REDIRECT_URIS_JSON" \
  --argjson standardFlowEnabled "$STANDARD_FLOW_ENABLED" \
  --argjson implicitFlowEnabled "$IMPLICIT_FLOW_ENABLED" \
  --argjson directAccessGrantsEnabled "$DIRECT_ACCESS_GRANTS_ENABLED" \
  --argjson serviceAccountsEnabled "$SERVICE_ACCOUNTS_ENABLED" \
  --arg oauth2DeviceAuthorizationGrantEnabled "$OAUTH2_DEVICE_AUTHORIZATION_GRANT_ENABLED" \
  --arg oauth2JwtAuthorizationGrantEnabled "$OAUTH2_JWT_AUTHORIZATION_GRANT_ENABLED" \
  --arg oauth2JwtAuthorizationGrantIdp "$OAUTH2_JWT_AUTHORIZATION_GRANT_IDP" \
  --arg oidcCibaGrantEnabled "$OIDC_CIBA_GRANT_ENABLED" \
  --arg standardTokenExchangeEnabled "$STANDARD_TOKEN_EXCHANGE_ENABLED" \
  '{
    clientId: $clientId,
    name: $clientName,
    redirectUris: $redirectUris,
    standardFlowEnabled: $standardFlowEnabled,
    implicitFlowEnabled: $implicitFlowEnabled,
    directAccessGrantsEnabled: $directAccessGrantsEnabled,
    serviceAccountsEnabled: $serviceAccountsEnabled,
    attributes: {
      "oauth2.device.authorization.grant.enabled": $oauth2DeviceAuthorizationGrantEnabled,
      "oauth2.jwt.authorization.grant.enabled": $oauth2JwtAuthorizationGrantEnabled,
      "oauth2.jwt.authorization.grant.idp": $oauth2JwtAuthorizationGrantIdp,
      "oidc.ciba.grant.enabled": $oidcCibaGrantEnabled,
      "standard.token.exchange.enabled": $standardTokenExchangeEnabled
    }
  }')"

KCREG_CREATE_CMD=(docker exec -i "$KC_CONTAINER_NAME" ./kcreg.sh create --realm "$REALM_NAME" --server "https://$KC_SERVER_HOSTNAME" -o)
if [[ -n "$INITIAL_TOKEN" ]]; then
  KCREG_CREATE_CMD+=(-t "$INITIAL_TOKEN")
fi
KCREG_CREATE_CMD+=(-f -)

CREATE_RESPONSE="$(printf '%s' "$CREATE_PAYLOAD" | "${KCREG_CREATE_CMD[@]}")"

echo "$CREATE_RESPONSE" | jq .

CLIENT_SECRET="$(echo "$CREATE_RESPONSE" | jq -r '.secret // empty')"
if [[ -n "$CLIENT_SECRET" ]]; then
  echo
  echo "Client secret: $CLIENT_SECRET"
else
  echo
  echo "Client created, but no secret was returned in the response."
fi

