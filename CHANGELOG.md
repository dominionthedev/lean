# Changelog

All notable changes to lean are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
lean uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- **Profile inheritance** — profiles can extend a parent via `# lean:extends <name>`, `# @extends <name>`, or `LEAN_EXTENDS=<name>`
  - `lean apply` resolves the full inheritance chain before writing `.env`
  - `lean get` / `lean format` resolve inherited values
  - `lean create --extends <parent>` writes the directive automatically
  - `lean list` shows inheritance arrows (`development → base`)
  - Cycle detection and max-depth guard
- `lean diff <left> <right>` — compare two profiles (supports `current` alias)
- `lean validate [profile]` — check required keys against a schema
  - Schema sources: `--schema`, `.env.schema`, `.env.example`, `.env.template`
- **Named snapshots**
  - `lean snapshot <name>` — save current `.env` under a human-readable name
  - `lean snapshots` — list named + automatic snapshots
  - `lean snapshot delete <name>` — remove a named snapshot
  - `lean restore <name>` accepts bare named-snapshot labels
- **Workspace awareness**
  - `lean apply` remembers cwd → profile in `~/.config/lean/config.json`
  - `lean import` applies the remembered profile for the current directory (`--yes` to skip confirm)
- **Profile metadata**
  - `lean meta [profile]` — view or set description, author, tags
  - Stored in `.lean/state.json` under `meta`
- **Contexts** — multi-file environment bundles
  - `lean context create <name> [--profile] [--description]`
  - `lean context add <name> <source> <target>`
  - `lean context apply <name>` — writes all mappings (env profiles resolve inheritance)
  - `lean context list|show|remove|delete`
  - Definitions stored in `.lean/contexts/<name>.json`

---

## [0.1.0] — 2026-02-28

### Added
- `lean init` — interactive setup wizard using Huh prompts
  - `--quiet` flag responds playfully and redirects to `lean create`
- `lean create` — create environment profiles
  - `--name` / `-n` flag for profile name
  - `--from` flag to copy from a template or existing file
  - `--strip` / `-s` flag to strip values (keys only)
  - `--interactive` / `-i` flag for guided prompt
- `lean apply` — apply a profile to `.env`
  - atomic write (temp file + rename — no partial writes)
  - automatic `.env` snapshot to `.lean/backups/` before every switch
- `lean list` — list all profiles, auto-discovers `.env.*` files on disk
- `lean current` — show the active profile
- `lean restore` — restore `.env` from a backup with an interactive picker
- `lean set KEY=VALUE` — set or update a variable in a profile
  - `--profile` / `-p` to target a non-active profile
  - syncs `.env` immediately if the profile is active
- `lean get KEY` — get a variable's value (plain output, pipeline-friendly)
  - `--profile` / `-p` to read from a non-active profile
- `lean delete KEY` — remove a variable from a profile
  - aliases: `del`, `rm`
  - `--profile` / `-p` to target a non-active profile
- `lean version` — print the current version
- Internal `env` package — structured `.env` parser with Get / Set / Delete / Strip / Write
- Internal `backup` package — snapshot, list, and restore backups
- Internal `ui` package — Lipgloss-based styling (Bolt, Ok, Fail, Warn, Info, Faint)
- `.lean/state.json` — tracks active profile, registered profiles, version
- CI workflow — build, vet, and cross-compile on push and pull requests
- Release workflow — GoReleaser on tag push, binaries for Linux / macOS / Windows

[Unreleased]: https://github.com/dominionthedev/lean/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/dominionthedev/lean/releases/tag/v0.1.0