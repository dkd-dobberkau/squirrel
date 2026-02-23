# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1] - 2026-02-24

### Fixed

- Auto-remove macOS quarantine flag during `brew install` via post-install hook
- Added caveat hint for manual quarantine removal if needed

## [0.2.0] - 2026-02-23

### Added

- GoReleaser configuration for automated multi-platform releases (darwin/linux, amd64/arm64)
- GitHub Actions workflow for release automation on tag push
- Homebrew cask distribution via `brew install dkd-dobberkau/tap/squirrel`

## [0.1.1] - 2026-02-23

### Added

- `--version` flag to display the current version
- Build-time version injection via `go build -ldflags "-X main.version=X.Y.Z"`

### Changed

- Removed `go-git` dependency in favor of native git commands for smaller binary and reliable `.gitignore` support

## [0.1.0] - 2026-02-23

Initial release.

### Added

- CLI with `status`, `stash`, `timeline`, and `project` commands (cobra)
- History parser for `~/.claude/history.jsonl` with per-project aggregation
- Sessions parser for `sessions-index.json` with project enrichment
- Git status detection (dirty state, branch, feature branch, uncommitted files)
- Project categorization into Open Work, Recent Activity, and Sleeping
- Priority scoring based on recency, activity, dirty state, and branch
- Terminal output with lipgloss styling
- JSON output mode (`--json`)
- Claude Code `/squirrel` skill integration
- `install-skill` command for automated skill setup

### Fixed

- Use native `git status --porcelain` instead of go-git to correctly respect `.gitignore`, `.git/info/exclude`, and global gitignore

[0.2.1]: https://github.com/dkd-dobberkau/squirrel/releases/tag/v0.2.1
[0.2.0]: https://github.com/dkd-dobberkau/squirrel/releases/tag/v0.2.0
[0.1.1]: https://github.com/dkd-dobberkau/squirrel/releases/tag/v0.1.1
[0.1.0]: https://github.com/dkd-dobberkau/squirrel/releases/tag/v0.1.0
