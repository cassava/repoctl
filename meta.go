// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"sort"

	"github.com/goulash/pacman"
)

// MetaPackage binds together the three places a package can reside:
// database, filesystem, and internet. The database takes first priority.
//
// The packages in Files are guaranteed to be sorted, so that the most
// recent version is first. It is illegal for neither Files nor Database
// to contain a valid Package.
type MetaPackage struct {
	Name string

	Files    pacman.Packages
	Database *pacman.Package
	AUR      *pacman.AURPackage
}

// Package returns the newest actual package available. This disregards
// whatever is in the database. If there are no files, then nil is returned.
func (mp *MetaPackage) Package() *pacman.Package {
	if len(mp.Files) == 0 {
		return nil
	}
	return mp.Files[len(mp.Files)-1]
}

// Version returns the newest actual package version available. This
// disregards whatever is in the database. If there are no files, then
// an empty string is returned.
func (mp *MetaPackage) Version() string {
	if p := mp.Package(); p != nil {
		return p.Version
	}
	return ""
}

// VersionRegistered returns the version that is registered in the database.
func (mp *MetaPackage) VersionRegistered() string {
	if mp.Database == nil {
		return ""
	}
	return mp.Database.Version
}

// Okay returns true when the package is completely up-to-date and there
// is nothing to do.
func (mp *MetaPackage) Okay() bool {
	if mp.HasPending() {
		return false
	}
	return mp.HasUpgrade()
}

// HasObsolete returns true if there are obsolete files to be deleted.
func (mp *MetaPackage) HasObsolete() bool {
	return len(mp.Files) > 1
}

// HasPending returns true if there are any pending changes to filesystem
// or database concerning this package. This has nothing to do with
// whether there is an upgrade available in AUR.
func (mp *MetaPackage) HasPending() bool {
	if mp.Database == nil || len(mp.Files) != 1 {
		// pending addition/deletion to/from database
		// or pending deletion/backup of obsolete files
		return true
	}
	// pending something if versions aren't identical
	p := mp.Package()
	if p.Filename != mp.Database.Filename {
		return true
	}
	return p.CompareVersion(mp.Database) != 0
}

// HasUpdate returns true when there is a newer package file available
// that hasn't been added to the database. This includes the case where
// there is no database entry for this package.
func (mp *MetaPackage) HasUpdate() bool {
	if len(mp.Files) == 0 {
		return false
	}
	return mp.Package().NewerThan(mp.Database)
}

// HasUpgrade returns true when there is a newer version than either
// file or database.
func (mp *MetaPackage) HasUpgrade() bool {
	if mp.AUR == nil {
		return false
	}
	c := pacman.VerCmp(mp.AUR.Version, mp.Version())
	return c > 0
}

type MetaPackages []*MetaPackage

func (mps MetaPackages) Len() int           { return len(mps) }
func (mps MetaPackages) Swap(i, j int)      { mps[i], mps[j] = mps[j], mps[i] }
func (mps MetaPackages) Less(i, j int) bool { return mps[i].Name < mps[j].Name }

// ReadMeta reads a list of MetaPackages. If pkgnames is empty, the entire
// repository is loaded. MetaPackages is returned sorted (ascending).
func (r *Repo) ReadMeta(h ErrHandler, aur bool, pkgnames ...string) (MetaPackages, error) {
	AssertHandler(&h)

	// Read the database and start tracking the packages.
	dbpkgs, err := r.ReadDatabase()
	if err != nil {
		return nil, err
	}
	mps := make(map[string]*MetaPackage)
	for _, p := range dbpkgs {
		mps[p.Name] = &MetaPackage{Name: p.Name, Database: p}
	}

	// Read the packages in the repository directory.
	fspkgs, err := r.ReadNames(h, pkgnames...)
	if err != nil {
		return nil, err
	}
	for _, p := range fspkgs {
		mp := mps[p.Name]
		if mp == nil {
			mps[p.Name] = &MetaPackage{Name: p.Name, Files: pacman.Packages{p}}
		} else {
			if mp.Files == nil {
				mp.Files = pacman.Packages{p}
			} else {
				mp.Files = append(mp.Files, p)
			}
		}
	}

	// What is in mps at this point (that is of relevance)?
	names := make([]string, 0, len(mps))
	if len(pkgnames) == 0 {
		for k := range mps {
			names = append(names, k)
		}
	} else {
		for _, k := range pkgnames {
			if _, ok := mps[k]; ok {
				names = append(names, k)
			}
		}
	}
	sort.Strings(names)

	// Read from AUR.
	if aur {
		aurpkgs, err := r.ReadAUR(h, names...)
		if err != nil {
			r.errorf("error reading aur: %s.\n", err)
		} else {
			for _, ap := range aurpkgs {
				mps[ap.Name].AUR = ap
			}
		}
	}

	pkgs := make(MetaPackages, len(names))
	i := 0
	for _, n := range names {
		mp := mps[n]
		if mp.Files != nil {
			sort.Sort(mp.Files)
		}
		pkgs[i] = mp
		i++
	}
	return pkgs, nil
}
