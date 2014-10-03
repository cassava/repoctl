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

func FilterUsage() {
	fmt.Println("repoctl filter <criteria...>")
	fmt.Println(`
Filter packages through a set of criteria combined in an AND fashion,
which can be prefixed with an exclamation mark to negate the effect.

Each filter belongs to a category, such as "aur", which can be omited
if the identifier is unambiguous.

It is only necessary to provide enough characters so that the identifier
is unambiguous; e.g. "aur.newer" can be "a.newer", "a.n", "newer", or "n".
(Not implemented yet.)

Filters available are:

    db.pending          packages to be added/removed from database
    file.dupes          packages with files to be deleted or backed up
    aur.newer           packages with newer versions in AUR
    aur.missing         packages not found in AUR
    aur.older           packages with older versions in AUR
    local.installed     packages that are installed on localhost
    local.upgradable    packages that can be upgraded on localhost
`)
}

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

	if len(c.Args) == 0 {
		FilterUsage()
		return nil
	}

nextFilter:
	for _, fltr := range c.Args {
		var negate bool
		if strings.HasPrefix(fltr, "!") {
			fltr = fltr[1:]
			negate = true
		}

		var ff filterFunc
		switch fltr {
		case "db.pending":
			ff = dbPendingFilter(getDB())
		case "file.dupes":
			ff = intersectsListFilter(mapPkgs(outdated, pkgFilename))
		case "aur.newer":
			ff = aurNewerFilter(getAUR())
		case "aur.older":
			ff = aurOlderFilter(getAUR())
		case "aur.missing":
			ff = intersectsListFilter(getAURUnavailable())
		case "local.installed":
			fmt.Fprintln(os.Stderr, `Error: filter "local.installed" is not implemented!`)
			continue nextFilter
		case "local.upgradable", "local.upgradeable":
			fmt.Fprintln(os.Stderr, `Error: filter "local.upgradable" is not implemented!`)
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

func dbPendingFilter(db map[string]*pacman.Package) filterFunc {
	return func(pkg *pacman.Package) bool {
		dbp, ok := db[pkg.Name]
		if !ok || dbp.OlderThan(pkg) {
			return true
		}
		return false
	}
}

func aurNewerFilter(aur map[string]*pacman.Package) filterFunc {
	return func(pkg *pacman.Package) bool {
		ap, ok := aur[pkg.Name]
		if ok && ap.NewerThan(pkg) {
			return true
		}
		return false
	}
}

func aurOlderFilter(aur map[string]*pacman.Package) filterFunc {
	return func(pkg *pacman.Package) bool {
		ap, ok := aur[pkg.Name]
		if ok && ap.OlderThan(pkg) {
			return true
		}
		return false
	}
}
