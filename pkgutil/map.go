// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pkgutil

import (
	"path"

	"github.com/goulash/pacman"
)

type MapFunc func(*pacman.Package) string

// Map maps pacman.Packages to some string characteristic of a pacman.Package.
//
// For example:
//
//  files := pkgutil.Map(ps, pkgutil.PkgFilename)
func Map(pkgs pacman.Packages, f MapFunc) []string {
	results := make([]string, len(pkgs))
	for i, p := range pkgs {
		results[i] = f(p)
	}
	return results
}

func MapPkg(pkgs pacman.Packages, f MapFunc) map[string]*pacman.Package {
	m := make(map[string]*pacman.Package)
	for _, p := range pkgs {
		m[f(p)] = p
	}
	return m
}

func MapBool(pkgs pacman.Packages, f MapFunc) map[string]bool {
	m := make(map[string]bool)
	for _, p := range pkgs {
		m[f(p)] = true
	}
	return m
}

// PkgFilename gets the filename of the package. There is no normalization
// done here. Comparing two pacman.Packages that refer to the same file might
// not have the same filename internally.
//
// If you need that, check out PkgBasename!
func PkgFilename(p *pacman.Package) string {
	return p.Filename
}

func PkgBasename(p *pacman.Package) string {
	return path.Base(p.Filename)
}

func PkgName(p *pacman.Package) string {
	return p.Name
}

func PkgBase(p *pacman.Package) string {
	return p.Base
}
