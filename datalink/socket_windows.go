//go:build windows

package datalink

import "syscall"

// Windows does not need SO_REUSEADDR for a single BACnet listener. Avoiding
// it also prevents WSAEINVAL on systems where the option is rejected while
// the socket is being created.
func configureSocket(network, address string, c syscall.RawConn) error {
	return nil
}
