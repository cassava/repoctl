// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package pkgutil provides useful functions for working with packages.
package pkgutil

import (
	"bytes"
	"path"

	"github.com/goulash/pacman"
)

type MapFunc func(pacman.AnyPackage) string

// Map maps pacman.Packages to some string characteristic of a pacman.Package.
//
// For example:
//
//  files := pkgutil.Map(ps, pkgutil.PkgFilename)
func Map(pkgs pacman.AnyPackages, f MapFunc) []string {
	results := make([]string, pkgs.Len())
	i := 0
	pkgs.Iterate(func(p pacman.AnyPackage) {
		results[i] = f(p)
		i++
	})
	return results
}

func MapPkg(pkgs pacman.AnyPackages, f MapFunc) map[string]*pacman.Package {
	m := make(map[string]*pacman.Package)
	pkgs.Iterate(func(p pacman.AnyPackage) {
		m[f(p)] = p.Pkg()
	})
	return m
}

func MapAny(pkgs pacman.AnyPackages, f MapFunc) map[string]pacman.AnyPackage {
	m := make(map[string]pacman.AnyPackage)
	pkgs.Iterate(func(p pacman.AnyPackage) {
		m[f(p)] = p
	})
	return m
}

func MapBool(pkgs pacman.AnyPackages, f MapFunc) map[string]bool {
	m := make(map[string]bool)
	pkgs.Iterate(func(p pacman.AnyPackage) {
		m[f(p)] = true
	})
	return m
}

func PkgName(ap pacman.AnyPackage) string {
	return ap.PkgName()
}

func PkgBase(ap pacman.AnyPackage) string {
	return ap.Pkg().Base
}

// PkgFilename gets the filename of the package. There is no normalization
// done here. Comparing two pacman.Packages that refer to the same file might
// not have the same filename internally.
//
// If you need that, check out PkgBasename!
func PkgFilename(ap pacman.AnyPackage) string {
	return ap.Pkg().Filename
}

func PkgBasename(ap pacman.AnyPackage) string {
	return path.Base(ap.Pkg().Filename)
}

// PkgFilter returns a combination of many fields, separated by a space.
func PkgFilter(ap pacman.AnyPackage) string {
	var buf bytes.Buffer

	write := func(s string) {
		buf.WriteRune(' ')
		buf.WriteString(s)
	}
	writeAll := func(ss []string) {
		for _, s := range ss {
			write(s)
		}
	}

	p := ap.Pkg()
	buf.WriteString(p.Name)
	write(p.Base)
	write(p.Description)
	write(p.URL)
	writeAll(p.Groups)
	writeAll(p.Replaces)
	writeAll(p.Provides)

	return buf.String()
}
