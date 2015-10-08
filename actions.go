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

// Copy copies the given files into the repository if they do not already
// exist there and adds them to the database.
func (r *Repo) Copy(h ErrHandler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, osutil.CopyFileLazy, "copying")
}

// Move moves the given files into the repository if they do not already
// exist there and adds them to the database. If the files already exist
// there, then they are still deleted from where they were. Thus, the
// move always appears to have worked, even if no work was done.
//
// The exception is that when the source and destination files are the
// same; then no move or deletion is performed.
func (r *Repo) Move(h ErrHandler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, osutil.MoveFileLazy, "moving")
}

// add does the hard work of Move and Copy.
func (r *Repo) add(h ErrHandler, pkgfiles []string, ar func(string, string) error, lbl string) error {
	AssertHandler(&h)
	if len(pkgfiles) == 0 {
		r.debugf("repoctl.(Repo).add: pkgfiles empty.\n")
		return nil
	}

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

	err := r.DatabaseAdd(added...)
	if err != nil {
		return err
	}

	pkgs, err := r.FindSimilar(h, added...)
	return r.Dispatch(h, pkgs.Map(pacman.PkgFilename)...)
}

// Remove removes the given names from the database and dispatches
// the files.
func (r *Repo) Remove(h ErrHandler, pkgnames ...string) error {
	AssertHandler(&h)
	if len(pkgnames) == 0 {
		r.debugf("repoctl.(Repo).Remove: pkgnames empty.\n")
		return nil
	}

	pkgs, err := r.ReadNames(h, pkgnames...)
	if err != nil {
		return err
	}
	err = h(r.DatabaseRemove(pkgs.Map(pacman.PkgName)...))
	if err != nil {
		return err
	}
	return r.Dispatch(h, pkgs.Map(pacman.PkgFilename)...)
}

// Dispatch either removes the given files or it backs them up.
func (r *Repo) Dispatch(h ErrHandler, pkgfiles ...string) error {
	AssertHandler(&h)
	if len(pkgfiles) == 0 {
		r.debugf("repoctl.(Repo).Dispatch: pkgfiles empty.\n")
		return nil
	}

	if r.Backup {
		return r.backup(h, pkgfiles)
	}
	return r.unlink(h, pkgfiles)
}

func (r *Repo) backup(h ErrHandler, pkgfiles []string) error {
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

func (r *Repo) unlink(h ErrHandler, pkgfiles []string) error {
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

// Update adds the newest package found for the given name to the
// database and dispatches the obsolete packages. Any obsolete entries
// in the database are removed.
//
// If pkgnames is empty, the entire repository is scanned.
//
// TODO: What happens when there are multiple files, and you delete
// the most recent one. Which file is deleted?
func (r *Repo) Update(h ErrHandler, pkgnames ...string) error {
	AssertHandler(&h)

	var pkgs pacman.Packages
	var err error
	if len(pkgnames) == 0 {
		pkgs, err = r.FindUpdates(h)
		if err != nil {
			return err
		}
	} else {
		pkgs, err = r.FindNewest(h, pkgnames...)
		if err != nil {
			return err
		}
	}

	files := pkgs.Map(pacman.PkgFilename)
	err = r.DatabaseAdd(files...)
	if err != nil {
		return err
	}

	// Remove entries from database that have no associated files.
	dbpkgs, err := r.ReadDatabase()
	if err != nil {
		return err
	}
	var rm []string
	for _, p := range dbpkgs {
		if !r.Exists(p) {
			rm = append(rm, p.Name)
		}
	}
	err = r.DatabaseRemove(rm...)
	if err != nil {
		return err
	}

	// Dispatch all obsolete files.
	pkgs, err = r.FindSimilar(h, files...)
	if err != nil {
		return err
	}
	return r.Dispatch(h, pkgs.Map(pacman.PkgFilename)...)
}
