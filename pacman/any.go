// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import "github.com/cassava/repoctl/pacman/alpm"

type AnyPackages interface {
	Iterate(func(AnyPackage))
	Pkgs() Packages

	Len() int
	Swap(i, j int)
	Less(i, j int) bool
}

type AnyPackage interface {
	Pkg() *Package
	PkgName() string
	PkgVersion() string
	PkgDepends() []string
	PkgMakeDepends() []string
}

// PkgOlder returns true if a's version is older than b's.
// If b is nil, then false is returned.
func PkgOlder(a, b AnyPackage) bool {
	if b == nil {
		return false
	}
	return alpm.VerCmp(a.PkgVersion(), b.PkgVersion()) == -1

}

// PkgNewer returns true if a's version is newer than b's.
// If b is nil, then true is returned.
func PkgNewer(a, b AnyPackage) bool {
	if b == nil {
		return true
	}
	return alpm.VerCmp(a.PkgVersion(), b.PkgVersion()) == 1
}
