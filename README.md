<p align="center">
  <a href="https://github.com/dominionthedev/lean">
    <img src="assets/logo.svg" alt="lean logo" width="400">
  </a>
</p>

# lean ⚡️

[![CI](https://github.com/dominionthedev/lean/actions/workflows/ci.yml/badge.svg)](https://github.com/dominionthedev/lean/actions/workflows/ci.yml)
[![Release](https://github.com/dominionthedev/lean/actions/workflows/release.yml/badge.svg)](https://github.com/dominionthedev/lean/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/dominionthedev/lean?label=latest)](https://github.com/dominionthedev/lean/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dominionthedev/lean)](https://pkg.go.dev/github.com/dominionthedev/lean)
[![License](https://img.shields.io/github/license/dominionthedev/lean)](LICENSE)

> A lightweight environment profile manager and orchestrator.

Lean manages named environment profiles, resolves inheritance, protects secrets, and keeps the application's `.env` synchronized with the active profile.

## Installation

```bash
go install github.com/dominionthedev/lean@latest
```

Or install a released binary from the [Releases](https://github.com/dominionthedev/lean/releases) page.

## Quick start

```bash
lean init
lean create --name prod
lean apply prod        # switch profile and update .env
lean update            # re-sync .env from the active profile
lean current
lean list
```

## Profiles

A normal profile is a root-level `.env.<name>` file. `.env` is not a profile: it is the resolved environment currently applied to the application.

For example:

```text
.env.base
.env.development
.env.production
.env                  # current application of the active profile
```

Switching profiles with `lean apply <profile>` resolves inheritance and writes the result to `.env`. `lean update` performs the same synchronization for the profile that is already active.

Lean snapshots the existing `.env` before an apply/update replacement. Snapshots are the existing rollback mechanism; Lean does not create a separate backup for every profile edit.

### Profile lifecycle

```bash
lean create --name staging
lean edit staging
lean profile edit staging
lean profile delete staging
lean profile archive staging
lean profile restore staging
lean profile list-archived
```

Archived profiles are the exception to the root-level layout. They are stored under `.lean/archive/profiles/` and can be restored to `.env.<name>`.

If a registered profile file disappears from disk, Lean reports it as stale and asks before removing the profile from Lean's state. It does not silently delete stale registrations.

## Templates and examples

Lean recognizes `.env.template` and `.env.example` as shareable starting points.

Lean does not copy profile values into these files during initialization. Values are stripped by default when creating a profile from a template as well:

```bash
lean template list
lean template add path/to/template
lean template create-from .env.template --name production
```

You can explicitly preserve values with `--strip=false`, but shareable templates and examples should contain only safe example data. Schema entries marked `secret = true` are the intended way to identify sensitive variables.

## Configuration

Project configuration lives at `.lean/lean.toml`.

User-wide configuration lives at `~/.lean/config.toml`.

Local configuration is project-specific and may contain:

```toml
[lean]
version = 1
default_profile = "development"

[schema]
path = ".lean/schema.toml"

[output]
format = "text"
```

Global configuration contains user-level settings such as the editor and secret backend:

```toml
[lean]
version = 1
editor = "nvim"

[secrets]
backend = "local"
```

Secret backend configuration is never written to `.lean/lean.toml` by Lean. The global file is `~/.lean/config.toml`.

```bash
lean config
lean config init
lean config init --global
lean config edit
lean config edit --global
lean config set --default-profile development
lean config set --global --editor nvim
lean config set --global --secrets-backend age
```

The old `~/.lean/lean.toml` location is accepted as a legacy read-only fallback when the new global config does not exist.

## Secrets

Lean stores encrypted secret material under the project-local `.lean/secrets/` directory. The encryption backend is configured globally.

Backends are `local`, `gpg`, `age`, and `ssh`.

For the local backend, Lean uses a random master key at `~/.lean/key` unless `secrets.master_key_file` is configured globally. Lean creates the local key automatically when a secret operation first needs it.

```bash
lean secret put JWT_SECRET=supersecret
lean secret get JWT_SECRET
lean secret list
lean secret inject
```

You can also explicitly create the local key:

```bash
lean secret keygen
```

## Git safety

When `lean init` runs inside a Git worktree, Lean adds these rules to the repository's local `.git/info/exclude`:

```gitignore
.lean
.env.*
!.env.template
!.env.example
```

Lean uses `.git/info/exclude` rather than modifying the project's tracked `.gitignore`. This keeps project-local Lean state and environment profiles out of Git while allowing the template and example files to be committed.

The ignore rules do not make a template intrinsically safe. Lean therefore generates `.env.template` and `.env.example` without values by default.

## Commands

### `lean init`

Interactive project initialization. It creates the project-local `.lean` state/configuration, schema template, first profile, `.env`, and safe value-stripped template/example files.

### `lean apply [profile]`

Switch the active profile. The profile is resolved, the previous `.env` is snapshotted, and the resolved result becomes `.env`.

### `lean update`

Synchronize `.env` with the active profile without changing which profile is active.

```bash
lean update
```

### `lean current`

Show the active profile.

### `lean list`

List registered profiles and reconcile profiles found as root-level `.env.<name>` files.

### `lean set KEY=VALUE`

Set a value in a profile. When targeting the active profile, Lean also re-resolves it into `.env`.

```bash
lean set DEBUG=true
lean set API_KEY=abc123 --profile production
```

### `lean delete KEY`

Delete a variable from a profile. Active profiles are re-resolved into `.env` afterward.

### `lean edit [profile]`

Open a profile in the configured editor, `$EDITOR`, or a detected editor.

### `lean profile`

Manage profile lifecycle and archived profiles.

### `lean template`

Manage `.env.template` and `.env.example` templates.

### `lean schema`

Edit the project schema interactively.

### `lean context`

Manage multi-file environment contexts.

### `lean snapshot` / `lean snapshots` / `lean restore`

Create, inspect, and restore `.env` snapshots. Applying or updating a profile automatically snapshots the current `.env` first.

### `lean validate`

Validate a profile against its schema.

### `lean format`

Render a profile as JSON, YAML, TOML, or env-style output.

### `lean import`

Use remembered workspace/profile mappings.

### `lean meta`

Manage profile metadata.

### `lean man`

Generate or display the Lean manual.

## Schema

The project schema is normally `.lean/schema.toml`:

```toml
[keys.DATABASE_URL]
required = true
secret = true

[keys.NODE_ENV]
values = ["development", "test", "production"]

[keys.JWT_SECRET]
required = true
secret = true
```

Schemas can also define defaults, dependencies, `same_as`, conditional requirements, and `from:` sources.

## Project and user state

A project normally looks like this:

```text
.
├── .env
├── .env.development
├── .env.production
├── .env.template
├── .env.example
└── .lean/
    ├── lean.toml
    ├── schema.toml
    ├── state.json
    ├── secrets/
    ├── backups/
    └── archive/
        └── profiles/
```

`~/.lean/` is for user-wide Lean state, such as:

```text
~/.lean/
├── config.toml
├── key
└── workspaces.json
```

Global templates are not the location for project profiles or project schemas. Normal project profiles remain `.env.<name>` at the project root, and project schema/configuration remain under the project's `.lean/` directory. Archived profiles are stored under the project's `.lean/archive/`.

## Safety model

- **Profile/application separation:** `.env.<name>` is a profile; `.env` is the resolved application state.
- **Snapshots on replacement:** applying or updating a profile snapshots the current `.env` before replacement.
- **Stale-profile confirmation:** missing registered profiles are reported and require confirmation before removal from state.
- **Secret encryption:** secrets are encrypted through the configured backend.
- **Local Git protection:** project-local Lean state and `.env.*` files are excluded through `.git/info/exclude` when initialization occurs in a Git worktree.
- **Safe templates:** initialization and template creation strip values by default, so secrets are not copied into shareable examples.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).

<p align="center">
<a href="https://github.com/dominionthedev">GitHub</a> • <a href="https://dominiondev.leraniode.org">Website</a>
</p>
