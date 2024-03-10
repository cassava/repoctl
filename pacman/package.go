// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import (
	"time"

	"github.com/cassava/repoctl/pacman/alpm"
)

// Packages is merely a list of packages
type Packages []*Package

func (pkgs Packages) Len() int      { return len(pkgs) }
func (pkgs Packages) Swap(i, j int) { pkgs[i], pkgs[j] = pkgs[j], pkgs[i] }
func (pkgs Packages) Less(i, j int) bool {
	if pkgs[i].Name != pkgs[j].Name {
		return pkgs[i].Name < pkgs[j].Name
	}
	return alpm.VerCmp(pkgs[i].Version, pkgs[j].Version) == -1
}

func (pkgs Packages) Pkgs() Packages { return pkgs }
func (pkgs Packages) Iterate(f func(AnyPackage)) {
	for _, p := range pkgs {
		f(p)
	}
}
func (pkgs Packages) ToMap() map[string]*Package {
	m := make(map[string]*Package)
	for _, p := range pkgs {
		m[p.Name] = p
	}
	return m
}

// PackageOrigin exists to document which fields in the Package type can be
// expected to be filled with data. Note that some fields may be blank
// because there is nothing to specify, such as MakeDepends.
type PackageOrigin int

const (
	// UnknownOrigin specifies unknown origin. No assumption may be made as
	// to what fields are filled.
	UnknownOrigin PackageOrigin = iota

	// FileOrigin specifies package file origin. All fields are filled in as
	// available.
	FileOrigin

	// DatabaseOrigin specifies database origin. All fields are filled in as
	// available.
	DatabaseOrigin

	// LocalOrigin specifies local origin. Not sure what fields are filled in.
	LocalOrigin

	// AUROrigin specifies AUR search origin. Only the following fields are
	// touched:
	//
	// 	Name
	//  Base
	// 	Version
	// 	Description
	// 	URL
	// 	License
	AUROrigin
)

// The Package datatype represents all the information that encompasses a Pacman
// package, including the filename of the package.
//
// Note: While we could include information from the database or an AUR search,
// we have decided against it for now. If you feel that this is important,
// please contact us.
type Package struct {
	// Filename is the file that the package is either read from, or that
	// the package refers to (for example from the database). There is no
	// guarantee over the format of the filename! It could be a partial
	// path or an absolute path.
	Filename string
	Origin   PackageOrigin

	Name            string    // pkgname
	Version         string    // pkgver
	Description     string    // pkgdesc
	Base            string    // pkgbase
	URL             string    // url
	BuildDate       time.Time // builddate
	Packager        string    // packager
	Size            uint64    // size
	Arch            string    // arch: one of any, i686, or x86_64
	License         string    // license
	Backups         []string  // backup
	Replaces        []string  // replaces
	Provides        []string  // provides
	Conflicts       []string  // conflict
	Groups          []string  // group
	Depends         []string  // depend
	OptionalDepends []string  // optdepend
	MakeDepends     []string  // makedepend
	CheckDepends    []string  // checkdepend
	MakeOptions     []string  // makepkgopt
	Xdata           []string  // xdata
}

func (p *Package) Pkg() *Package            { return p }
func (p *Package) PkgName() string          { return p.Name }
func (p *Package) PkgVersion() string       { return p.Version }
func (p *Package) PkgDepends() []string     { return p.Depends }
func (p *Package) PkgMakeDepends() []string { return p.MakeDepends }

// Check if one package is the same as another.
//
// The equality comparisons for the []string attributes
// are set comparisons.
func (p *Package) Equals(a *Package) bool {
	// If the pointer is the same, we are wasting time.
	if p == a {
		return true
	}

	if p.Filename != a.Filename {
		return false
	}
	if p.Origin != a.Origin {
		return false
	}
	if p.Name != a.Name {
		return false
	}
	if p.Version != a.Version {
		return false
	}
	if p.Description != a.Description {
		return false
	}
	if p.Base != a.Base {
		return false
	}
	if p.URL != a.URL {
		return false
	}
	if p.BuildDate != a.BuildDate {
		return false
	}
	if p.Packager != a.Packager {
		return false
	}
	if p.Size != a.Size {
		return false
	}
	if p.Arch != a.Arch {
		return false
	}
	if p.License != a.License {
		return false
	}
	if !isequalset(p.Backups, a.Backups) {
		return false
	}
	if !isequalset(p.Replaces, a.Replaces) {
		return false
	}
	if !isequalset(p.Provides, a.Provides) {
		return false
	}
	if !isequalset(p.Conflicts, a.Conflicts) {
		return false
	}
	if !isequalset(p.Groups, a.Groups) {
		return false
	}
	if !isequalset(p.Depends, a.Depends) {
		return false
	}
	if !isequalset(p.OptionalDepends, a.OptionalDepends) {
		return false
	}
	if !isequalset(p.MakeDepends, a.MakeDepends) {
		return false
	}
	if !isequalset(p.CheckDepends, a.CheckDepends) {
		return false
	}
	if !isequalset(p.MakeOptions, a.MakeOptions) {
		return false
	}
	if !isequalset(p.Xdata, a.Xdata) {
		return false
	}

	return true
}

// Older returns true if pkg's version is older than alt's.
// If alt is nil, then false is returned.
// It takes the Epoch value into account.
func (pkg *Package) Older(alt *Package) bool {
	if alt == nil {
		return false
	}
	return alpm.VerCmp(pkg.Version, alt.Version) == -1
}

// Newer returns true if pkg's version is newer than alt's.
// If alt is nil, then true is returned.
// It takes the Epoch value into account.
func (pkg *Package) Newer(alt *Package) bool {
	if alt == nil {
		return true
	}
	return alpm.VerCmp(pkg.Version, alt.Version) == 1
}

func isequalset(a, b []string) bool {
	if &a == &b || (len(a) == 0 && len(b) == 0) {
		return true
	}
	return issubset(a, b) && issubset(b, a)
}

func issubset(a, b []string) bool {
	m := make(map[string]bool)
	for _, k := range b {
		m[k] = true
	}
	for _, k := range a {
		if !m[k] {
			return false
		}
	}
	return true
}
