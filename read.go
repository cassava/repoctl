// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"fmt"
	"path"

	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
)

// ReadDatabase reads the database at r.Directory/r.Database.
// If the database does not exist, then an empty list is returned.
func (r *Repo) ReadDatabase() (pacman.Packages, error) {
	dbpath := path.Join(r.Directory, r.Database)
	ex, err := osutil.FileExists(dbpath)
	if err != nil {
		return nil, err
	} else if !ex {
		return make(pacman.Packages, 0), nil
	}
	return pacman.ReadDatabase(dbpath)
}

// ReadDirectory reads all packages that are found in the repository
// directory.
func (r *Repo) ReadDirectory(h ErrHandler) (pacman.Packages, error) {
	AssertHandler(&h)
	return pacman.ReadDir(r.Directory, h)
}

// ReadRepository reads all packages that are found in the repository
// directory as well as all packages that are found in the database.
// These packages are then merged as neccesary, so that you can see
// which packages are synced, only in the database, and only as files.
func (r *Repo) ReadRepository(h ErrHandler) (synced pacman.Packages, dbonly pacman.Packages, fsonly pacman.Packages, err error) {
	AssertHandler(&h)
	dbpkgs, err := r.ReadDatabase()
	if err != nil {
		return nil, nil, nil, err
	}
	filepkgs, err := r.ReadDirectory(h)
	if err != nil {
		return nil, nil, nil, err
	}

	synced = make(pacman.Packages, 0)
	dbonly = make(pacman.Packages, 0)
	fsonly = make(pacman.Packages, 0)

	db := dbpkgs.MapPkg(pacman.PkgFilename)
	for _, p := range filepkgs {
		if db[p.Filename] != nil {
			synced = append(synced, p)
			delete(db, p.Filename)
		} else {
			fsonly = append(fsonly, p)
		}
	}
	for _, p := range db {
		dbonly = append(dbonly, p)
	}
	return
}

// ReadNames returns all packages in the repository that match the given
// names. If no names are given, all packages found are returned.
func (r *Repo) ReadNames(h ErrHandler, pkgnames ...string) (pacman.Packages, error) {
	AssertHandler(&h)
	if len(pkgnames) == 0 {
		return r.ReadDirectory(h)
	}
	return pacman.ReadMatchingNames(r.Directory, pkgnames, h)
}

// ReadAUR reads the given package names from AUR. If no package names
// are given, ReadAUR reads all the names found in the repository.
func (r *Repo) ReadAUR(h ErrHandler, pkgnames ...string) (pacman.AURPackages, error) {
	AssertHandler(&h)
	var err error
	if len(pkgnames) == 0 {
		pkgnames, err = r.OnlyNames(h)
		if err != nil {
			return nil, err
		}
	}

	return pacman.ReadAllAUR(pkgnames)
}

// Upgrade represents an available AUR upgrade.
type Upgrade struct {
	Old *pacman.Package
	New *pacman.AURPackage
}

// Name returns the package name of the upgrade.
func (u *Upgrade) Name() string {
	return u.New.Name
}

// DownloadURL returns the download URL of the upgrade.
func (u *Upgrade) DownloadURL() string {
	return u.New.DownloadURL()
}

// Versions returns the old and the new version of the packages.
func (u *Upgrade) Versions() (from string, to string) {
	if u.Old == nil {
		return "", u.New.Version
	}
	return u.Old.Version, u.New.Version
}

// String returns a representation of the available upgrade, in the
// form:
//
//  pkgname: oldver -> newver
func (u *Upgrade) String() string {
	from, to := u.Versions()
	if from == "" {
		return fmt.Sprintf("%s: %s", u.Name(), to)
	}
	return fmt.Sprintf("%s: %s -> %s", u.Name(), from, to)
}

// Upgrades is a list of Upgrade, which can be sorted using sort.Sort.
type Upgrades []*Upgrade

