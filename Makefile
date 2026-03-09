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
	 printf "$$BOLD$$YELLOW Realm Admin$$RESET $$DIM->$$RESET $$GREEN login $$DIM then$$RESET $$GREEN get-realms $$DIM/$$RESET$$GREEN create-realm $$DIM/$$RESET$$GREEN delete-realm$$RESET\n"; \
	 printf "$$BOLD$$YELLOW User/Group Admin$$RESET $$DIM->$$RESET $$GREEN get-users $$DIM/$$RESET$$GREEN get-groups $$DIM/$$RESET$$GREEN create-user $$DIM/$$RESET$$GREEN create-group $$DIM/$$RESET$$GREEN add-user-to-group $$DIM/$$RESET$$GREEN delete-user $$DIM/$$RESET$$GREEN delete-group$$RESET\n"; \
	 printf "$$BOLD$$YELLOW Client Reg$$RESET $$DIM->$$RESET $$GREEN get-clients $$DIM/$$RESET$$GREEN new-client-initial-token $$DIM then$$RESET $$GREEN create-client $$DIM/$$RESET$$GREEN delete-client$$RESET\n\n"; \
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
.PHONY: login get-realms create-realm delete-realm
login: ## Authenticate with Keycloak using admin credentials from environment variables
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_PASSWORD="${KC_ADMIN_PASSWORD}" ./scripts/kcadm-login.sh

get-realms: login ## List all realms in Keycloak, parsed with jq and print only the realm names as a list
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-get-realms.sh

create-realm: login ## Create a new realm in Keycloak
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-create-realm.sh

delete-realm: login ## Delete one or more realms (interactive; master realm is excluded)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-delete-realm.sh


######################
# User / Group admin
######################
.PHONY: get-users get-groups create-user create-group add-user-to-group delete-user delete-group
get-users: login ## List users in a realm (supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-get-users.sh

get-groups: login ## List groups in a realm (supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-get-groups.sh

create-user: login ## Create a new user in a realm (interactive; supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-create-user.sh

create-group: login ## Create a new group (or subgroup) in a realm (interactive; supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-create-group.sh

add-user-to-group: login ## Add existing user(s) to existing group(s) via interactive multi-select (supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-add-user-to-group.sh

delete-user: login ## Delete one or more users from a realm (interactive; supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-delete-user.sh

delete-group: login ## Delete one or more groups from a realm (interactive; supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-delete-group.sh


######################
# Client registration
######################
.PHONY: get-clients new-client-initial-token create-client delete-client
get-clients: login ## List clients in a realm (supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-get-clients.sh

new-client-initial-token: login ## Generate a new initial access token for a realm
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-new-client-initial-token.sh

create-client: ## Create a new client in Keycloak using kcreg (supports INITIAL_TOKEN and REALM_NAME env vars)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) ./scripts/kcreg-create-client.sh

delete-client: login ## Delete one or more clients from a realm (interactive; supports REALM_NAME env var)
	@KC_CONTAINER_NAME=$(KC_CONTAINER_NAME) KC_SERVER_HOSTNAME=$(KC_SERVER_HOSTNAME) KC_ADMIN_SECRET_HEADER="${KC_ADMIN_SECRET_HEADER}" ./scripts/kcadm-delete-client.sh

