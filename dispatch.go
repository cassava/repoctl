// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"os"
	"path"

	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
)

func (r *Repo) Remove(pkgnames []string, h ErrHandler) error {
	pkgs, err := pacman.ReadMatchingNames(r.Directory, pkgnames, h)
	if err != nil {
		return err
	}
	err = h(r.DatabaseRemove(pkgs.Map(pacman.PkgName)...))
	if err != nil {
		return err
	}
	return r.Dispatch(pkgs.Map(pacman.PkgFilename), h)
}

func (r *Repo) Dispatch(pkgfiles []string, h ErrHandler) error {
	if r.Backup {
		return r.backup(pkgfiles, h)
	}
	return r.unlink(pkgfiles, h)
}

func (r *Repo) backup(pkgfiles []string, h ErrHandler) error {
	for _, f := range pkgfiles {
		src := path.Base(f)
		r.printf("backing up: %s\n", src)
		dst := path.Join(r.Directory, r.BackupDir, src)
		err := osutil.MoveFileLazy(src, dst)
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func (r *Repo) unlink(pkgfiles []string, h ErrHandler) error {
	for _, f := range pkgfiles {
		src := path.Base(f)
		r.printf("deleting: %s\n", src)
		err := os.Remove(src)
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}
