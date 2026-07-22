//go:build windows

package main

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// dialHerdr connects to the Herdr API named pipe. On Windows, Herdr's
// HERDR_SOCKET_PATH is a named-pipe path (e.g. \\.\pipe\herdr-...).
func dialHerdr(path string) (net.Conn, error) {
	timeout := 5 * time.Second
	return winio.DialPipe(path, &timeout)
}
