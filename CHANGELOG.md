# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- This changelog.

### Changed
- CI: bump `actions/checkout` to v7, `actions/setup-go` to v7, and
  `softprops/action-gh-release` to v3 (Node 24 runtime; clears the Node 20
  deprecation warnings).
- Build on the latest stable Go; `go.mod` now targets Go 1.26.

## [0.2.0] - 2026-07-22

Initial public release.

### Added
- Scope the agent panel to the focused space via a transient `agent.view.set`
  filter (`workspace_id == current_workspace_id`), re-applied on every
  `workspace.focused` event so it tracks focus and survives a server restart.
- `enable` / `clear` actions (with `-windows` variants) for manual toggling.
- Cross-platform Go binary: unix socket on macOS/Linux, named pipe on Windows
  (via [go-winio](https://github.com/microsoft/go-winio)).
- Prebuilt binaries for macOS, Linux, and Windows (arm64 + x86-64), published
  to GitHub Releases with SHA-256 checksums.
- Install-time fetch-or-build: download the verified prebuilt, or build from
  source with Go when no prebuilt is available.

[Unreleased]: https://github.com/ShankyJS/herdr-space-scoped-agents/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/ShankyJS/herdr-space-scoped-agents/releases/tag/v0.2.0
