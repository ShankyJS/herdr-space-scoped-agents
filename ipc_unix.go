//go:build !windows

package main

import "net"

// dialHerdr connects to the Herdr API unix socket.
func dialHerdr(path string) (net.Conn, error) {
	return net.Dial("unix", path)
}
