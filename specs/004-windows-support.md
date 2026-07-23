# 004 — Windows support

## Context

Herdr runs on Windows, where the API "socket" is a **named pipe**, not a unix
socket, and where plugin command spawning has quirks that bit the
herdr-file-viewer plugin (its manifest documents them, verified on real
hardware). We want Windows to work rather than be declared unsupported.

## Decision

1. **Transport.** `ipc_windows.go` dials the named pipe with
   `winio.DialPipe(path, …)` (from go-winio), where `path` is the
   `HERDR_SOCKET_PATH` value Herdr provides. Unix uses `net.Dial("unix", …)`.
   Both are selected by build tags so each OS compiles only its own transport.

2. **Unique action ids per platform.** Herdr rejects duplicate action ids at
   manifest load time *regardless of platform gating*. So the Windows launchers
   use distinct ids `enable-windows` / `clear-windows` (the unix ones are
   `enable` / `clear`). Bind the `-windows` ids on Windows.

3. **Launch by absolute path.** Herdr cannot reliably spawn a *relative* program
   on Windows (it resolves the relative program against Herdr's own directory).
   So every Windows command invokes the binary by absolute path via
   `$HERDR_PLUGIN_ROOT`, stripping the `\\?\` verbatim prefix Herdr may report:

   ```powershell
   $p=$env:HERDR_PLUGIN_ROOT
   if ($p -and $p.StartsWith('\\?\')) { $p = $p.Substring(4) }
   & (Join-Path $p 'bin\herdr-space-scoped-agents.exe') apply
   ```

## Rationale / alternatives

- **Raw syscalls for the pipe (rejected):** error-prone; go-winio is the standard,
  well-maintained option and is Windows-only by construction.
- **A single set of action ids (rejected):** fails Herdr's duplicate-id check.

## Status / caveat

The Windows build is cross-compiled and CI-verified to compile, and uses the same
protocol as unix. Runtime has had less real-hardware testing than macOS/Linux;
issues and reports are welcome. No aarch64-specific Windows caveats are known;
both `windows-amd64` and `windows-arm64` are published.
