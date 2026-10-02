<p align="center">
  <a href="https://github.com/dominionthedev/lean">
    <img src="assets/logo.svg" alt="lean logo" width="400">
  </a>
</p>

# lean ⚡️

[![CI](https://github.com/dominionthedev/lean/actions/workflows/ci.yml/badge.svg)](https://github.com/dominionthedev/lean/actions/workflows/ci.yml)
[![Release](https://github.com/dominionthedev/lean/actions/workflows/release.yml/badge.svg)](https://github.com/dominionthedev/lean/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/dominionthedev/lean?color=205&label=latest)](https://github.com/dominionthedev/lean/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/dominionthedev/lean)](https://pkg.go.dev/github.com/dominionthedev/lean)
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
lean create --name staging --extends base           # inherit from base
lean create --interactive                           # guided prompt
```

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | Profile name |
| `--from` | | Copy from a template or existing file |
| `--strip` | `-s` | Strip values (keep keys only) |
| `--extends` | | Inherit from a parent profile |
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

### `lean edit`
Open a profile in your editor.
```bash
lean edit
lean edit prod
```
Uses `$EDITOR`, or defaults to common editors like `nano` or `vim`.

---

### `lean template`
Manage environment templates. Auto-discovers `.env.template` and `.env.example`.
```bash
lean template list
lean template add path/to/template
lean template create-from .env.template --name prod
```

---

### `lean format`
Convert a profile to different formats.
```bash
lean format --type json
lean format prod --type yaml
```
Supported types: `json`, `yaml`, `toml`, `env`.

---

### `lean current`
Show the active profile.
```bash
lean current
```

---

### `lean restore`
Restore `.env` from a backup. lean takes a snapshot every time `lean apply` runs.
Also accepts named snapshot labels.
```bash
lean restore              # interactive picker
lean restore before-migration
lean restore dev-20250228-143022.env   # direct
```

---

### `lean snapshot`
Save a named snapshot of the current `.env`.
```bash
lean snapshot before-migration
lean snapshot before-testing
lean snapshots                 # list named + automatic
lean snapshot delete before-testing
```

---

### `lean import`
Workspace awareness — lean remembers which profile you last applied in each directory.
```bash
lean apply api-dev             # remembers ~/Projects/api → api-dev
cd ~/Projects/api
lean import                    # suggests / applies api-dev
lean import --yes              # skip confirmation
```

---

### `lean meta`
View or set profile metadata (description, author, tags).
```bash
lean meta production
lean meta production --description "Main production API" --author DominionDev --tags aws,production
```

---

### `lean config`
Project-level configuration via `lean.toml`.
```bash
lean config init
lean config
lean config set --default-profile development
lean config set --secrets-backend gpg --secrets-recipient you@example.com
```

---

### `lean fill`
Fill missing keys from schema defaults, `same_as` links, and `from:` sources.
```bash
lean fill
lean fill production --dry-run
```

---

### `lean secret`
Encrypt secrets into `.lean/secrets/` (never plain in git).
```bash
lean secret keygen                         # writes ~/.lean/key (mode 0600)
lean secret put JWT_SECRET=supersecret
lean secret get JWT_SECRET
lean secret list
lean secret inject                         # write secrets into .env
```
Master key resolution (local backend): `LEAN_MASTER_KEY` env → `secrets.master_key_file` → `~/.lean/key`.

Backends: `local` (AES-256-GCM), `gpg`, `age`, `ssh` (age + SSH pubkey).
```toml
# lean.toml — use your SSH key
[secrets]
backend = "ssh"
recipient = "~/.ssh/id_ed25519.pub"
identity = "~/.ssh/id_ed25519"
```

---

### Advanced schema (`.lean/schema.toml`)
```toml
[keys.DATABASE_URL]
required = true

[keys.DEBUG]
values = ["true", "false"]
default = "false"

[keys.REDIS_URL]
same_as = "DATABASE_URL"

[keys.SMTP_HOST]
required_when = { MAIL_DRIVER = "smtp" }
deactivated_when = { MAIL_DRIVER = "log" }

[keys.JWT_SECRET]
required = true
secret = true

[keys.BUILD_SHA]
from = "command:git rev-parse --short HEAD"

[keys.API_KEY]
from = "file:.secrets/api_key"
secret = true

[templates.production]
resolves_to = ".env.production"
extends = "base"
```

---

### `lean context`
Multi-file environment bundles. A context maps several sources onto targets and applies them together.
```bash
lean context create production --profile production --description "Production API stack"
lean context add production .env.secret.prod .env.secret
lean context add production configs/prod.toml config.toml
lean context show production
lean context apply production
lean context list
```
```
✓ .env.production   → .env          (inheritance resolved)
✓ .env.secret.prod  → .env.secret
✓ configs/prod.toml → config.toml
```
Alias: `lean ctx`

---

### `lean diff`
Compare two profiles after resolving inheritance.
```bash
lean diff dev prod
lean diff current prod
```
```
⚡ Diff  development  ↔  production

PORT:
  development: 8080
  production: 80

DEBUG:
  development: true
  production: false
```

---

### `lean validate`
Check that required keys exist in a profile. Schema is loaded from `--schema`, `.env.schema`, `.env.example`, or `.env.template`.
```bash
lean validate production
lean validate staging --schema .env.schema
```
```
✓ PORT
✓ DB_HOST
✗ JWT_SECRET missing
✗ SMTP_PASSWORD missing
```

---

### Profile inheritance
Profiles can extend a parent so shared keys live in one place:

```bash
# .env.base
APP_NAME=Lean
PORT=8080

# .env.development
# lean:extends base
DEBUG=true
DB=localhost

# .env.production
# lean:extends base
DEBUG=false
DB=prod.internal
PORT=80
```

```bash
lean create --name staging --extends base
lean apply production   # merges base + production → .env
lean get PORT -p development   # 8080 (from base)
```

Supported directives (first match wins):
- `# lean:extends base`
- `# @extends base`
- `LEAN_EXTENDS=base`

---

### `lean completion`
Generate shell completion scripts.
```bash
source <(lean completion bash)
```
Supported shells: `bash`, `zsh`, `fish`, `powershell`.

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
  state.json       ← active profile, registered profiles, metadata, version
  backups/         ← timestamped + named .env snapshots
  contexts/        ← multi-file context definitions (*.json)
```

`state.json` and `contexts/` are safe to commit. The backups folder is local only.
Workspace mappings live in `~/.config/lean/config.json`.

---

## Safety

- **Atomic writes** — lean never writes directly to `.env`. It writes to a temp file and renames, so a crash mid-write can't corrupt your env.
- **Backup on apply** — every `lean apply` snapshots the current `.env` before replacing it. Run `lean restore` to get it back.
- **`.gitignore` aware** — lean's own `.gitignore` excludes `.env` and `.env.*` by default, keeping secrets off GitHub.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

MIT — see [LICENSE](LICENSE).

---
<p align="center">
<a href="https://github.com/dominionthedev">GitHub</a> • <a href="https://dominionthedev.github.io">Website</a>
</p>
<p align="center">
  <a href="https://github.com/dominionthedev/dominionthedev">
    <img src="https://raw.githubusercontent.com/dominionthedev/dominionthedev/main/assets/watermark-animated.svg" width="700"/>
  </a>
</p>
