// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"sort"

	"github.com/goulash/errs"
	"github.com/goulash/pacman"
	"github.com/goulash/pacman/pkgutil"
)

func (r *Repo) ListDatabase(f func(*pacman.Package) string) ([]string, error) {
	if f == nil {
		f = pkgutil.PkgName
	}

	pkgs, err := r.ReadDatabase()
	if err != nil {
		return nil, err
	}
	return List(pkgs, f), nil
}

func (r *Repo) ListDirectory(h errs.Handler, f func(*pacman.Package) string) ([]string, error) {
	errs.Init(&h)
	if f == nil {
		f = pkgutil.PkgName
	}

	pkgs, err := r.ReadDirectory(h)
	if err != nil {
		return nil, err
	}
	return List(pkgs, f), nil
}

func (r *Repo) ListMeta(h errs.Handler, aur bool, f func(*MetaPackage) string) ([]string, error) {
	errs.Init(&h)
	if f == nil {
		f = func(mp *MetaPackage) string { return mp.Name }
	}

	pkgs, err := r.ReadMeta(h, aur)
	if err != nil {
		return nil, err
	}
	return ListMeta(pkgs, f), nil
}

func ListMeta(mpkgs MetaPackages, f func(*MetaPackage) string) []string {
	sort.Sort(mpkgs)

	ls := make([]string, 0, len(mpkgs))
	for _, p := range mpkgs {
		s := f(p)
		if s != "" {
			ls = append(ls, s)
		}
	}
	return unique(ls)
}

func List(pkgs pacman.Packages, f pkgutil.MapFunc) []string {
	sort.Sort(pkgs)

	ls := make([]string, 0, len(pkgs))
	for _, p := range pkgs {
		s := f(p)
		if s != "" {
			ls = append(ls, s)
		}
	}
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
