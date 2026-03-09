# keycloak-cli

Small helper project to run a Keycloak CLI container and manage realms, clients, users and groups with `kcadm.sh` and `kcreg.sh` through `make` targets.

Demo:

[![asciicast](https://asciinema.org/a/791850.svg)](https://asciinema.org/a/791850)


## Prerequisites

- Docker
- GNU Make
- `jq`
- `gum` (for interactive prompts and multi-select UI)

Install missing CLI tools on macOS:

```bash
brew install jq gum
```

## Project layout

- `Dockerfile`: builds a Keycloak image with preview features enabled
- `Makefile`: entrypoint for build/run/admin/registration commands
- `scripts/kcadm-common.sh`: shared `kcadm` helper (conditionally appends admin header)
- `scripts/container-common.sh`: shared Docker container check helper
- `scripts/kcadm-login.sh`: authenticate with Keycloak
- `scripts/kcadm-get-realms.sh`: list realms
- `scripts/kcadm-create-realm.sh`: create a realm
- `scripts/kcadm-delete-realm.sh`: interactive multi-select realm deletion
- `scripts/kcadm-get-users.sh`: list users with group memberships
- `scripts/kcadm-get-groups.sh`: list groups with members
- `scripts/kcadm-create-user.sh`: interactive user creation helper
- `scripts/kcadm-create-group.sh`: interactive group creation helper
- `scripts/kcadm-add-user-to-group.sh`: interactively assign existing users to existing groups
- `scripts/kcadm-delete-user.sh`: interactive multi-select user deletion
- `scripts/kcadm-delete-group.sh`: interactive multi-select group deletion
- `scripts/kcadm-new-client-initial-token.sh`: generate a client initial access token
- `scripts/kcadm-get-clients.sh`: list clients in a realm
- `scripts/kcadm-delete-client.sh`: interactive multi-select client deletion
- `scripts/kcreg-create-client.sh`: interactive client creation helper

## Configuration

The following variables are used by targets:

- `VERSION` (default: `26.5`)
- `IMAGE` (default: `ghcr.io/bcollard/keycloak-cli`)
- `KC_CONTAINER_NAME` (default: `keycloak-cli`)
- `KC_SERVER_HOSTNAME` (default: `keycloak.kong.runlocal.dev`)
- `KC_ADMIN_PASSWORD` (required for `login`) - password for the `admin` user in the `master` realm, used to obtain an access token for admin operations
- `KC_ADMIN_SECRET_HEADER` (used by admin endpoints) - extra header value required by my LB to protect KC admin endpoints

Example:

```bash
export KC_SERVER_HOSTNAME="keycloak.kong.runlocal.dev" # the domain name your Keycloak instance is accessible at
export KC_ADMIN_PASSWORD="<admin-password>"            # password for the `admin` user in the `master` realm
# optional
export KC_ADMIN_SECRET_HEADER="<secret-header-value>"
```

When `KC_ADMIN_SECRET_HEADER` is set, admin scripts append `-h keycloak-kong=<value>` to `kcadm.sh` calls.
When it is unset/empty, no extra header is added.

## Quick start

Build and run container:

```bash
make docker-run
```

Show all commands:

```bash
make help
```

## Key targets

### Docker

```bash
make docker-build
make docker-run
make docker-stop
make docker-cleanup
```

### Realm admin (`kcadm`)

```bash
make login
make get-realms
make create-realm
make delete-realm
```

`delete-realm` presents a multi-select list of all realms (the `master` realm is excluded). Requires confirmation before proceeding.

### User and group management (`kcadm`)

All targets below support a `REALM_NAME` environment variable to skip the realm prompt.

```bash
make get-users
make get-groups
make create-user
make create-group
make add-user-to-group
make delete-user
make delete-group
```

`get-users` lists all users enriched with their group memberships (username, email, first/last name, enabled, groups).

`get-groups` lists all top-level groups enriched with their members (name, path, subGroupCount, members).

`create-user` prompts for realm, username, optional email/first name/last name, optional password (with temporary flag), and optional group assignment via multi-select.

`create-group` prompts for realm and group name, and optionally creates the group as a subgroup of an existing group by parent group ID.

`add-user-to-group` shows a multi-select list of users then a multi-select list of groups, and adds every selected user to every selected group.

`delete-user` shows a multi-select list of users and deletes the chosen ones after confirmation.

`delete-group` shows a multi-select list of groups and deletes the chosen ones after confirmation.

```bash
REALM_NAME="myrealm" make get-users
REALM_NAME="myrealm" make get-groups
REALM_NAME="myrealm" make create-user
REALM_NAME="myrealm" make create-group
REALM_NAME="myrealm" make add-user-to-group
REALM_NAME="myrealm" make delete-user
REALM_NAME="myrealm" make delete-group
```

### Client management (`kcadm` + `kcreg`)

```bash
make get-clients
make new-client-initial-token
make create-client
make delete-client
```

`get-clients` lists all clients in a realm (clientId, name, description, enabled, publicClient, serviceAccountsEnabled).

`new-client-initial-token` generates a client initial access token and prints the raw token value. Requires a realm name prompt.

`create-client` prompts for realm name and initial token, then launches the interactive `gum` flow picker (see below).

`delete-client` shows a multi-select list of clients and deletes the chosen ones after confirmation.

All four targets support `REALM_NAME` (and `INITIAL_TOKEN` for `create-client`):

```bash
REALM_NAME="myrealm" make get-clients
REALM_NAME="myrealm" make delete-client
REALM_NAME="myrealm" INITIAL_TOKEN="<token>" make create-client
```

## Interactive client creation behavior

The `create-client` script prompts for:

- Realm (unless provided by env)
- Initial access token (unless provided by env)
- Client ID / name
- Enabled OAuth/OIDC flows (multi-select)
- Redirect URIs (only when Authorization Code Flow is enabled; accepts comma-separated and/or multiline values)
- JWT Authorization Grant IdP (only when JWT grant is enabled)

Then it:

- Builds JSON payload safely with `jq`
- Calls `./kcreg.sh create`
- Prints the full create response
- Prints the client secret if returned

## Advanced usage

Run scripts directly with env vars:

```bash
REALM_NAME="myrealm" INITIAL_TOKEN="<token>" KC_CONTAINER_NAME="keycloak-cli" ./scripts/kcreg-create-client.sh
REALM_NAME="myrealm" KC_CONTAINER_NAME="keycloak-cli" ./scripts/kcadm-get-users.sh
```
