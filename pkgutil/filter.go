// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pkgutil

import "github.com/goulash/pacman"

// FilterFunc is a function that given a package, returns true if the package
// is ok, and false if it should not be included (filtered out).
//
// It is used with the various Filter* functions to turn one set of pacman.Packages
// into another.
//
// FilterFuncs can be combined and negated. The idea is that you implement
// your own filter functions.
type FilterFunc func(p *pacman.Package) bool

// And performs a logical AND of f and fs. True is returned only iff
// all the filter functions in fs and f return true.
func (f FilterFunc) And(fs ...FilterFunc) FilterFunc {
	return func(p *pacman.Package) bool {
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
	return func(p *pacman.Package) bool {
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
	return func(p *pacman.Package) bool {
		return !f(p)
	}
}

// Filter filters a set of packages with some filter functions.
func Filter(pkgs pacman.Packages, f FilterFunc) pacman.Packages {
	fps := make(pacman.Packages, 0, len(pkgs))
	for _, p := range pkgs {
		if f(p) {
			fps = append(fps, p)
		}
	}
	return fps
}

// FilterAll filters the packages through all of the filter functions.
func FilterAll(pkgs pacman.Packages, fs []FilterFunc) pacman.Packages {
	fps := make(pacman.Packages, 0, len(pkgs))
	for _, p := range pkgs {
		keep := true
		for _, f := range fs {
			if !f(p) {
				keep = false
				break
			}
		}
		if keep {
			fps = append(fps, p)
		}
	}
	return fps
}

// FilterAny filters the packages through the filters in fs,
// where at least one must return true for it to be included.
func FilterAny(pkgs pacman.Packages, fs []FilterFunc) pacman.Packages {
	fps := make(pacman.Packages, 0, len(pkgs))
	for _, p := range pkgs {
		for _, f := range fs {
			if f(p) {
				fps = append(fps, p)
				break
			}
		}
	}
	return fps
}
