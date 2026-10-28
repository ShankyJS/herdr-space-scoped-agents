//go:build windows

package main

import (
	"net"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

const pipePrefix = `\\.\pipe\`

// dialHerdr connects to the Herdr API named pipe. On Windows, Herdr's
// HERDR_SOCKET_PATH is a file path (e.g. C:\...\herdr\herdr.sock) and the pipe
// is that path under \\.\pipe\, as Herdr's own integrations dial it.
func dialHerdr(path string) (net.Conn, error) {
	if !strings.HasPrefix(path, pipePrefix) {
		path = pipePrefix + path
	}
	timeout := 5 * time.Second
	return winio.DialPipe(path, &timeout)
}