func (u Upgrades) Len() int           { return len(u) }
func (u Upgrades) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u Upgrades) Less(i, j int) bool { return u[i].New.Name < u[j].New.Name }

// FindUpgrades finds all upgrades it finds to the given packages. If
// no package names are given, all available package names are searched.
func (r *Repo) FindUpgrades(h ErrHandler, pkgnames ...string) (Upgrades, error) {
	AssertHandler(&h)
	pkgs, err := r.FindNewest(h, pkgnames...)
	if err != nil {
		return nil, err
	}
	aur, err := pacman.ReadAllAUR(pkgs.Map(pacman.PkgName))
	if err != nil {
		return nil, err
	}

	am := make(map[string]*pacman.AURPackage)
	for _, p := range aur {
		am[p.Name] = p
	}
	upgrades := make(Upgrades, 0)
	for _, p := range pkgs {
		a := am[p.Name]
		if a.Package().NewerThan(p) {
			upgrades = append(upgrades, &Upgrade{p, a})
		}
	}
	return upgrades, nil
}

// FindNewest returns the newest package files found for the given
// package names. If no names are given, all names are searched for.
func (r *Repo) FindNewest(h ErrHandler, pkgnames ...string) (pacman.Packages, error) {
	AssertHandler(&h)

	var pkgs pacman.Packages
	var err error
	if len(pkgnames) == 0 {
		pkgs, err = r.ReadDirectory(h)
	} else {
		pkgs, err = r.ReadNames(h, pkgnames...)
	}
	if err != nil {
		return nil, err
	}

	return FilterNewest(pkgs), nil
}

// FindSimilar finds package files in the repository that
// have the same package name. The pkgfile given is filtered from
// the results.
//
// This makes this function useful for finding all the other
// packages given a particular package file.
//
// If no pkgfiles are given, nil is returned.
func (r *Repo) FindSimilar(h ErrHandler, pkgfiles ...string) (pacman.Packages, error) {
	AssertHandler(&h)
	if len(pkgfiles) == 0 {
		r.debugf("repoctl.(Repo).FindSimilar: pkgfiles empty.\n")
		return nil, nil
	}

	pkgs, err := ReadPackages(h, pkgfiles...)
	if err != nil {
		return nil, err
	}
	similar, err := r.ReadNames(h, pkgs.Map(pacman.PkgName)...)
	if err != nil {
		return nil, err
	}

	basefiles := pkgs.MapBool(pacman.PkgFilename)
	return pacman.Filter(similar, func(p *pacman.Package) bool {
		return !basefiles[p.Filename]
	}), nil
}

// FindUpdates finds all given packages with updates. If no names are
// given, all names are checked.
func (r *Repo) FindUpdates(h ErrHandler, pkgnames ...string) (pacman.Packages, error) {
	AssertHandler(&h)

	pkgs, err := r.FindNewest(h, pkgnames...)
	if err != nil {
		return nil, err
	}
	dbpkgs, err := r.ReadDatabase()
	if err != nil {
		return nil, err
	}
	updates := make(pacman.Packages, 0)
	db := dbpkgs.MapPkg(pacman.PkgName)
	for _, p := range pkgs {
		if p.NewerThan(db[p.Name]) {
			updates = append(updates, p)
		}
	}
	return updates, nil
}

// OnlyNames reads all possible package names from the repository.
// This includes packages from the database where the files have been
// deleted.
func (r *Repo) OnlyNames(h ErrHandler) ([]string, error) {
	AssertHandler(&h)
	dbpkgs, err := r.ReadDatabase()
	if err = h(err); err != nil {
		return nil, err
	}
	filepkgs, err := r.ReadDirectory(h)
	if err = h(err); err != nil {
		return nil, err
	}

	m := dbpkgs.MapBool(pacman.PkgName)
	for _, p := range filepkgs {
		m[p.Name] = true
	}
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	return names, nil
}
