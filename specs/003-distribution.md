# 003 — Distribution: fetch-or-build

## Context

`herdr plugin install <owner>/<repo>` clones the repo and runs the manifest's
`[[build]]` step on the user's machine (it does *not* run on `herdr plugin link`,
which is for local dev). We want installing to be easy and not require a
toolchain, while still working if no prebuilt binary exists for a platform.

Two reference points shaped this:

- **herdr-spreader** builds with `cargo build --release` at install → every user
  needs Rust. Easiest for the maintainer, worst for users.
- **herdr-file-viewer** downloads a checksum-verified prebuilt and falls back to
  building from source. Best for users; needs a release pipeline.

## Decision

Use the **fetch-or-build** model (file-viewer's approach), per platform:

- `[[build]]` runs `scripts/fetch-or-build.sh` (unix) or `fetch-or-build.ps1`
  (Windows).
- The script reads the version from `herdr-plugin.toml`, downloads the matching
  prebuilt asset from the GitHub Release `v<version>`, and **verifies its
  SHA-256**.
- On any miss (no matching release, download/checksum failure, unsupported
  platform), it falls back to `go build`.

So a normal install needs **no toolchain**; Go is only required for the
source-build fallback.

## Rationale / alternatives

- **Compile-on-install (rejected):** forces Go on every user for what is normally
  a download.
- **Download-only, no fallback (rejected):** a missing asset would make install
  fail hard; the source fallback keeps installs resilient.
- **Version coupling:** the download URL is derived from the manifest `version`,
  so the manifest version and the git tag **must** match. This is enforced in
  practice by release automation (see [005](005-release-automation.md)), which
  bumps the manifest and tag together. The version line carries a
  `# x-release-please-version` marker; the install scripts strip the trailing
  comment when parsing.

## Assets

Each release publishes 6 binaries + 6 `.sha256` files:
`herdr-space-scoped-agents-{darwin,linux,windows}-{amd64,arm64}[.exe]`.
