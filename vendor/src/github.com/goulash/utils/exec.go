// Copyright (c) 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"log"
	"os/exec"
	"strings"
)

// RunCmd runs the command and returns whether it completed correctly or not.
// If any error occurs, then the entire output is printed to the log file.
func RunCmd(cmd *exec.Cmd) bool {
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error running '%s':\n%s\n", strings.Join(cmd.Args, " "), string(output))
		return false
	}
	return true
}

// CombineCmdArgs puts the given arguments all in a slice. If one of the
// arguments is a string, then it is split by spaces, and if it is a slice,
// then the elements are added to the slice. This is all returned.
//
// Warning: this is a very obscure and weird function
// (it's already bitten me once), so be careful.
func CombineCmdArgs(args ...interface{}) []string {
	output := make([]string, 0, 24) // 24 is just a guess
	for _, any := range args {
		switch v := any.(type) {
		case string:
			for _, i := range strings.Split(v, " ") {
				output = append(output, i)
			}
		case []string:
			for _, i := range v {
				output = append(output, i)
			}
		default:
			panic("invalid type! want string or []string")
		}
	}
	return output
}
