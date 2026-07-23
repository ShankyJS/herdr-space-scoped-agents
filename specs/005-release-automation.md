# 005 — Release automation

## Context

The plugin's version lives in `herdr-plugin.toml`, and the install script derives
the download URL from it, so **the manifest version and the git tag must match**.
Doing that by hand (bump manifest, edit changelog, tag, push) is error-prone.

Two failure modes to avoid:

1. Publishing a release whose binaries failed to build (a tagged, "latest"
   release with no assets — install would break).
2. A version/tag mismatch that points the install script at a nonexistent asset.

## Decision

Use **release-please** (googleapis/release-please-action v5), driven by
Conventional Commits:

- `fix:` → patch, `feat:` → minor, `feat!:` / `BREAKING CHANGE` → major.
- release-please maintains a **release PR** that bumps the version in both
  `CHANGELOG.md` and `herdr-plugin.toml` (via a `generic` extra-files updater on
  the `# x-release-please-version` marker) and `.release-please-manifest.json`.
  This keeps the manifest and tag in lockstep, killing failure mode 2.
- Merging the release PR creates the release **as a draft** (`"draft": true`).
  A draft release does **not** create the git tag.
- The same workflow then builds all six targets, uploads them, and only on
  success runs `gh release edit --draft=false --latest`, which is what creates
  the tag and marks the release latest. A failed build leaves an unpublished
  draft with no tag — no real release ships. This kills failure mode 1.

The build is folded into the release-please workflow (not a separate
tag-triggered one) because a tag created with the default `GITHUB_TOKEN` does
not trigger other workflows.

## Rationale / alternatives

- **semantic-release (rejected):** Node-centric, assumes `package.json`, and
  publishes on every push with no gate. Bumping a TOML version needs extra
  plugins. release-please fits a Go + TOML repo and offers a review gate (the PR).
- **Manual tag + tag-triggered build (previous state):** worked but required
  manual version/changelog discipline and could publish a broken release if the
  build failed after tagging.

## Required repo setting

release-please opens PRs, so the repo must allow it:
**Settings → Actions → General → Workflow permissions →**
*Read and write* **+** *Allow GitHub Actions to create and approve pull requests*
(API: `default_workflow_permissions=write`, `can_approve_pull_request_reviews=true`).

## CI

`.github/workflows/ci.yml` runs `go vet` and cross-compiles all six targets on
every push/PR, so a broken build is caught on the release PR *before* merge —
a second layer of protection in front of the draft-publish gate.
