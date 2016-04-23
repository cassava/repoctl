// Copyright 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Copyright 2011, The Go Authors. All rights reserved.

// +build linux

package pr

import (
	"syscall"
	"unsafe"
)

// GetTerminalWidth returns the current width of the connected terminal for the
// given file descriptor. If fd is not connected to a terminal, then -1 is
// returned.
//
// Note: this only works on Linux.
func GetTerminalWidth(fd int) int {
	var dimensions [4]uint16

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		return -1
	}

	return int(dimensions[1])
}

// StdoutTerminalWidth returns the current width of the terminal (if any)
// connected to Stdout.
//
// Note: this only works on Linux.
func StdoutTerminalWidth() int {
	return GetTerminalWidth(syscall.Stdout)
}

// StderrTerminalWidth returns the current width of the terminal (if any)
// connected to Stderr.
//
// Note: this only works on Linux.
func StderrTerminalWidth() int {
	return GetTerminalWidth(syscall.Stderr)
}
