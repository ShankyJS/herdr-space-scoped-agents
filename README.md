# herdr-space-scoped-agents

[![ci](https://github.com/ShankyJS/herdr-space-scoped-agents/actions/workflows/ci.yml/badge.svg)](https://github.com/ShankyJS/herdr-space-scoped-agents/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/ShankyJS/herdr-space-scoped-agents)](https://github.com/ShankyJS/herdr-space-scoped-agents/releases/latest)
[![license](https://img.shields.io/github/license/ShankyJS/herdr-space-scoped-agents)](LICENSE)

<p align="center">
  <a href="#how-it-works">how it works</a> · <a href="#install">install</a> · <a href="#actions">actions</a> · <a href="#windows">windows</a> · <a href="#limitations">limitations</a> · <a href="#build-from-source">build</a>
</p>

A [herdr](https://herdr.dev) plugin that scopes the **agent panel** to the space
you're focused on. Only the agents in the current space are listed; switch
spaces and the panel follows. Notifications, toasts, and every other surface
stay global — this touches the agent panel and the agent-keybind navigation
order only.

Without it, the panel lists every agent across every space at once. In a
workspace with many spaces and many agents, the panel you care about is buried.

## How it works

herdr **0.7.5+** exposes *transient declarative agent views* over its API socket
through the `agent.view.set` / `agent.view.clear` methods. This plugin sets a
view filtered to the focused space:

```jsonc
{ "op": "eq", "field": "workspace_id",
  "value": { "context": "current_workspace_id" } }
```

The `current_workspace_id` context makes the view track whichever space has
focus. Agent views are **transient server-side state** — not written to
`config.toml`, and dropped on a server restart. So the plugin also declares a
`workspace.focused` event hook that re-applies the filter on every space switch,
which additionally restores it after a restart on your first focus.

The work is done by a small, dependency-free **Go binary** that speaks the API
socket's newline-delimited JSON protocol — a unix socket on macOS/Linux, a
named pipe on Windows — reading the socket path from the `HERDR_SOCKET_PATH`
environment variable herdr injects into every plugin command.

## Install

```bash
herdr plugin install ShankyJS/herdr-space-scoped-agents
herdr plugin list
```

On install, herdr runs a build step that **downloads the prebuilt binary for
your platform** from the matching [GitHub Release](https://github.com/ShankyJS/herdr-space-scoped-agents/releases)
and **verifies its SHA-256**. No toolchain required. If no prebuilt exists for
your platform/version, it falls back to building from source with `go` (see
[Build from source](#build-from-source)).

Prebuilt targets: macOS (arm64, x86-64), Linux (arm64, x86-64), Windows
(arm64, x86-64).

The filter applies automatically the first time you focus a space after install.
To apply it immediately, invoke the `enable` action (below).

**Update** by reinstalling:

```bash
herdr plugin uninstall herdr-space-scoped-agents && herdr plugin install ShankyJS/herdr-space-scoped-agents
```

## Actions

Two actions, from the command palette or a keybinding:

| Action | Effect |
| --- | --- |
| `enable` | Apply the current-space filter now |
| `clear`  | Clear the filter and show agents from every space |

Bind them in herdr's `config.toml` (keybindings live in user config, not the
plugin manifest; the value is `<plugin_id>.<action_id>`):

```toml
[[keys.command]]
key = "prefix+f"
type = "plugin_action"
command = "herdr-space-scoped-agents.enable"

[[keys.command]]
key = "prefix+F"
type = "plugin_action"
command = "herdr-space-scoped-agents.clear"
```

On **Windows**, bind the `-windows`-suffixed ids instead
(`herdr-space-scoped-agents.enable-windows` / `.clear-windows`) — see
[Windows](#windows).

## Manage

```bash
herdr plugin list
herdr plugin log list --plugin herdr-space-scoped-agents   # inspect hook runs
herdr plugin disable herdr-space-scoped-agents             # turn off
herdr plugin enable  herdr-space-scoped-agents             # turn back on
herdr plugin uninstall herdr-space-scoped-agents           # remove
```

Disabling stops the hook but leaves any active view in place — run `clear` (or
restart herdr) to drop the filter.

## Windows

Windows is supported, with two platform quirks handled in the manifest (both
learned from the [herdr-file-viewer](https://github.com/smarzban/herdr-file-viewer)
plugin's verified findings):

- **Action ids must be unique across platforms** — herdr rejects duplicate
  action ids regardless of platform gating. The Windows launchers use the ids
  `enable-windows` and `clear-windows`; bind those.
- **Launch by absolute path** — herdr can't reliably spawn a relative program on
  Windows, so every command invokes the binary through `$HERDR_PLUGIN_ROOT`
  (stripping the `\\?\` verbatim prefix herdr may report).

> Note: Windows builds are cross-compiled and CI-verified to compile, and use
> [go-winio](https://github.com/microsoft/go-winio) for the named-pipe
> transport. Runtime has had less real-hardware testing than macOS/Linux —
> reports welcome.

## Limitations

- **Applies on first focus after a restart, not at boot.** herdr has no
  "server started" plugin hook, so after a restart the filter re-applies the
  first time you focus a space. Use the `enable` action for an immediate apply.
- **Transient by design.** The view lives in the running server, not in
  `config.toml`; the event hook is what keeps it applied.
- **Scopes by space only.** It filters on `workspace_id` — not by agent kind,
  status, or tab.

## Build from source

Requires [Go](https://go.dev) 1.26+.

```bash
git clone https://github.com/ShankyJS/herdr-space-scoped-agents
cd herdr-space-scoped-agents
go build -o bin/herdr-space-scoped-agents .   # .exe on Windows
herdr plugin link .
```

`herdr plugin link` does not run the install build step, so build the binary
into `bin/` yourself as above. The manifest launches `bin/herdr-space-scoped-agents`.

Cross-compile any target with `GOOS`/`GOARCH`, e.g.:

```bash
GOOS=linux GOARCH=arm64 go build -o bin/herdr-space-scoped-agents .
```

## Trust

herdr plugin listings are discovered automatically from the GitHub topic
`herdr-plugin` and are **not reviewed by herdr** — install at your own
discretion. The source is small: one Go program plus the manifest and install
scripts. It uses only the Go standard library (and go-winio on Windows), makes
no network calls at runtime, and only talks to your local herdr API socket to
set/clear the agent-panel view. The install scripts download a checksum-verified
release binary.

## License

[MIT](LICENSE).
