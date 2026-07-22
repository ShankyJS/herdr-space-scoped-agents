#!/usr/bin/env python3
"""Space-scoped agents plugin for Herdr.

Sets a transient declarative agent view (`agent.view.set`) that filters the
agent panel down to the currently focused space, or clears it
(`agent.view.clear`). Talks to the Herdr API socket directly using the
newline-delimited JSON protocol.

Usage:
    space_view.py apply   # filter panel to the focused space
    space_view.py clear   # show agents from every space again
"""
import json
import os
import socket
import sys

# Identifies this plugin as the owner of the view so it can be cleared
# without disturbing views set by other sources.
SOURCE = "herdr-space-scoped-agents"
LABEL = "Current space"


def socket_path():
    # Herdr injects HERDR_SOCKET_PATH into every plugin command. Fall back to
    # the default stable-server socket for manual runs outside a hook.
    return os.environ.get("HERDR_SOCKET_PATH") or os.path.expanduser(
        "~/.config/herdr/herdr.sock"
    )


def call(method, params):
    if not hasattr(socket, "AF_UNIX"):
        # Windows uses a named pipe transport; not handled by this script.
        raise RuntimeError("space-scoped-agents currently supports unix sockets only")

    path = socket_path()
    sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    sock.settimeout(5)
    try:
        sock.connect(path)
        request = {
            "id": f"{SOURCE}:{method}",
            "method": method,
            "params": params,
        }
        sock.sendall((json.dumps(request) + "\n").encode())

        buf = b""
        while b"\n" not in buf:
            chunk = sock.recv(65536)
            if not chunk:
                break
            buf += chunk
    finally:
        sock.close()

    if not buf:
        raise RuntimeError("empty response from herdr api socket")

    response = json.loads(buf.decode().splitlines()[0])
    if "error" in response:
        raise RuntimeError(f"api error: {response['error']}")
    return response


def apply_view():
    return call(
        "agent.view.set",
        {
            "source": SOURCE,
            "label": LABEL,
            "filter": {
                "op": "eq",
                "field": "workspace_id",
                "value": {"context": "current_workspace_id"},
            },
        },
    )


def clear_view():
    return call("agent.view.clear", {"source": SOURCE})


def main():
    action = sys.argv[1] if len(sys.argv) > 1 else "apply"
    if action == "apply":
        response = apply_view()
    elif action == "clear":
        response = clear_view()
    else:
        print(f"unknown action: {action!r} (expected 'apply' or 'clear')", file=sys.stderr)
        return 2

    print(json.dumps(response.get("result", response)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
