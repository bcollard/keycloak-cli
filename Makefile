# Makefile for Keycloak on Cloud Run with Terraform
.ONESHELL:
.DEFAULT_GOAL := help

# Configuration
VERSION ?= 26.5
KC_CONTAINER_NAME ?= keycloak-cli
IMAGE ?= ghcr.io/bcollard/keycloak-cli

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


######################
# Docker commands
######################
.PHONY: docker-build docker-stop docker-run docker-exec
docker-build: ## Build the Docker image for Keycloak
	@docker build -t $(IMAGE):$(VERSION) --build-arg KEYCLOAK_VERSION=$(VERSION) .

docker-stop: ## Stop and remove the Docker container
	@docker stop $(KC_CONTAINER_NAME) || true
	@docker rm $(KC_CONTAINER_NAME) || true

docker-run: docker-build ## Build and run the Keycloak Docker container
	@docker run -d --name $(KC_CONTAINER_NAME) -it $(IMAGE):$(VERSION)

docker-exec: ## Execute a bash shell inside the running Keycloak container
	@docker exec -it $(KC_CONTAINER_NAME) /bin/bash
# ./kcadm.sh config credentials --server https://keycloak.kong.runlocal.dev --realm master --user admin

docker-cleanup: docker-stop ## Clean up Docker container
	@docker rmi $(IMAGE):$(VERSION) || true


######################
# Keycloak Admin
######################
.PHONY: kcadm-help kcadm-login kcadm-get-realms kcadm-create-realm kcadm-new-client-initial-token
kcadm-help: ## Execute a bash shell inside the running Keycloak container
	@docker exec -it $(KC_CONTAINER_NAME) ./kcadm.sh help

kcadm-login: ## Authenticate with Keycloak using credentials from environment variables
	@docker exec -it $(KC_CONTAINER_NAME) ./kcadm.sh config credentials --server ${KC_SERVER} --realm master --user admin --password "${KC_ADMIN_PASSWORD}" --client admin-cli

kcadm-get-realms: ## List all realms in Keycloak, parsed with jq and print only the realm names as a list
	@docker exec -it $(KC_CONTAINER_NAME) ./kcadm.sh get realms --fields id,realm,enabled -h keycloak-kong=${KC_ADMIN_SECRET_HEADER}
#  | jq '.[].realm'

kcadm-create-realm: ## Create a new realm in Keycloak
	@read -p "Enter new realm name: " NEW_REALM_NAME; \
	 docker exec -it $(KC_CONTAINER_NAME) ./kcadm.sh create realms -s realm=$${NEW_REALM_NAME} -s enabled=true -h keycloak-kong=${KC_ADMIN_SECRET_HEADER}

kcadm-new-client-initial-token: kcadm-login ## Generate a new initial access token for a realm
	@read -p "Enter realm name: " REALM_NAME; \
	 docker exec -i $(KC_CONTAINER_NAME) ./kcadm.sh create clients-initial-access -r $${REALM_NAME} -h keycloak-kong=${KC_ADMIN_SECRET_HEADER} -s expiration=3600 -s count=15 -o | jq -r '.token' | tr -d '\r'


######################
# Client registration
######################
.PHONY: kcreg-help kcreg-login kcreg-create-client kcreg-create-client-with-initial-token
kcreg-help: ## Show help for kcreg command
	@docker exec -it $(KC_CONTAINER_NAME) ./kcreg.sh help

kcreg-create-client-prompt-token: ## Create a new client in Keycloak using kcreg (supports INITIAL_TOKEN and REALM_NAME env vars)
	@read -p "Enter realm name: " REALM_NAME; \
	 read -p "Enter initial token: " INITIAL_TOKEN; \
	 REALM_NAME="$${REALM_NAME}" INITIAL_TOKEN="$${INITIAL_TOKEN}" KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) ./scripts/kcreg-create-client.sh

kcreg-create-client-generate-initial-token: ## Generate an initial token and use it for kcreg client creation
	@read -p "Enter realm name: " REALM_NAME; \
	 INITIAL_TOKEN="$$(docker exec -i $(KC_CONTAINER_NAME) ./kcadm.sh create clients-initial-access -r $${REALM_NAME} -h keycloak-kong=${KC_ADMIN_SECRET_HEADER} -s expiration=3600 -s count=15 -o | jq -r '.token' | tr -d '\r')"; \
	 REALM_NAME="$${REALM_NAME}" INITIAL_TOKEN="$${INITIAL_TOKEN}" KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) ./scripts/kcreg-create-client.sh




