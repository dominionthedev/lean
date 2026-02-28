# Changelog

All notable changes to lean are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
lean uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

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