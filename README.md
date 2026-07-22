# herdr-space-scoped-agents

[![License](https://img.shields.io/github/license/ShankyJS/herdr-space-scoped-agents)](LICENSE)

<p align="center">
  <a href="#how-it-works">how it works</a> · <a href="#requirements">requirements</a> · <a href="#install">install</a> · <a href="#actions">actions</a> · <a href="#limitations">limitations</a>
</p>

A [herdr](https://herdr.dev) plugin that scopes the **agent panel** to the space
you're focused on. Only the agents in the current space are listed; switch
spaces and the panel follows. Notifications, toasts, and every other surface
stay global — this affects the agent panel and the agent-keybind navigation
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
focus. Agent views are **transient server-side state** — they are not written to
`config.toml` and do not survive a server restart. So the plugin also declares a
`workspace.focused` event hook that re-applies the filter on every space switch,
which additionally restores it after a restart on your first focus.

`space_view.py` speaks the API socket's newline-delimited JSON protocol
directly, reading the socket path from the `HERDR_SOCKET_PATH` environment
variable that herdr injects into every plugin command. A small `run.sh` wrapper
locates a `python3` interpreter so the plugin does not depend on a fixed
interpreter path.

## Requirements

- **herdr ≥ 0.7.5** — enforced via `min_herdr_version`; older versions lack the
  `agent.view.*` API and linking is refused.
- **python3** on `PATH` (or at a common location: `/usr/bin`, `/usr/local/bin`,
  `/opt/homebrew/bin`). Standard library only — nothing to `pip install`.
- **A POSIX shell** (`sh`) — present by default on macOS and Linux.
- **macOS or Linux.** Unix sockets only; see [Limitations](#limitations).

No third-party packages, no build step, no network access.

## Install

### From GitHub

```bash
herdr plugin install ShankyJS/herdr-space-scoped-agents
herdr plugin list
```

### Local checkout (development)

```bash
git clone https://github.com/ShankyJS/herdr-space-scoped-agents
herdr plugin link ./herdr-space-scoped-agents
```

The filter applies automatically the first time you focus a space. To apply it
immediately, invoke the `enable` action (below).

**To update**, reinstall — nothing is keyed to the checkout path:

```bash
herdr plugin uninstall herdr-space-scoped-agents && herdr plugin install ShankyJS/herdr-space-scoped-agents
```

## Actions

Two actions, available from the command palette or a keybinding:

| Action | Effect |
| --- | --- |
| `enable` | Apply the current-space filter now |
| `clear`  | Clear the filter and show agents from every space |

Bind them in herdr's `config.toml` (keybindings live in user config, not the
plugin manifest — note it is `<plugin_id>.<action_id>`):

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

## Manage

```bash
herdr plugin list
herdr plugin log list --plugin herdr-space-scoped-agents   # inspect hook runs
herdr plugin disable herdr-space-scoped-agents             # turn off
herdr plugin enable  herdr-space-scoped-agents             # turn back on
herdr plugin unlink  herdr-space-scoped-agents             # remove a local link
herdr plugin uninstall herdr-space-scoped-agents           # remove a GitHub install
```

Disabling the plugin stops the hook but leaves any active view in place — run
the `clear` action (or restart herdr) to drop the filter.

## Example

Given three spaces — **Space A** (2 agents), **Space B** (5 agents), **Space C**
(no agents):

| Focused space | Agent panel shows |
| --- | --- |
| Space A | Space A's 2 agents |
| Space B | Space B's 5 agents |
| Space C | nothing (`no matching agents`) |

## Limitations

- **Unix sockets only.** `space_view.py` connects with `AF_UNIX`. Windows uses a
  named-pipe transport this plugin does not implement, so it declares
  `platforms = ["macos", "linux"]` and will not run on Windows.
- **Applies on first focus after a restart, not at boot.** herdr has no
  "server started" plugin hook, so after a restart the filter re-applies the
  first time you focus a space. Use the `enable` action for an immediate apply.
- **Transient by design.** The view lives in the running server, not in
  `config.toml`; the event hook is what keeps it applied.
- **Scopes by space only.** It filters on `workspace_id` — not by agent kind,
  status, or tab.

## License

[MIT](LICENSE).
