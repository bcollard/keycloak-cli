# Makefile for Keycloak on Cloud Run with Terraform
.ONESHELL:
.DEFAULT_GOAL := help

# Configuration
VERSION ?= 26.5

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


######################
# Docker commands
######################
.PHONY: docker-build docker-stop docker-run docker-exec
docker-build: ## Build the Docker image for Keycloak
	@docker build -t ghcr.io/bcollard/keycloak-cli:$(VERSION) --build-arg KEYCLOAK_VERSION=$(VERSION) .

docker-stop: ## Stop and remove the Docker container
	@docker stop keycloak-cli || true
	@docker rm keycloak-cli || true

docker-run: ## Run the Keycloak Docker container
	@docker run -d --name keycloak-cli -it ghcr.io/bcollard/keycloak-cli:$(VERSION)

docker-exec: ## Execute a bash shell inside the running Keycloak container
	@docker exec -it keycloak-cli /bin/bash
# ./kcadm.sh config credentials --server https://keycloak.kong.runlocal.dev --realm master --user admin


######################
# Keycloak Admin
######################
.PHONY: kcadm-help kcadm-login kcadm-get-realms kcadm-create-realm kcadm-new-client-initial-token
kcadm-help: ## Execute a bash shell inside the running Keycloak container
	@docker exec -it keycloak-cli ./kcadm.sh help

kcadm-login: ## Authenticate with Keycloak using credentials from environment variables
	@docker exec -it keycloak-cli ./kcadm.sh config credentials --server ${KC_SERVER} --realm master --user admin --password "${KC_ADMIN_PASSWORD}" --client admin-cli

kcadm-get-realms: ## List all realms in Keycloak, parsed with jq and print only the realm names as a list
	@docker exec -it keycloak-cli ./kcadm.sh get realms --fields id,realm,enabled -h keycloak-kong=${KC_ADMIN_SECRET_HEADER}
#  | jq '.[].realm'

kcadm-create-realm: ## Create a new realm in Keycloak
	@read -p "Enter new realm name: " NEW_REALM_NAME; \
	 docker exec -it keycloak-cli ./kcadm.sh create realms -s realm=$${NEW_REALM_NAME} -s enabled=true -h keycloak-kong=${KC_ADMIN_SECRET_HEADER}

kcadm-new-client-initial-token: ## Generate a new initial access token for a realm
	@read -p "Enter realm name: " REALM_NAME; \
	 docker exec -it keycloak-cli ./kcadm.sh create clients-initial-access -r $${REALM_NAME} -h keycloak-kong=${KC_ADMIN_SECRET_HEADER} -s expiration=3600 -s count=15 -o | jq '.token'


######################
# Client registration
######################
.PHONY: kcreg-help kcreg-login kcreg-create-client
kcreg-help: ## Show help for kcreg command
	@docker exec -it keycloak-cli ./kcreg.sh help

kcreg-login: ## Login to Keycloak Client Registration Service
	@read -p "Enter realm name: " REALM_NAME; \
	 docker exec -it keycloak-cli ./kcreg.sh config credentials --server ${KC_SERVER} --realm $${REALM_NAME}

kcreg-create-client: ## Create a new client in Keycloak using kcreg
	@read -p "Enter realm name: " REALM_NAME; \
	 read -p "Enter client ID: " CLIENT_ID; \
	 docker exec -it keycloak-cli ./kcreg.sh create -s clientId=$${CLIENT_ID} --realm $${REALM_NAME} -s 'redirectUris=["https://example.com/*"]' -t -




