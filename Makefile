# Makefile for Keycloak on Cloud Run with Terraform
.ONESHELL:
.DEFAULT_GOAL := help

# Configuration
VERSION ?= 26.5
KC_CONTAINER_NAME ?= keycloak-cli
IMAGE ?= ghcr.io/bcollard/keycloak-cli
KC_SERVER_HOSTNAME ?= keycloak.kong.runlocal.dev

.PHONY: help
help: ## Show this help message
	@BOLD="\033[1m"; \
	 DIM="\033[2m"; \
	 RESET="\033[0m"; \
	 PURPLE="\033[35m"; \
	 CYAN="\033[36m"; \
	 GREEN="\033[32m"; \
	 YELLOW="\033[33m"; \
	 BLUE="\033[34m"; \
	 printf "$$BOLD$$PURPLE\n== keycloak-cli Make targets ==\n$$RESET"; \
	 printf "$$DIM----------------------------------------------------------------------$$RESET\n"; \
	 printf "$$BOLD$$YELLOW Lifecycle$$RESET $$DIM->$$RESET $$GREEN docker-run $$DIM/$$RESET$$GREEN docker-cleanup$$RESET\n"; \
	 printf "$$BOLD$$YELLOW Realm Admin$$RESET $$DIM->$$RESET $$GREEN login $$DIM then$$RESET $$GREEN get-realms $$DIM/$$RESET$$GREEN create-realm $$DIM then$$RESET $$GREEN new-client-initial-token$$RESET\n"; \
	 printf "$$BOLD$$YELLOW Client Reg$$RESET $$DIM->$$RESET $$GREEN create-client$$RESET\n\n"; \
	 printf "$$BOLD$$BLUE Available targets$$RESET\n"; \
	 grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[1;36m%-36s\033[0m \033[2m%s\033[0m\n", $$1, $$2}'


######################
# Docker commands
######################
.PHONY: docker-build docker-stop docker-run docker-exec docker-cleanup
docker-build: ## Build the Docker image for Keycloak
	@docker build -t $(IMAGE):$(VERSION) --build-arg KEYCLOAK_VERSION=$(VERSION) --build-arg KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) .

docker-stop: ## Stop and remove the Docker container
	@docker stop $(KC_CONTAINER_NAME) || true
	@docker rm $(KC_CONTAINER_NAME) || true

docker-run: docker-build ## Build and run the Keycloak Docker container
	@docker run -d --name $(KC_CONTAINER_NAME) -it $(IMAGE):$(VERSION)

docker-exec: ## Execute a bash shell inside the running Keycloak container
	@docker exec -it $(KC_CONTAINER_NAME) /bin/bash

docker-cleanup: docker-stop ## Clean up Docker container
	@docker rmi $(IMAGE):$(VERSION) || true


######################
# Keycloak Admin
######################
.PHONY: login get-realms create-realm new-client-initial-token
login: ## Authenticate with Keycloak using admin credentials from environment variables
	@docker exec -it $(KC_CONTAINER_NAME) ./kcadm.sh config credentials --server https://$(KC_SERVER_HOSTNAME) --realm master --user admin --password "${KC_ADMIN_PASSWORD}" --client admin-cli

get-realms: ## List all realms in Keycloak, parsed with jq and print only the realm names as a list
	@docker exec -it $(KC_CONTAINER_NAME) ./kcadm.sh get realms --fields id,realm,enabled -h keycloak-kong=${KC_ADMIN_SECRET_HEADER}

create-realm: ## Create a new realm in Keycloak
	@read -p "Enter new realm name: " NEW_REALM_NAME; \
	 docker exec -it $(KC_CONTAINER_NAME) ./kcadm.sh create realms -s realm=$${NEW_REALM_NAME} -s enabled=true -h keycloak-kong=${KC_ADMIN_SECRET_HEADER}

new-client-initial-token: login ## Generate a new initial access token for a realm
	@read -p "Enter realm name: " REALM_NAME; \
	 docker exec -i $(KC_CONTAINER_NAME) ./kcadm.sh create clients-initial-access -r $${REALM_NAME} -h keycloak-kong=${KC_ADMIN_SECRET_HEADER} -s expiration=3600 -s count=15 -o | jq -r '.token' | tr -d '\r'


######################
# Client registration
######################
.PHONY: create-client

create-client: ## Create a new client in Keycloak using kcreg (supports INITIAL_TOKEN and REALM_NAME env vars)
	@read -p "Enter realm name: " REALM_NAME; \
	 read -p "Enter initial token: " INITIAL_TOKEN; \
	 REALM_NAME="$${REALM_NAME}" INITIAL_TOKEN="$${INITIAL_TOKEN}" KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) ./scripts/kcreg-create-client.sh

