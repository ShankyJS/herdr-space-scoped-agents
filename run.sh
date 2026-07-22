#!/bin/sh
# Locate a python3 interpreter and run space_view.py with it.
#
# Invoked by the plugin manifest as `sh run.sh <apply|clear>`. Using a POSIX
# shell wrapper (sh is always present) instead of a hard-coded interpreter path
# lets the plugin adapt to wherever python3 is installed — system, Homebrew,
# /usr/local, pyenv shims on PATH, etc.
set -eu

# Plugin root: Herdr injects HERDR_PLUGIN_ROOT; fall back to this script's dir.
root="${HERDR_PLUGIN_ROOT:-$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)}"

for py in python3 /usr/bin/python3 /usr/local/bin/python3 /opt/homebrew/bin/python3 python; do
    if command -v "$py" >/dev/null 2>&1; then
        exec "$py" "$root/space_view.py" "$@"
    fi
done

echo "herdr-space-scoped-agents: no python3 interpreter found on PATH" >&2
exit 127
