// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package pacman provides routines for dealing with Pacman packages.
package pacman

import (
	"fmt"
	"io"
)

// DebugWriter is used to write debugging information from this module.
// If it is nil, then no debugging messages are printed.
var DebugWriter io.Writer = nil

func debugf(format string, obj ...interface{}) {
	if DebugWriter != nil {
		fmt.Fprintf(DebugWriter, format, obj...)
	}
}
