// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repo

import (
	"sort"

	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman"
	"github.com/cassava/repoctl/pacman/pkgutil"
	"github.com/goulash/errs"
)

func (r *Repo) ListDatabase(f pkgutil.MapFunc) ([]string, error) {
	if f == nil {
		f = pkgutil.PkgName
	}

	pkgs, err := r.ReadDatabase()
	if err != nil {
		return nil, err
	}
	return List(pkgs, f), nil
}

func (r *Repo) ListDirectory(h errs.Handler, f pkgutil.MapFunc) ([]string, error) {
	errs.Init(&h)
	if f == nil {
		f = pkgutil.PkgName
	}

	pkgs, err := r.ReadDir(h)
	if err != nil {
		return nil, err
	}
	return List(pkgs, f), nil
}

func (r *Repo) ListMeta(h errs.Handler, aur bool, f func(pacman.AnyPackage) string) ([]string, error) {
	errs.Init(&h)
	if f == nil {
		f = pkgutil.PkgName
	}

	pkgs, err := r.ReadMeta(h)
	if err != nil {
		return nil, err
	}
	if aur {
		term.Debugf("Querying AUR for packages ...\n")
		_ = pkgs.ReadAUR()
	}
	return List(pkgs, f), nil
}

func List(pkgs pacman.AnyPackages, f pkgutil.MapFunc) []string {
	sort.Sort(pkgs)

	ls := make([]string, 0, pkgs.Len())
	pkgs.Iterate(func(p pacman.AnyPackage) {
		s := f(p)
		if s != "" {
			ls = append(ls, s)
		}
	})
	return unique(ls)
}

func unique(ls []string) []string {
	if len(ls) == 0 {
		return ls
	}

	nls := make([]string, len(ls))
	nls[0] = ls[0]
	i := 0
	for _, s := range ls[1:] {
		if s != nls[i] {
			i++
			nls[i] = s
		}
	}
	return nls[:i+1]
}
