// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/goulash/pacman"
)

// Filter prints package names that are filtered by the specified filters.
func Filter(c *Config) error {
	pkgs, outdated := getRepoPkgs(c.path)

	var (
		readDB  bool
		readAUR bool

		db          map[string]*pacman.Package
		missed      []string
		aur         map[string]*pacman.Package
		unavailable []string

		getDB = func() map[string]*pacman.Package {
			if !readDB {
				db, missed = getDatabasePkgs(c.Repository)
				readDB = true
			}
			return db
		}

		getAUR = func() map[string]*pacman.Package {
			if !readAUR {
				aur, unavailable = getAURPkgs(mapPkgs(pkgs, pkgName))
				readAUR = true
			}
			return aur
		}

		getAURUnavailable = func() []string {
			if !readAUR {
				getAUR()
			}
			return unavailable
		}
	)

	for _, f := range c.Args {
		switch f {
		case "duplicates":
			pkgs = filterPkgs(pkgs, intersectsListFilter(mapPkgs(outdated, pkgFilename)))
		case "pending":
			pkgs = filterPkgs(pkgs, pendingFilter(getDB()))
		case "installed":
			fmt.Fprintln(os.Stderr, "filter installed has not been implemented yet!")
		case "outdated":
			pkgs = filterPkgs(pkgs, outdatedFilter(getAUR()))
		case "missing":
			pkgs = filterPkgs(pkgs, intersectsListFilter(getAURUnavailable()))
		default:
			fmt.Fprintf(os.Stderr, "Unknown filter '%s'!", f)
		}
	}

	printSet(mapPkgs(pkgs, pkgName), "", c.Columnate)
	return nil
}
