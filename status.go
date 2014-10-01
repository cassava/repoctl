// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/goulash/pacman"
)

// Status prints a thorough status of the current repository.
//
// Filters available are:
// 	duplicates      files to be deleted or backed up
// 	pending         packages to be added/removed from database
// 	installed       packages locally installed
// 	outdated        packages with newer versions in AUR
// 	missing         packages not found in AUR
func Status(c *Config) error {
	if len(c.Args) > 0 {
		return filter(c)
	}
	return status(c)
}

func filter(c *Config) error {
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

func status(c *Config) error {
	pkgs, outdated := getRepoPkgs(c.path)
	db, missed := getDatabasePkgs(c.Repository)

	name := c.database[:strings.IndexByte(c.database, '.')]
	fmt.Printf("On repo %s\n", name)

	// We assume that there is nothing to do, and if there is,
	// then this is set to false.
	var nothing = true

	if len(outdated) > 0 {
		printSet(mapPkgs(outdated, pkgFilename), "Obsolete packages to be removed/backed up:", c.Columnate)
		nothing = false
	}

	if len(missed) > 0 {
		printSet(missed, "Database entries pending removal:", c.Columnate)
		nothing = false
	}

	pending := filterPkgs(pkgs, pendingFilter(db))
	if len(pending) > 0 {
		printSet(mapPkgs(pending, pkgName), "Database entries pending addition:", c.Columnate)
		nothing = false
	}

	aur, unavailable := getAURPkgs(mapPkgs(pkgs, pkgName))
	if len(unavailable) > 0 {
		printSet(unavailable, "Packages unavailable in AUR:", c.Columnate)
		nothing = false
	}

	updates := filterPkgs(pkgs, outdatedFilter(aur))
	if len(updates) > 0 {
		printSet(mapPkgs(updates, pkgName), "Packages with updates in AUR:", c.Columnate)
		nothing = false
	}

	if nothing {
		fmt.Println("Everything up-to-date.")
	}
	return nil
}
