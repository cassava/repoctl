// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import "path"

// Packages is merely a list of packages with support for some
// set and list functions.
type Packages []*Package

type MapFunc func(*Package) string

// Map maps Packages to some string characteristic of a Package.
func (pkgs Packages) Map(f MapFunc) []string {
	results := make([]string, len(pkgs))
	for i, p := range pkgs {
		results[i] = f(p)
	}
	return results
}

func nkgFilename(p *Package) string {
	return p.Filename
}

func PkgBasename(p *Package) string {
	return path.Base(p.Filename)
}

func PkgName(p *Package) string {
	return p.Name
}

// FilterFunc is a function that given a package, returns true if the package
// is ok, and false if it should not be included (filtered out).
//
// It is used with the various Filter* functions to turn one set of Packages
// into another.
//
// FilterFuncs can be combined and negated. The idea is that you implement
// your own filter functions.
type FilterFunc func(p *Package) bool

// And performs a logical AND of f and fs. True is returned only iff
// all the filter functions in fs and f return true.
func (f FilterFunc) And(fs ...FilterFunc) FilterFunc {
	return func(p *Package) bool {
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
	return func(p *Package) bool {
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
	return func(p *Package) bool {
		return !f(p)
	}
}

// Filter filters a set of packages with some filter functions.
func Filter(pkgs Packages, f FilterFunc) Packages {
	fps := make(Packages, 0, len(pkgs))
	for _, p := range pkgs {
		if f(p) {
			fps = append(fps, p)
		}
	}
	return fps
}

// FilterAll filters the packages through all of the filter functions.
func FilterAll(pkgs Packages, fs []FilterFunc) Packages {
	fps := make(Packages, 0, len(pkgs))
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
func FilterAny(pkgs Packages, fs []FilterFunc) Packages {
	fps := make(Packages, 0, len(pkgs))
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

// Intersect returns the intersection between two sets of Packages.
func Intersect(ps1, ps2 Packages) Packages {
	fps := make(Packages, 0, min(len(ps1), len(ps2)))
	s := make(set)
	s.InsertAll(ps1)
	for _, p := range ps2 {
		if s.Contains(p) {
			fps = append(fps, p)
		}
	}
	return fps
}

// Union returns the union of two sets of Packages.
func Union(ps1, ps2 Packages) Packages {
	s := make(set)
	s.InsertAll(ps1)
	s.InsertAll(ps2)
	return s.All()
}

// Subtract returns ps1 minus all packages that are found in ps2.
func Subtract(ps1, ps2 Packages) Packages {
	s := make(set)
	s.InsertAll(ps1)
	for _, p := range ps2 {
		s.Remove(p)
	}
	return s.All()
}

type set map[string]Packages

func (s set) All() Packages {
	fps := make(Packages, 0, len(s))
	for _, pkgs := range s {
		for _, p := range pkgs {
			fps = append(fps, p)
		}
	}
	return fps
}

func (s set) InsertAll(pkgs Packages) {
	for _, p := range pkgs {
		s.Insert(p)
	}
}

func (s set) Insert(p *Package) {
	m := s[p.Name]
	if m == nil {
		s[p.Name] = Packages{p}
	} else {
		for _, v := range m {
			if p.Equal(v) {
				return
			}
		}
		s[p.Name] = append(m, p)
	}
}

func (s set) Contains(p *Package) bool {
	m := s[p.Name]
	if m == nil {
		return false
	}
	for _, v := range m {
		if p.Equal(v) {
			return true
		}
	}
	return false
}

func (s set) Remove(p *Package) {
	m := s[p.Name]
	if len(m) <= 1 {
		delete(s, p.Name)
	} else {
		i := -1
		for j, v := range m {
			if p.Equal(v) {
				i = j
				break
			}
		}
		if i != -1 {
			if len(s) == i+1 {
				s[p.Name] = m[:i]
			} else {
				s[p.Name] = append(m[:i], m[i+1:]...)
			}
		}
	}
}
