// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Package meta binds together the three places a package can reside:
// filesystem, database, and internet. The filesystem takes first priority.
package meta

import (
	"github.com/goulash/pacman"
	"github.com/goulash/pacman/alpm"
	"github.com/goulash/pacman/aur"
)

type Packages []*Package

func (mps Packages) Len() int           { return len(mps) }
func (mps Packages) Swap(i, j int)      { mps[i], mps[j] = mps[j], mps[i] }
func (mps Packages) Less(i, j int) bool { return mps[i].Name < mps[j].Name }

func (mps Packages) Pkgs() pacman.Packages {
	pkgs := make(pacman.Packages, len(mps))
	for i, p := range mps {
		pkgs[i] = p.Pkg()
	}
	return pkgs
}
func (mps Packages) Iterate(f func(pacman.AnyPackage)) {
	for _, p := range mps {
		f(p)
	}
}

// Package binds together the three places a package can reside:
// filesystem, database, and internet. The filesystem takes first priority.
//
// The packages in Files are guaranteed to be sorted, so that the most
// recent version is first. It is illegal for neither Files nor Database
// to contain a valid Package.
type Package struct {
	Name string

	Files    pacman.Packages
	Database *pacman.Package
	AUR      *aur.Package
}

// Package returns the newest actual package available. This disregards
// whatever is in the database. If there are no files, then the database
// package is returned.
func (mp *Package) Pkg() *pacman.Package {
	if len(mp.Files) == 0 {
		return mp.Database
	}
	return mp.Files[0]
}
func (mp *Package) PkgName() string    { return mp.Name }
func (mp *Package) PkgVersion() string { return mp.Version() }

// Version returns the newest actual package version available. This
// disregards whatever is in the database. If there are no files, then
// an empty string is returned.
func (mp *Package) Version() string {
	if p := mp.Pkg(); p != nil {
		return p.Version
	}
	return ""
}

// VersionRegistered returns the version that is registered in the database.
func (mp *Package) VersionRegistered() string {
	if mp.Database == nil {
		return ""
	}
	return mp.Database.Version
}

// IsSynced returns true when the package is completely up-to-date and there
// is nothing to do.
func (mp *Package) IsSynced() bool {
	if mp.HasPending() {
		return false
	}
	return mp.HasUpgrade()
}

// HasObsolete returns true if there are obsolete files to be deleted.
func (mp *Package) HasObsolete() bool {
	return len(mp.Files) > 1
}

// HasPending returns true if there are any pending changes to filesystem
// or database concerning this package. This has nothing to do with
// whether there is an upgrade available in AUR.
func (mp *Package) HasPending() bool {
	if mp.Database == nil || len(mp.Files) != 1 {
		// pending addition/deletion to/from database
		// or pending deletion/backup of obsolete files
		return true
	}
	// pending something if versions aren't identical
	p := mp.Pkg()
	if p.Filename != mp.Database.Filename {
		return true
	}
	return alpm.VerCmp(p.Version, mp.Database.Version) != 0
}

// HasFiles returns true when there are files for this package.
func (mp *Package) HasFiles() bool {
	return len(mp.Files) > 0
}

// IsRegistered returns true when the package registered in the database
// does not exist or when the package is not registered.
func (mp *Package) IsRegistered() bool {
	if mp.Database == nil {
		return false
	}
	for _, p := range mp.Files {
		if p.Filename == mp.Database.Filename {
			return true
		}
	}
	return false
}

// HasUpdate returns true when there is a newer package file available
// that hasn't been added to the database. This includes the case where
// there is no database entry for this package.
func (mp *Package) HasUpdate() bool {
	if len(mp.Files) == 0 {
		return false
	}
	return !mp.IsRegistered() || mp.Pkg().Newer(mp.Database)
}

// HasUpgrade returns true when there is a newer version than either
// file or database.
func (mp *Package) HasUpgrade() bool {
	if mp.AUR == nil {
		return false
	}
	c := alpm.VerCmp(mp.AUR.Version, mp.Version())
	return c > 0
}

func (mp *Package) Obsolete() pacman.Packages {
	if mp.HasObsolete() {
		return mp.Files[1:]
	}
	return nil
}
