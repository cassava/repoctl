// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repo

import (
	"path"

	"github.com/goulash/pacman"
)

func (r *Repo) ReadPackages() pacman.Packages {
	pkgs, err := pacman.ReadDir(r.Directory, PrinterEH(r.Error))
	if err != nil {
		r.errorf("error: %s\n", err)
	}
	return pkgs
}

func (r *Repo) FindPackages(pkgname string) pacman.Packages {
	pkgs, err := pacman.ReadMatchingName(r.Directory, pkgname, PrinterEH(r.Error))
	if err != nil {
		r.errorf("error: %s\n", err)
	}
	return pkgs
}

// FindSimilar finds package files in the repository that
// have the same package name. The pkgfile given is filtered from
// the results.
//
// This makes this function useful for finding all the other
// packages given a particular package file. These other package files
// can be
func (r *Repo) FindSimilar(pkgfile string) pacman.Packages {
	name, err := ReadPackageName(pkgfile)
	if err != nil {
		r.errorf("error: %s\n", err)
		return nil
	}

	pkgs, err := pacman.ReadMatchingName(r.Directory, pkgname, PrinterEH(r.Error))
	if err != nil {
		r.errorf("error: %s\n", err)
		return pkgs
	}

	basefile := path.Base(pkgfile)
	return pacman.Filter(pkgs, func(p *pacman.Package) bool {
		return path.Base(p.Filename) != basefile
	})
}

func ReadPackageName(pkgfile string) (string, error) {
	pkg, err := pacman.ReadPackage(pkgfile)
	if err != nil {
		return "", err
	}
	return pkg.Name, nil
}
