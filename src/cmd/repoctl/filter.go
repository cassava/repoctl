// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// This file contains filtering logic, not only for the action Filter,
// but also for other routines.

package main

import (
	"strings"

	// gb packages
	"shortry"

	"github.com/goulash/pacman"
	"github.com/spf13/cobra"
)

var unknownFilterError = ErrorTmpl{
	Short: `query {{sprintf "%q" .}} does not exist`,
	Long: `You have specified a filter that does not exist.
Run repoctl help filter for more information on which filters exist.`,
}

var ambiguousFilterError = ErrorTmpl{
	Short: `query {{sprintf "%q" .Filter}} matches multiple filters`,
	Long: `You have used a query that matches multiple filters:{{range .Matches}}
  {{.}}
{{end}}
Please make your query unambiguous.`,
}

var FilterCmd = &cobra.Command{
	Use:   "filter <criteria...>",
	Short: "filter packages by one or more criteria",
	Long: `Filter packages through a set of criteria.
    
  The criteria are combined in an AND fashion, and each filter can be
  prefixed with an exclamation mark "!" to negate the effect of the
  filter.

  It is only necessary to provide enough characters so that the
  identifier of a filter is unambiguous; e.g. "aur.newer" can be
  "a.newer" or "a.n". Omissions can occur in hierarchical fashion: "d"
  means "db.missing" and "db.pending" (most of the time this makes no
  sense however).

  Filters available are:

    db.missing          packages to be removed from the database
    db.pending          packages to be added to the database
    file.dupes          packages with files to be deleted or backed up
    aur.newer           packages with newer versions in AUR
    aur.missing         packages not found in AUR
    aur.older           packages with older versions in AUR
    local.installed     packages that are installed on localhost
    local.upgradable    packages that can be upgraded on localhost
`,
	Example: `  repoctl filter aur.newer
  repoctl filter aur.n file.d
  repoctl filter a.n !f
  repoctl filter a.n | cower -d`,
	Run: func(cmd *cobra.Command, args []string) {
		err := Filter(args)
		dieOnError(err)
	},
}

// Filter prints package names that are filtered by the specified filters.
func Filter(args []string) error {
	pkgs, outdated := getRepoPkgs(Conf.repodir)

	// This function looks huge, but the actual body is pretty small.
	// We define a lot of anonymous functions to do the work for us.
	//
	// getDB, getMissing, getAUR, and getAURUnavailable reduce the number
	// of times packages are read to once. The functions essentially cache
	// the result and return the cached result if it exists.
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
				db, missed = getDatabasePkgs(Conf.Repository)
				readDB = true
			}
			return db
		}

		getMissing = func() []*pacman.Package {
			if !readDB {
				db, missed = getDatabasePkgs(Conf.Repository)
				readDB = true
			}
			return missed
		}

		getAUR = func() map[string]*pacman.Package {
			if !readAUR {
				nps := removeIgnored(pkgs)
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

		// shor finds a filterFunc that will filter the packages in the correct way.
		// The package shortry lets us match "d.m" to "db.missing" for example.
		shor = shortry.New(map[string]interface{}{
			"db.missing": func(negate bool) filterFunc {
				// This is a special case. Normally we don't print packages
				// that are in the database but do not exist. Sometimes we
				// want to see them though.
				if !negate && !mergeDB {
					m := getMissing()
					pkgs = make([]*pacman.Package, len(m))
					copy(pkgs, m)
				}
				return nil
			},
			"db.pending": func(_ bool) filterFunc {
				return dbPendingFilter(getDB())
			},
			"file.dupes": func(_ bool) filterFunc {
				return intersectsListFilter(mapPkgs(outdated, pkgFilename))
			},
			"aur.newer": func(_ bool) filterFunc {
				return aurNewerFilter(getAUR())
			},
			"aur.missing": func(_ bool) filterFunc {
				return intersectsListFilter(getAURUnavailable())
			},
			"aur.older": func(_ bool) filterFunc {
				return aurOlderFilter(getAUR())
			},
			"local.installed": func(_ bool) filterFunc {
				filterDie(`Error: filter "local.installed" is not implemented!`)
				return nil
			},
			"local.upgradable": func(_ bool) filterFunc {
				filterDie(`Error: filter "local.upgradable" is not implemented!`)
				return nil
			},
		})
	)

	for _, fltr := range args {
		var negate bool
		if strings.HasPrefix(fltr, "!") {
			fltr = fltr[1:]
			negate = true
		}

		f, err := shor.Get(fltr)
		if err != nil {
			if err == shortry.ErrNotExists {
				return unknownFilterError.Instatiate(err, fltr)
			} else if err == shortry.ErrAmbiguous {
				return ambiguousFilterError.Instantiate(err, struct {
					Query   string
					Matches []string
				}{fltr, shor.Matches(fltr)})
			} else {
				panic("unknown error: " + err.Error())
			}
		}

		ff := f.(func(bool) filterFunc)(negate)
		if ff == nil {
			continue
		}
		if negate {
			ff = negateFilter(ff)
		}
		pkgs = filterPkgs(pkgs, ff)
	}

	printSet(mapPkgs(pkgs, pkgName), "", Conf.Columnate)
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

func removeIgnored(pkgs []*pacman.Package) []*pacman.Package {
	n := len(pkgs) - len(Conf.IgnoreAUR)
	if n < 0 {
		n = 0
	}

	aur := make(map[string]bool)
	for _, k := range Conf.IgnoreAUR {
		aur[k] = true
	}

	nps := make([]*pacman.Package, 0, n)
	for _, p := range pkgs {
		if !aur[p.Name] {
			nps = append(nps, p)
		}
	}
	return nps
}
