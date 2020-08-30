// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"path"

	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman"
	"github.com/cassava/repoctl/pacman/aur"
	"github.com/cassava/repoctl/pacman/meta"
	pu "github.com/cassava/repoctl/pacman/pkgutil"
	"github.com/goulash/errs"
	"github.com/goulash/osutil"
)

// ReadDatabase reads the database at r.Directory/r.Database.
// If the database does not exist, then an empty list is returned.
func (r *Repo) ReadDatabase() (pacman.Packages, error) {
	dbpath := r.DatabasePath()
	ex, err := osutil.FileExists(dbpath)
	if err != nil {
		return nil, fmt.Errorf("cannot read database %s: %w", dbpath, err)
	} else if !ex {
		return make(pacman.Packages, 0), nil
	}
	pkgs, err := pacman.ReadDatabase(dbpath)
	r.MakeAbs(pkgs)
	return pkgs, err
}

func (r *Repo) ReadDir(h errs.Handler) (pacman.Packages, error) {
	pkgs, err := pacman.ReadDir(h, r.Directory, r.DatabasePath())
	r.MakeAbs(pkgs)
	return pkgs, err
}

// ReadNames returns all packages in the repository that match the given
// names. If no names are given, all packages found are returned.
func (r *Repo) ReadNames(h errs.Handler, pkgnames ...string) (pacman.Packages, error) {
	errs.Init(&h)
	if len(pkgnames) == 0 {
		return r.ReadDir(h)
	}

	pkgs, err := pacman.ReadNames(h, r.Directory, pkgnames...)
	r.MakeAbs(pkgs)
	return pkgs, err
}

// ReadMeta returns all meta packages in the repository that match the given
// names.  If no names are given, all packages in repository are returned.
func (r *Repo) ReadMeta(h errs.Handler, pkgnames ...string) (meta.Packages, error) {
	errs.Init(&h)

	pkgs, err := meta.Read(h, r.Directory, r.DatabasePath())
	if err != nil {
		return nil, err
	}
	if len(pkgnames) == 0 {
		return pkgs, nil
	}
	return pu.Filter(pkgs, pu.NameFltr(pkgnames)).(meta.Packages), nil
}

// ReadAUR reads the given package names from AUR. If no package names
// are given, ReadAUR reads all the names found in the repository.
//
// If you don't need this special feature on zero packages, then please
// use aur.ReadAll instead!
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
			term.Debugf("Note: package filename data incorrect: %s\n", p.Filename)
		}
		p.Filename = filepath
	}
}
