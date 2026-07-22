// Command herdr-space-scoped-agents sets or clears a transient Herdr "agent
// view" that filters the agent panel to the currently focused space.
//
// It speaks Herdr's newline-delimited JSON API protocol directly over the
// local socket named by $HERDR_SOCKET_PATH (a unix socket on macOS/Linux, a
// named pipe on Windows).
//
//	herdr-space-scoped-agents apply   # filter the panel to the focused space
//	herdr-space-scoped-agents clear   # show agents from every space again
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// version is stamped at build time with -ldflags "-X main.version=...".
var version = "dev"

// source identifies this plugin as the owner of the agent view, so it can be
// cleared without disturbing views set by anything else.
const (
	source = "herdr-space-scoped-agents"
	label  = "Current space"
)

type request struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

// socketPath returns the Herdr API socket path. Herdr injects
// HERDR_SOCKET_PATH into every plugin command; the fallback covers manual runs
// outside a plugin hook.
func socketPath() string {
	if p := os.Getenv("HERDR_SOCKET_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "herdr", "herdr.sock")
}

// call sends one request and returns the decoded response object.
func call(method string, params any) (map[string]any, error) {
	path := socketPath()
	if path == "" {
		return nil, fmt.Errorf("could not determine herdr socket path (set HERDR_SOCKET_PATH)")
	}

	conn, err := dialHerdr(path)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", path, err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

	payload, err := json.Marshal(request{ID: source + ":" + method, Method: method, Params: params})
	if err != nil {
		return nil, err
	}
	if _, err := conn.Write(append(payload, '\n')); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil && len(line) == 0 {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var resp map[string]any
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if e, ok := resp["error"]; ok {
		return nil, fmt.Errorf("api error: %v", e)
	}
	return resp, nil
}

func apply() (map[string]any, error) {
	return call("agent.view.set", map[string]any{
		"source": source,
		"label":  label,
		"filter": map[string]any{
			"op":    "eq",
			"field": "workspace_id",
			"value": map[string]any{"context": "current_workspace_id"},
		},
	})
}

func clear() (map[string]any, error) {
	return call("agent.view.clear", map[string]any{"source": source})
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <apply|clear|version>\n", filepath.Base(os.Args[0]))
}

func main() {
	action := "apply"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	var (
		resp map[string]any
		err  error
	)
	switch action {
	case "apply", "enable":
		resp, err = apply()
	case "clear", "disable":
		resp, err = clear()
	case "version", "--version", "-v":
		fmt.Println(version)
		return
	case "help", "--help", "-h":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown action %q\n", action)
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", source, err)
		os.Exit(1)
	}

	if result, ok := resp["result"]; ok {
		if out, err := json.Marshal(result); err == nil {
			fmt.Println(string(out))
			return
		}
	}
	fmt.Println("ok")
}
