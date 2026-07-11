//go:build !windows

package datalink

import "syscall"

func configureSocket(network, address string, c syscall.RawConn) error {
	var socketErr error
	if err := c.Control(func(fd uintptr) {
		socketErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	}); err != nil {
		return err
	}
	return socketErr
}
