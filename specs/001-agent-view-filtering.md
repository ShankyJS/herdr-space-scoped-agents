# 001 — Agent-view filtering

## Context

The goal is to scope Herdr's agent panel to the space you're focused on, so only
that space's agents are listed. Earlier this behavior existed as a fork of Herdr
with a persistent config key (`ui.agent_panel_scope = "current"`). Upstream Herdr
took a different design: as of 0.7.5 it exposes **transient declarative agent
views** over the API socket via `agent.view.set` / `agent.view.clear`, and
`ui.agent_panel_sort` only *groups* the panel (`spaces` / `priority`) without
hiding other spaces.

There is no config key and no CLI verb for `agent.view.*` — they are socket
methods only.

## Decision

Set an agent view whose filter matches the currently focused workspace:

```jsonc
{ "op": "eq", "field": "workspace_id",
  "value": { "context": "current_workspace_id" } }
```

The `current_workspace_id` **context** is the key: it is resolved dynamically by
the server, so the view tracks whichever space has focus rather than being pinned
to one id.

Because agent views are **transient server-side state** — not persisted to
`config.toml`, and dropped when the server restarts — the plugin also registers a
`workspace.focused` **event hook** that re-applies the filter on every space
switch. That keeps the filter correct across switches and re-establishes it after
a server restart (on the first focus).

Two manual actions, `enable` (apply) and `clear` (`agent.view.clear`), let the
user toggle it from the command palette or a keybinding.

## Rationale / alternatives

- **`ui.agent_panel_sort = "spaces"` (rejected as insufficient):** groups the
  panel by space but still shows every space's agents. It does not filter.
- **Pinning to a literal `workspace_id` instead of the context (rejected):**
  would require re-setting on every focus change to a concrete id and would not
  self-heal; the dynamic context is simpler and correct by construction. The
  focus hook is still used, but only to (re)assert the view, not to compute an id.
- **Wire protocol:** the API socket speaks newline-delimited JSON with no
  handshake — connect, write one `{"id","method","params"}` line, read one line
  back. This is what makes a tiny external binary a viable client.

## Consequence / known limitation

There is no "server started" plugin hook, so after a restart the filter is
re-applied on the first `workspace.focused` (in practice, as soon as you focus a
space), not instantly at boot. See the README Limitations section.
