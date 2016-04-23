// Copyright (c) 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// This example program imitates the command
//	\ls -AU
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/goulash/pr"
)

func main() {
	width := flag.Int("width", -1, "width of the terminal")
	sortme := flag.Bool("sort", false, "sort the items")
	flag.Parse()

	var args []string
	if flag.NArg() > 0 {
		args = flag.Args()
	} else {
		args = []string{"."}
	}

	for _, path := range args {
		func() {
			file, err := os.Open(path)
			if err != nil {
				fmt.Errorf("Error: %s\n", err)
				return
			}
			defer file.Close()

			stat, err := file.Stat()
			if err != nil {
				fmt.Errorf("Error: %s\n", err)
				return
			}

			var slice []string
			if stat.IsDir() {
				slice, err = file.Readdirnames(0)
				if err != nil {
					fmt.Errorf("Error: %s\n", err)
					return
				}
				if *sortme {
					ss := sort.StringSlice(slice)
					ss.Sort()
					slice = []string(ss)
				}
			} else {
				slice = []string{stat.Name()}
			}

			if *width <= 0 {
				pr.PrintFlex(slice)
			} else {
				pr.FprintFlex(os.Stdout, *width, slice)
			}
		}()
	}
}
