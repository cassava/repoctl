// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/goulash/pacman"
	"github.com/goulash/pr"
)

// printSet prints a set of items and optionally a header.
func printSet(list []string, h string, cols bool) {
	sort.Strings(list)
	if h != "" {
		fmt.Printf("\n%s\n", h)
	}
	if cols {
		pr.PrintFlex(list)
	} else if h != "" {
		for _, j := range list {
			fmt.Println(" ", j)
		}
	} else {
		for _, j := range list {
			fmt.Println(j)
		}
	}
}

// dieOnError prints error to stderr and dies if err != nil.
func dieOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		os.Exit(1)
	}
}

func pkgNameVersion(db map[string]*pacman.Package) func(*pacman.Package) string {
	return func(p *pacman.Package) string {
		dp, ok := db[p.Name]
		if ok {
			return fmt.Sprintf("%s %s -> %s", p.Name, dp.Version, p.Version)
		}
		return fmt.Sprintf("%s -> %s", p.Name, p.Version)
	}
}
