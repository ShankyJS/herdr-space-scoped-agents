# specs

Design notes and engineering decisions for `herdr-space-scoped-agents`, one
topic per doc. These explain *why* the plugin is built the way it is, so future
changes are made with the original reasoning in view.

| # | Doc | Topic |
| --- | --- | --- |
| 001 | [agent-view-filtering.md](001-agent-view-filtering.md) | How the panel is scoped: the transient `agent.view.*` API and the focus hook |
| 002 | [why-go-binary.md](002-why-go-binary.md) | Why a Go binary rather than a script or Rust |
| 003 | [distribution.md](003-distribution.md) | Fetch-or-build install: prebuilt download + checksum, source fallback |
| 004 | [windows-support.md](004-windows-support.md) | Named-pipe transport and Windows launch quirks |
| 005 | [release-automation.md](005-release-automation.md) | release-please and the draft → build → publish flow |

Format: each doc states the **context**, the **decision**, and the
**rationale / alternatives** considered.
