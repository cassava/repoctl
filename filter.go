// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// This file contains filtering logic, not only for the action Filter,
// but also for other routines.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/goulash/pacman"
)

// Filter prints package names that are filtered by the specified filters.
func Filter(c *Config) error {
	pkgs, outdated := getRepoPkgs(c.path)

	var (
		readDB  bool
		readAUR bool

		db          map[string]*pacman.Package
		aur         map[string]*pacman.Package
		unavailable []string

		getDB = func() map[string]*pacman.Package {
			if !readDB {
				db, _ = getDatabasePkgs(c.Repository)
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

nextFilter:
	for _, fltr := range c.Args {
		var negate bool
		if strings.HasPrefix(fltr, "!") {
			fltr = fltr[1:]
			negate = true
		}

		var ff filterFunc
		switch fltr {
		case "pending":
			ff = pendingFilter(getDB())
		case "duplicates":
			ff = intersectsListFilter(mapPkgs(outdated, pkgFilename))
		case "outdated":
			ff = outdatedFilter(getAUR())
		case "missing":
			ff = intersectsListFilter(getAURUnavailable())
		case "local":
			fmt.Fprintln(os.Stderr, `Error: filter "installed" has not been implemented yet!`)
			continue nextFilter
		default:
			fmt.Fprintf(os.Stderr, "Error: unknown filter %q!", ff)
			continue nextFilter
		}
		if negate {
			ff = negateFilter(ff)
		}

		pkgs = filterPkgs(pkgs, ff)
	}

	printSet(mapPkgs(pkgs, pkgName), "", c.Columnate)
	return nil
}

type filterFunc func(*pacman.Package) bool

func filterPkgs(pkgs []*pacman.Package, f filterFunc) []*pacman.Package {
	filtered := make([]*pacman.Package, 0, len(pkgs))
	for _, p := range pkgs {
		if f(p) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func negateFilter(f filterFunc) filterFunc {
	return func(pkg *pacman.Package) bool {
		return !f(pkg)
	}
}

func intersectsFilter(set map[string]bool) filterFunc {
	return func(pkg *pacman.Package) bool {
		return set[pkg.Name]
	}
}

func intersectsListFilter(list []string) filterFunc {
	if len(list) < 3 {
		return func(pkg *pacman.Package) bool {
			for _, p := range list {
				if pkg.Name == p {
					return true
				}
			}
			return false
		}
	}

	set := make(map[string]bool)
	for _, p := range list {
		set[p] = true
	}
	return intersectsFilter(set)
}

func pendingFilter(db map[string]*pacman.Package) filterFunc {
	return func(pkg *pacman.Package) bool {
		dbp, ok := db[pkg.Name]
		if !ok || dbp.OlderThan(pkg) {
			return true
		}
		return false
	}
}

func outdatedFilter(aur map[string]*pacman.Package) filterFunc {
	return func(pkg *pacman.Package) bool {
		ap, ok := aur[pkg.Name]
		if ok && ap.NewerThan(pkg) {
			return true
		}
		return false
	}
}
