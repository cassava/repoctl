// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"path"

	"github.com/goulash/errs"
	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
	"github.com/goulash/pacman/aur"
	"github.com/goulash/pacman/pkgutil"
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
	pkgs, err := pacman.ReadDatabase(dbpath)
	r.MakeAbs(pkgs)
	return pkgs, err
}

// ReadDirectory reads all packages that are found in the repository
// directory.
func (r *Repo) ReadDirectory(h errs.Handler) (pacman.Packages, error) {
	errs.Init(&h)

	pkgs, err := pacman.ReadDir(h, r.Directory)
	r.MakeAbs(pkgs)
	return pkgs, err
}

// ReadRepository reads all packages that are found in the repository
// directory as well as all packages that are found in the database.
// These packages are then merged as neccesary, so that you can see
// which packages are synced, only in the database, and only as files.
func (r *Repo) ReadRepository(h errs.Handler) (synced pacman.Packages, dbonly pacman.Packages, fsonly pacman.Packages, err error) {
	errs.Init(&h)

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

	db := pkgutil.MapPkg(dbpkgs, pkgutil.PkgFilename)
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
func (r *Repo) ReadNames(h errs.Handler, pkgnames ...string) (pacman.Packages, error) {
	errs.Init(&h)
	if len(pkgnames) == 0 {
		return r.ReadDirectory(h)
	}

	pkgs, err := pacman.ReadNames(h, r.Directory, pkgnames...)
	r.MakeAbs(pkgs)
	return pkgs, err
}

// ReadAUR reads the given package names from AUR. If no package names
// are given, ReadAUR reads all the names found in the repository.
func (r *Repo) ReadAUR(h errs.Handler, pkgnames ...string) (aur.Packages, error) {
	errs.Init(&h)
	var err error
	if len(pkgnames) == 0 {
		pkgnames, err = r.OnlyNames(h)
		if err != nil {
			return nil, err
		}
	}

	return aur.ReadAll(pkgnames)
}

// MakeAbs makes all package filenames absolute. It is much easier
// to do this to all packages than figure out when we need it and when
// we don't.
func (r *Repo) MakeAbs(pkgs pacman.Packages) {
	for _, p := range pkgs {
		filepath := path.Join(r.Directory, path.Base(p.Filename))
		if p.Filename != filepath {
			r.debugf("repoctl.(Repo).Absolutify: pkgfile filename incorrect: %s\n", p.Filename)
		}
		p.Filename = filepath
	}
}
