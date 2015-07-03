// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"

	"github.com/goulash/pacman"
)

// List displays all the packages available for the database.
// Note that they don't need to be registered with the database.
func List(c *Config) error {
	// Get packages in repo directory and database and find out duplicates.
	// Depinding on the configuration, some of this might be unecessary work,
	// so we have to declare the variabes here.
	var (
		db     map[string]*pacman.Package
		missed []*pacman.Package
		dups   map[string]int
		aur    map[string]*pacman.Package
	)
	pkgs, old := getRepoPkgs(c.path)
	if c.Pending {
		db, missed = getDatabasePkgs(c.Repository)
	}
	if c.Duplicates {
		dups = make(map[string]int)
		for _, p := range old {
			dups[p.Name]++
		}
	}
	if c.Synchronize {
		aur = getAURMap(mapPkgs(pkgs, pkgName))
	}

	// Create a list.
	var pkgnames []string
	if c.Pending {
		for _, p := range missed {
			pkgnames = append(pkgnames, fmt.Sprintf("-%s-", p.Name))
		}
	}
	for _, p := range pkgs {
		buf := bytes.NewBufferString(p.Name)

		if c.Pending {
			dbp, ok := db[p.Name]
			if !ok || dbp.OlderThan(p) {
				buf.WriteRune('*')
			}
		}
		if c.Versioned {
			buf.WriteRune(' ')
			buf.WriteString(p.Version)
		}
		if c.Synchronize {
			ap := aur[p.Name]
			if ap == nil {
				buf.WriteString(" <?>")
			} else if ap.NewerThan(p) {
				if c.Versioned {
					buf.WriteString(" -> ")
					buf.WriteString(ap.Version)
				} else {
					buf.WriteString(" <!>")
				}
			} else if ap.OlderThan(p) {
				if c.Versioned {
					buf.WriteString(" <- ")
					buf.WriteString(ap.Version)
				} else {
					buf.WriteString(" <*>")
				}
			}
		}
		if c.Duplicates && dups[p.Name] > 0 {
			buf.WriteString(fmt.Sprintf(" (%v)", dups[p.Name]))
		}

		pkgnames = append(pkgnames, buf.String())
	}

	// Print packages to stdout
	printSet(pkgnames, "", c.Columnate)
	return nil
}
