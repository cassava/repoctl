// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repo

import (
	"path"

	"github.com/goulash/osutil"
)

func (r *Repo) Copy(pkgfile string) error {
	return r.add([]string{pkgfile}, QuiterEH(), osutil.CopyFileLazy, "copying")
}

func (r *Repo) CopyAll(pkgfiles []string, h ErrHandler) error {
	return r.add(pkgfiles, h, osutil.CopyFileLazy, "copying")
}

func (r *Repo) Move(pkgfile string) error {
	return r.add([]string{pkgfile}, QuiterEH(), osutil.MoveFileLazy, "moving")
}

func (r *Repo) MoveAll(pkgfiles []string, h ErrHandler) error {
	return r.add(pkgfiles, h, osutil.MoveFileLazy, "moving")
}

func (r *Repo) add(pkgfiles []string, h ErrHandler, ar func(string, string) error, lbl string) {
	added := make([]string, 0, len(pkgfiles))
	for _, src := range pkgfiles {
		dst := path.Join(r.Directory, path.Base(src))
		r.printf("%s and adding to repository: %s\n", lbl, dst)
		err := ar(src, dst)
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			continue
		}
		added = append(added, dst)
	}

	err = r.DatabaseAdd(added...)
	if err != nil {
		return err
	}

	pkgs := FindSimilar(dst)
	return r.Dispatch(mapPkgs(pkgs, pkgFilename))
}
