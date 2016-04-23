// Copyright (c) 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// +build linux

package pr

import (
	"os"
)

func PrintGrid(columns int, list []string) {
	tw := StdoutTerminalWidth()
	FprintGrid(os.Stdout, tw, columns, list)
}

func PrintFlex(list []string) {
	tw := StdoutTerminalWidth()
	FprintFlex(os.Stdout, tw, list)
}
