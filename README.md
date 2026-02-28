<p align="center">
  <img src="assets/logo.svg" alt="lean logo" width="400">
</p>

# lean ⚡

[![CI](https://github.com/dominionthedev/lean/actions/workflows/ci.yml/badge.svg)](https://github.com/dominionthedev/lean/actions/workflows/ci.yml)
[![Release](https://github.com/dominionthedev/lean/actions/workflows/release.yml/badge.svg)](https://github.com/dominionthedev/lean/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/dominionthedev/lean?color=205&label=latest)](https://github.com/dominionthedev/lean/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dominionthedev/lean)](https://go.dev)
[![License](https://img.shields.io/github/license/dominionthedev/lean)](LICENSE)


> A lightweight, expressive environment profile manager.

lean keeps your `.env` files safe, organized, and human-friendly.
Switch between profiles, protect secrets, restore backups — all from one CLI.

---

## Installation
```bash
go install github.com/dominionthedev/lean@latest
```

Or grab a binary from the [Releases](https://github.com/dominionthedev/lean/releases) page
(Linux, macOS, Windows — amd64 + arm64).

---

## Quick start
```bash
lean init              # interactive setup — creates your first profile
lean create --name prod
lean apply prod        # .env.prod → .env  (backs up the old .env first)
lean list              # see all profiles
lean current           # which profile is active right now
```

---

## Commands

### `lean init`
Interactive setup wizard. Creates your first profile and writes `.env`.
```bash
lean init
```

> Running `lean init --quiet`? lean has feelings about that.

---

### `lean create`
Create a new environment profile.
```bash
lean create --name staging
lean create --name prod --from .env.template
lean create --name test  --from .env.dev --strip   # keys only, no values
lean create --interactive                           # guided prompt
```

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Profile name |
| `--from` | | Copy from a template or existing file |
| `--strip` | `-s` | Strip values (keep keys only) |
| `--interactive` | `-i` | Prompt for name interactively |

---

### `lean apply`
Switch the active environment. Backs up the current `.env` before overwriting.
```bash
lean apply dev
lean apply prod
```

---

### `lean set`
Set (or update) a variable in a profile.
```bash
lean set DEBUG=true
lean set API_KEY=abc123 --profile prod
```

If the profile is currently active, `.env` is updated immediately.

| Flag | Short | Description |
|------|-------|-------------|
| `--profile` | `-p` | Target profile (default: active) |

---

### `lean get`
Get the value of a variable. Output is plain — pipeline-friendly.
```bash
lean get DEBUG
lean get DATABASE_URL --profile prod
lean get SECRET_KEY --profile staging | pbcopy
```

| Flag | Short | Description |
|------|-------|-------------|
| `--profile` | `-p` | Target profile (default: active) |

---

### `lean delete`
Remove a variable from a profile.
```bash
lean delete OLD_KEY
lean delete LEGACY_TOKEN --profile staging
```

Aliases: `del`, `rm`

| Flag | Short | Description |
|------|-------|-------------|
| `--profile` | `-p` | Target profile (default: active) |

---

### `lean list`
List all known profiles. Auto-discovers any `.env.*` files on disk.
```bash
lean list
```
```
⚡ Profiles

  ▶ dev    (active)
  · prod
  · staging
```

---

### `lean current`
Show the active profile.
```bash
lean current
```

---

### `lean restore`
Restore `.env` from a backup. lean takes a snapshot every time `lean apply` runs.
```bash
lean restore              # interactive picker
lean restore dev-20250228-143022.env   # direct
```

---

### `lean version`
Print the current version.
```bash
lean version
```

---

## How it works

lean keeps a `.lean/` folder in your project:
```
.lean/
  state.json       ← active profile, registered profiles, version
  backups/         ← timestamped .env snapshots (created on every apply)
```

`state.json` is safe to commit. The backups folder is local only.

---

## Safety

- **Atomic writes** — lean never writes directly to `.env`. It writes to a temp file and renames, so a crash mid-write can't corrupt your env.
- **Backup on apply** — every `lean apply` snapshots the current `.env` before replacing it. Run `lean restore` to get it back.
- **`.gitignore` aware** — lean's own `.gitignore` excludes `.env` and `.env.*` by default, keeping secrets off GitHub.

---

## Releasing

lean uses [GoReleaser](https://goreleaser.com). To cut a release:
```bash
git tag v1.0.0
git push origin v1.0.0
```

The release workflow builds binaries for Linux, macOS (Intel + Apple Silicon), and Windows — then attaches them to the GitHub release automatically.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

MIT — see [LICENSE](LICENSE).