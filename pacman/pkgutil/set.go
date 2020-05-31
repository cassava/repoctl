// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pkgutil

import "github.com/goulash/pacman"

// Intersect returns the intersection between two sets of pacman.Packages.
func Intersect(ps1, ps2 pacman.Packages) pacman.Packages {
	fps := make(pacman.Packages, 0, min(len(ps1), len(ps2)))
	s := NewSet()
	s.InsertAll(ps1)
	for _, p := range ps2 {
		if s.Contains(p) {
			fps = append(fps, p)
		}
	}
	return fps
}

// Union returns the union of two sets of pacman.Packages.
func Union(ps1, ps2 pacman.Packages) pacman.Packages {
	s := NewSet()
	s.InsertAll(ps1)
	s.InsertAll(ps2)
	return s.All()
}

// Subtract returns ps1 minus all packages that are found in ps2.
func Subtract(ps1, ps2 pacman.Packages) pacman.Packages {
	s := NewSet()
	s.InsertAll(ps1)
	for _, p := range ps2 {
		s.Remove(p)
	}
	return s.All()
}

type Set map[string]pacman.Packages

func NewSet() Set {
	return make(Set)
}

func (s Set) All() pacman.Packages {
	fps := make(pacman.Packages, 0, len(s))
	for _, pkgs := range s {
		for _, p := range pkgs {
			fps = append(fps, p)
		}
	}
	return fps
}

func (s Set) InsertAll(pkgs pacman.Packages) {
	for _, p := range pkgs {
		s.Insert(p)
	}
}

func (s Set) Insert(p *pacman.Package) {
	m := s[p.Name]
	if m == nil {
		s[p.Name] = pacman.Packages{p}
	} else {
		for _, v := range m {
			if p.Equals(v) {
				return
			}
		}
		s[p.Name] = append(m, p)
	}
}

func (s Set) Contains(p *pacman.Package) bool {
	m := s[p.Name]
	if m == nil {
		return false
	}
	for _, v := range m {
		if p.Equals(v) {
			return true
		}
	}
	return false
}

func (s Set) Remove(p *pacman.Package) {
	m := s[p.Name]
	if len(m) <= 1 {
		delete(s, p.Name)
	} else {
		i := -1
		for j, v := range m {
			if p.Equals(v) {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
