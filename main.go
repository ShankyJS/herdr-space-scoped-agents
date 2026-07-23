// Command herdr-space-scoped-agents scopes Herdr's agent panel to the focused
// space (or shows every space), by driving Herdr's transient "agent view" over
// the local API socket named by $HERDR_SOCKET_PATH (a unix socket on
// macOS/Linux, a named pipe on Windows).
//
// The chosen mode ("current" or "all") is persisted in the plugin state dir so
// it sticks across space switches and server restarts; the workspace.focused
// hook calls "sync" to (re)assert whichever mode is active.
//
//	herdr-space-scoped-agents current   # scope to the focused space
//	herdr-space-scoped-agents all       # show agents from every space
//	herdr-space-scoped-agents toggle    # flip between current and all
//	herdr-space-scoped-agents sync      # apply whatever mode is persisted (hook)
//	herdr-space-scoped-agents status    # print the persisted mode
//
// (apply/enable and clear/disable remain accepted aliases for current/all.)
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// version is stamped at build time with -ldflags "-X main.version=...".
var version = "dev"

// source identifies this plugin as the owner of the agent view, so it can be
// cleared without disturbing views set by anything else.
const (
	source = "herdr-space-scoped-agents"
	label  = "Current space"

	modeCurrent = "current" // scope the panel to the focused space
	modeAll     = "all"     // show agents from every space
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
	// Fallback for manual runs outside a plugin hook: honor XDG_CONFIG_HOME the
	// same way Herdr locates its own config, before defaulting to ~/.config.
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "herdr", "herdr.sock")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "herdr", "herdr.sock")
}

// modeFile returns the path of the persisted-mode file, or "" if no writable
// location is known (manual runs outside a plugin hook).
func modeFile() string {
	dir := os.Getenv("HERDR_PLUGIN_STATE_DIR")
	if dir == "" {
		dir = os.Getenv("HERDR_PLUGIN_CONFIG_DIR")
	}
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "mode")
}

// readMode returns the persisted mode, defaulting to "current" (this plugin is
// installed specifically to scope the panel, so scoping is the sensible default).
func readMode() string {
	path := modeFile()
	if path == "" {
		return modeCurrent
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return modeCurrent
	}
	if strings.TrimSpace(string(data)) == modeAll {
		return modeAll
	}
	return modeCurrent
}

// writeMode persists the mode (best-effort; a failure is reported but does not
// abort the view change).
func writeMode(mode string) {
	path := modeFile()
	if path == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "%s: could not create state dir: %v\n", source, err)
		return
	}
	if err := os.WriteFile(path, []byte(mode+"\n"), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "%s: could not persist mode: %v\n", source, err)
	}
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

// applyView filters the agent panel to the focused space.
func applyView() (map[string]any, error) {
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

// clearView removes this plugin's view, showing agents from every space.
func clearView() (map[string]any, error) {
	return call("agent.view.clear", map[string]any{"source": source})
}

// applyMode performs the view change for the given mode.
func applyMode(mode string) (map[string]any, error) {
	if mode == modeAll {
		return clearView()
	}
	return applyView()
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s <current|all|toggle|sync|status|version>\n", filepath.Base(os.Args[0]))
}

func main() {
	action := "sync"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	var (
		resp map[string]any
		err  error
	)
	switch action {
	case "current", "apply", "enable":
		writeMode(modeCurrent)
		resp, err = applyView()
	case "all", "clear", "disable":
		writeMode(modeAll)
		resp, err = clearView()
	case "toggle":
		mode := modeCurrent
		if readMode() == modeCurrent {
			mode = modeAll
		}
		writeMode(mode)
		resp, err = applyMode(mode)
	case "sync":
		// Used by the workspace.focused hook: assert whatever mode is persisted.
		resp, err = applyMode(readMode())
	case "status":
		fmt.Println(readMode())
		return
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
