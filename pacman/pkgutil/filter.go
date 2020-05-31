// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pkgutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
	"github.com/goulash/pacman/alpm"
	"github.com/goulash/pacman/aur"
	"github.com/goulash/pacman/meta"
)

// FilterFunc is a function that given a package, returns true if the package
// is ok, and false if it should not be included (filtered out).
//
// It is used with the various Filter* functions to turn one set of pacman.Packages
// into another.
//
// FilterFuncs can be combined and negated. The idea is that you implement
// your own filter functions.
type FilterFunc func(pacman.AnyPackage) bool

// And performs a logical AND of f and fs. True is returned only iff
// all the filter functions in fs and f return true.
func (f FilterFunc) And(fs ...FilterFunc) FilterFunc {
	return func(p pacman.AnyPackage) bool {
		for _, f := range fs {
			if !f(p) {
				return false
			}
		}
		return f(p)
	}
}

// Or performs a logical OR of f and fs. True is returned as soon
// as any of the filter functions in fs and f return true.
func (f FilterFunc) Or(fs ...FilterFunc) FilterFunc {
	return func(p pacman.AnyPackage) bool {
		for _, f := range fs {
			if f(p) {
				return true
			}
		}
		return f(p)
	}
}

// Not negates the effect of the filter. Therefore true becomes false
// and false becomes true.
func (f FilterFunc) Not() FilterFunc {
	return func(p pacman.AnyPackage) bool {
		return !f(p)
	}
}

// Filter filters a set of packages with some filter functions.
//
// The idea is that you can do this:
//
//  pkgs := pkgutil.FilterAll(pkgs, func(ap AnyPackage) bool {
//      p := ap.(*meta.Package)
//      p :=
func Filter(pkgs pacman.AnyPackages, f FilterFunc) pacman.AnyPackages {
	switch ps := pkgs.(type) {
	case pacman.Packages:
		return filterPacmanPkgs(ps, f)
	case meta.Packages:
		return filterMetaPkgs(ps, f)
	case aur.Packages:
		return filterAURPkgs(ps, f)
	default:
		panic("unexpected")
	}
}

func filterPacmanPkgs(pkgs pacman.Packages, f FilterFunc) pacman.Packages {
	fps := make(pacman.Packages, 0, len(pkgs))
	for _, p := range pkgs {
		if f(p) {
			fps = append(fps, p)
		}
	}
	return fps
}

func filterAURPkgs(pkgs aur.Packages, f FilterFunc) aur.Packages {
	fps := make(aur.Packages, 0, len(pkgs))
	for _, p := range pkgs {
		if f(p) {
			fps = append(fps, p)
		}
	}
	return fps
}

func filterMetaPkgs(pkgs meta.Packages, f FilterFunc) meta.Packages {
	fps := make(meta.Packages, 0, len(pkgs))
	for _, p := range pkgs {
		if f(p) {
			fps = append(fps, p)
		}
	}
	return fps
}

// FilterAll filters the packages through all of the filter functions.
func FilterAll(pkgs pacman.AnyPackages, fs ...FilterFunc) pacman.AnyPackages {
	return Filter(pkgs, func(p pacman.AnyPackage) bool {
		for _, f := range fs {
			if !f(p) {
				return false
			}
		}
		return true
	})
}

// FilterAny filters the packages through the filters in fs,
// where at least one must return true for it to be included.
func FilterAny(pkgs pacman.AnyPackages, fs ...FilterFunc) pacman.AnyPackages {
	return Filter(pkgs, func(p pacman.AnyPackage) bool {
		for _, f := range fs {
			if f(p) {
				return true
			}
		}
		return false
	})
}

// FilterNewest returns the newest of the given packages.
func FilterNewest(pkgs pacman.AnyPackages) pacman.AnyPackages {
	m := make(map[string]pacman.AnyPackage)
	pkgs.Iterate(func(p pacman.AnyPackage) {
		if pacman.PkgNewer(p, m[p.PkgName()]) {
			m[p.PkgName()] = p
		}
	})

	return Filter(pkgs, func(p pacman.AnyPackage) bool {
		return alpm.VerCmp(p.PkgVersion(), m[p.PkgName()].PkgVersion()) == 0
	})
}

// NewerFltr passes all packages through that are really newer than the
// packages given.
func NewerFltr(pkgs pacman.AnyPackages) FilterFunc {
	m := make(map[string]pacman.AnyPackage)
	pkgs.Iterate(func(p pacman.AnyPackage) {
		if pacman.PkgNewer(p, m[p.PkgName()]) {
			m[p.PkgName()] = p
		}
	})

	return func(p pacman.AnyPackage) bool {
		return pacman.PkgNewer(p, m[p.PkgName()])
	}
}

// NewestFltr passes all packages through that are at least as new as the
// packages given; this is a superset of NewerFltr.
func NewestFltr(pkgs pacman.AnyPackages) FilterFunc {
	m := make(map[string]pacman.AnyPackage)
	pkgs.Iterate(func(p pacman.AnyPackage) {
		if pacman.PkgNewer(p, m[p.PkgName()]) {
			m[p.PkgName()] = p
		}
	})

	return func(p pacman.AnyPackage) bool {
		return !pacman.PkgOlder(p, m[p.PkgName()])
	}
}

// WordFltr passes all packages through that contain the given word.
func WordFltr(word string, mf MapFunc) FilterFunc {
	return func(p pacman.AnyPackage) bool {
		return strings.Contains(mf(p), word)
	}
}

// RegexFltr passes all packages through that match the regular expression.
func RegexFltr(regex string, mf MapFunc) (FilterFunc, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	return func(p pacman.AnyPackage) bool {
		return r.MatchString(mf(p))
	}, nil
}

// MustRegexFltr is the same as RegexFltr, except that it quits the program
// if regular expression is invalid.
func MustRegexFltr(regex string, mf MapFunc) FilterFunc {
	ff, err := RegexFltr(regex, mf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid regular expression:", err)
		os.Exit(1)
	}
	return ff
}

// GlobFltr passes all packages through that match the glob pattern.
func GlobFltr(glob string, mf MapFunc) FilterFunc {
	return func(p pacman.AnyPackage) bool {
		matched, err := filepath.Match(glob, mf(p))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Glob pattern malformed:", err)
		}
		return matched
	}
}

// MissingFltr passes all packages through that do not exist in the filesystem.
func MissingFltr() FilterFunc {
	checkExistence := func(s string) bool {
		if s == "" {
			return true
		}
		ex, err := osutil.FileExists(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %s", s, err)
		}
		return !ex
	}

	return func(p pacman.AnyPackage) bool {
		switch p := p.(type) {
		case *pacman.Package:
			if p.Origin == pacman.FileOrigin {
				return false
			}
			return checkExistence(p.Filename)
		case *aur.Package:
			return false
		case *meta.Package:
			return p.HasFiles()
		default:
			return checkExistence(p.Pkg().Filename)
		}
	}
}

// NameFltr passes all packages through that have one of the names.
func NameFltr(names []string) FilterFunc {
	m := make(map[string]bool)
	for _, n := range names {
		m[n] = true
	}

	return func(p pacman.AnyPackage) bool {
		return m[p.PkgName()]
	}
}
