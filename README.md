# keycloak-cli

A portable Go CLI (`kc`) to manage Keycloak realms, users, groups and clients.
No Docker. No Java. A single static binary.

Demo:

[![asciicast](https://asciinema.org/a/791850.svg)](https://asciinema.org/a/791850)


## Install

### Homebrew (macOS / Linux)

```bash
brew tap bcollard/keycloak-cli
brew install --cask keycloak-cli
```

### go install

```bash
go install github.com/bcollard/keycloak-cli@latest
```

### Download binary

Grab the latest release from [GitHub Releases](https://github.com/bcollard/keycloak-cli/releases),
extract the archive and move `kc` to somewhere on your `PATH`.

### Build from source

```bash
git clone https://github.com/bcollard/keycloak-cli
cd keycloak-cli
make install        # builds ./kc and copies to /usr/local/bin
```


## Prerequisites

- `kc` binary (see Install above)
- A running Keycloak instance reachable over HTTPS


## Quick start

```bash
kc login
# → prompts for server URL, admin username, password
# → stores token in ~/.config/kc/token.json

kc realm list
kc realm create
kc client list
```


## Configuration

The following environment variables are supported:

| Variable | Description |
|---|---|
| `KC_SERVER` | Server URL (e.g. `https://keycloak.example.com`), skips the login prompt |
| `KC_ADMIN_USER` | Admin username (default: `admin`), skips the login prompt |
| `KC_ADMIN_PASSWORD` | Admin password, skips the login prompt |
| `KC_ADMIN_SECRET_HEADER` | Optional extra header value sent as `keycloak-kong: <value>` on every admin request |
| `REALM_NAME` | Pre-select realm for any command that asks for one |

Example `.envrc` (direnv):

```bash
export KC_SERVER="https://keycloak.kong.runlocal.dev"
export KC_ADMIN_PASSWORD="<admin-password>"
# optional
export KC_ADMIN_SECRET_HEADER="<secret-header-value>"
export REALM_NAME="my-realm"
```

Token and config are stored in `~/.config/kc/` (mode `0600`).
The token is refreshed automatically before it expires.


## Commands

### `kc login`

Authenticate with Keycloak. Stores the access + refresh token locally.

```bash
kc login
KC_SERVER=https://kc.example.com KC_ADMIN_PASSWORD=secret kc login
```

### Realm

```bash
kc realm list
kc realm create
kc realm delete     # interactive multi-select + confirmation; master realm excluded
```

### User

All user commands support `REALM_NAME`.

```bash
kc user list        # lists users enriched with their group memberships
kc user create      # interactive: username, email, name, password, group assignment
kc user delete      # interactive multi-select + confirmation
```

```bash
REALM_NAME=myrealm kc user list
REALM_NAME=myrealm kc user create
REALM_NAME=myrealm kc user delete
```

### Group

```bash
kc group list       # lists groups enriched with their members
kc group create     # interactive: name, optional parent group
kc group delete     # interactive multi-select + confirmation
kc group add-member # multi-select users × groups (Cartesian assignment)
```

```bash
REALM_NAME=myrealm kc group list
REALM_NAME=myrealm kc group add-member
```

### Client

```bash
kc client list
kc client create    # interactive: client ID, name, OAuth2 flows, redirect URIs
kc client delete    # interactive multi-select + confirmation
```

```bash
REALM_NAME=myrealm kc client list
REALM_NAME=myrealm kc client create
REALM_NAME=myrealm kc client delete
```

`kc client create` flow picker supports:
- Authorization Code Flow
- Implicit Flow
- Resource Owner Password Credentials
- Client Credentials Grant
- Device Authorization Grant
- Token Exchange

### Client scope

```bash
kc client-scope create   # name + description; type=default, include.in.token.scope=true
kc client-scope add      # select client → default/optional → select scope
```

```bash
REALM_NAME=myrealm kc client-scope create
REALM_NAME=myrealm kc client-scope add
```


## Distribution

Releases are built with [goreleaser](https://goreleaser.com/) and published to GitHub Releases
via the `.github/workflows/release.yml` workflow on every `v*` tag push.

Binaries are provided for:
- macOS (amd64, arm64)
- Linux (amd64, arm64)
- Windows (amd64)


## Legacy Docker-based scripts

The original shell scripts (`scripts/`) and Docker targets remain available for reference
but are no longer the primary interface. See `make help` for the full target list.

