# keycloak-cli

A portable Go CLI (`kc`) to manage Keycloak realms, users, groups and clients.
No Docker. No Java. A single static binary.


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
# one-time setup
kc config set server https://keycloak.example.com
kc config set username admin
kc config set password                 # prompts
kc config set secret-header-name my-header
kc config set secret-header-value     # prompts

# authenticate (uses stored config — no flags needed)
kc login

# use
kc realm list
kc user list -r myrealm
kc client list -r myrealm
```


## Configuration

### Stored config (`~/.config/kc/config.json`)

Use `kc config` to manage persistent configuration. Values here are the lowest-priority
fallback — overridden by env vars or CLI flags.

```bash
kc config list                        # list available keys
kc config show                        # show current values (sensitive keys masked)
kc config set <key> [value]           # set a key; prompts for value if sensitive and omitted
```

| Key | Description |
|---|---|
| `server` | Keycloak server URL |
| `username` | Admin username |
| `password` | Admin password *(sensitive — masked in show)* |
| `secret-header-name` | Extra header name forwarded on every admin request |
| `secret-header-value` | Extra header value *(sensitive — masked in show)* |

Config and token files are stored in `~/.config/kc/` (mode `0600`).
The token is refreshed automatically before it expires.

### Environment variables

Env vars override stored config. Useful for CI or scripting.

| Variable | Description |
|---|---|
| `KC_SERVER` | Keycloak server URL |
| `KC_ADMIN_USER` | Admin username |
| `KC_ADMIN_PASSWORD` | Admin password |
| `KC_ADMIN_SECRET_HEADER_NAME` | Extra header name |
| `KC_ADMIN_SECRET_HEADER_VALUE` | Extra header value |
| `REALM_NAME` | Pre-select realm for any command that asks for one |

Example `.envrc` (direnv):

```bash
export KC_SERVER="https://keycloak.example.com"
export KC_ADMIN_PASSWORD="<admin-password>"
# optional
export KC_ADMIN_SECRET_HEADER_NAME="my-header"
export KC_ADMIN_SECRET_HEADER_VALUE="<header-value>"
export REALM_NAME="my-realm"
```

### Priority

For every value: **CLI flag → env var → stored config → interactive prompt**


## Commands

### `kc login`

Authenticate with Keycloak. Stores the access + refresh token in `~/.config/kc/token.json`.
Uses stored config as a fallback — if all values are already set via `kc config`, running
`kc login` requires no flags or prompts.

```
Flags:
  -s, --server string               Keycloak server URL (overrides KC_SERVER / stored config)
  -u, --user string                 Admin username (overrides KC_ADMIN_USER / stored config)
      --secret-header-name string   Extra header name (overrides KC_ADMIN_SECRET_HEADER_NAME / stored config)
```

The header value is never accepted as a CLI flag — use `KC_ADMIN_SECRET_HEADER_VALUE`,
stored config, or `kc login` will prompt for it.

```bash
kc login
kc login -s https://kc.example.com -u admin
kc login --secret-header-name my-header
```

### `kc config`

```bash
kc config list
kc config show
kc config set server https://keycloak.example.com
kc config set username admin
kc config set password                 # prompts
kc config set secret-header-name my-header
kc config set secret-header-value     # prompts
```

### `kc version`

```bash
kc version
kc --version
```

### Realm

```bash
kc realm list
kc realm create
kc realm delete     # interactive multi-select + confirmation; master realm excluded
```

### User

All user commands support `-r`/`--realm` (flag → `REALM_NAME` env var → stored config → prompt).

```bash
kc user list        # lists users enriched with their group memberships
kc user create      # interactive: username, email, name, password, group assignment
kc user delete      # interactive multi-select + confirmation
```

```bash
kc user list -r myrealm
kc user create -r myrealm
kc user delete -r myrealm
```

### Group

```bash
kc group list       # lists groups enriched with their members
kc group create     # interactive: name, optional parent group
kc group delete     # interactive multi-select + confirmation
kc group add-member # multi-select users × groups (Cartesian assignment)
```

```bash
kc group list -r myrealm
kc group add-member -r myrealm
```

### Client

```bash
kc client list
kc client create    # interactive: client ID, name, OAuth2 flows, redirect URIs
kc client delete    # interactive multi-select + confirmation
```

```bash
kc client list -r myrealm
kc client create -r myrealm
kc client delete -r myrealm
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
kc client-scope create -r myrealm
kc client-scope add -r myrealm
```


## Distribution

Releases are built with [goreleaser](https://goreleaser.com/) and published to GitHub Releases
via the `.github/workflows/release.yml` workflow on every `v*` tag push.

Binaries are provided for:
- macOS (amd64, arm64)
- Linux (amd64, arm64)


## Legacy Docker-based scripts

The original shell scripts (`scripts/`) and Docker targets remain available for reference
but are no longer the primary interface. See `make help` for the full target list.
