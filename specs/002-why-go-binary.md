# 002 — Why a Go binary

## Context

The plugin's actual work is tiny: open the Herdr API socket, write one JSON line,
read one line back. Herdr plugins are plain subprocesses — the manifest `command`
is an argv the server runs per event/action — so almost any executable works
(shell, interpreter, or native binary). The first version was a Python script.

## Decision

Ship a single, dependency-free **Go binary**.

## Rationale / alternatives

- **Python script (replaced):** worked, but depends on a `python3` interpreter
  being present and reachable. It also pushed a runtime dependency onto every
  user's machine. Fine for a personal tool; weak for something published.
- **POSIX shell only (rejected):** writing to a unix socket *and* a Windows named
  pipe from portable shell is painful and unreliable (`nc -U` behavior varies; no
  clean Windows story).
- **Rust (rejected):** perfectly capable, but for a 6-target cross-platform CLI it
  means fighting cross-compilation toolchains, and it buys nothing over Go for a
  socket client.
- **Go (chosen):** trivial cross-compilation via `GOOS`/`GOARCH` to all six
  targets (darwin/linux/windows × amd64/arm64), small static binaries, native
  unix-socket support, and a clean Windows named-pipe path via
  [go-winio](https://github.com/microsoft/go-winio). Fast CI. The only third-party
  dependency is go-winio, imported behind a `//go:build windows` constraint so
  non-Windows builds stay pure standard library.

## Layout

- `main.go` — CLI (`apply` / `clear` / `version`), request build, socket I/O.
- `ipc_unix.go` (`//go:build !windows`) — `net.Dial("unix", …)`.
- `ipc_windows.go` (`//go:build windows`) — `winio.DialPipe(…)`.

The binary reads the socket path from `HERDR_SOCKET_PATH` (injected by Herdr into
every plugin command), falling back to the XDG/`~/.config` location for manual
runs.
