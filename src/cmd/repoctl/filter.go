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

	// gb packages
	"shortry"

	"github.com/goulash/pacman"
)

func FilterUsage() {
	fmt.Println("repoctl filter <criteria...>")
	fmt.Println(`
Filter packages through a set of criteria combined in an AND fashion,
which can be prefixed with an exclamation mark to negate the effect.

It is only necessary to provide enough characters so that the identifier
is unambiguous; e.g. "aur.newer" can be "a.newer", "a.n", or "a".
Omissions occur can occur in hierarchical fashion.

Filters available are:

    db.missing          packages to be removed from the database
    db.pending          packages to be added to the database
    file.dupes          packages with files to be deleted or backed up
    aur.newer           packages with newer versions in AUR
    aur.missing         packages not found in AUR
    aur.older           packages with older versions in AUR
    local.installed     packages that are installed on localhost
    local.upgradable    packages that can be upgraded on localhost
`)
}

var shor = shortry.New(map[string]interface{}{
	"db.missing":        nil,
	"db.pending":        nil,
	"file.dupes":        nil,
	"aur.newer":         nil,
	"aur.missing":       nil,
	"aur.older":         nil,
	"local.installed":   nil,
	"local.upgradable":  nil,
	"local.upgradeable": nil,
})

func filterDie(msg string) {
	fmt.Fprintln(os.Stderr, msg, "\n")
	FilterUsage()
	os.Exit(1)
}

// Filter prints package names that are filtered by the specified filters.
func Filter(c *Config) error {
	// TODO: pkgs only contains real packages. This is a problem later
	// when for example db.pending does not show missing packages.
	pkgs, outdated := getRepoPkgs(c.path)

	var (
		readDB  bool
		readAUR bool
		mergeDB bool

		db          map[string]*pacman.Package
		missed      []*pacman.Package
		aur         map[string]*pacman.Package
		unavailable []string

		getDB = func() map[string]*pacman.Package {
			if !readDB {
				db, missed = getDatabasePkgs(c.Repository)
				readDB = true
			}
			return db
		}

		getMissing = func() []*pacman.Package {
			if !readDB {
				db, missed = getDatabasePkgs(c.Repository)
				readDB = true
			}
			return missed
		}

		getAUR = func() map[string]*pacman.Package {
			if !readAUR {
				nps := removeIgnored(c, pkgs)
				aur, unavailable = getAURPkgs(mapPkgs(nps, pkgName))
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

		f := shor.Matches(fltr)
		if len(f) == 0 {
			filterDie(fmt.Sprintf("Error: unknown filter %q", fltr))
		} else if len(f) > 1 {
			filterDie(fmt.Sprintf("Error: ambiguous filter %q matches %v", fltr, f))
		} else {
			fltr = f[0]
		}

		var ff filterFunc
		switch fltr {
		case "db.missing":
			// This is a special case. Normally we don't print packages
			// that are in the database but do not exist. Sometimes we
			// want to see them though.
			if negate || mergeDB {
				continue nextFilter
			}
			m := getMissing()
			pkgs = make([]*pacman.Package, len(m))
			copy(pkgs, m)
			continue
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
			filterDie(`Error: filter "local.installed" is not implemented!`)
		case "local.upgradable", "local.upgradeable":
			filterDie(`Error: filter "local.upgradable" is not implemented!`)
		default:
			filterDie(fmt.Sprintf("Error: unknown filter %q", fltr))
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

func removeIgnored(c *Config, pkgs []*pacman.Package) []*pacman.Package {
	n := len(pkgs) - len(c.IgnoreAUR)
	if n < 0 {
		n = 0
	}
	nps := make([]*pacman.Package, 0, n)
	for _, p := range pkgs {
		if !c.IgnoreAUR[p.Name] {
			nps = append(nps, p)
		}
	}
	return nps
}
