// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"os"
	"path"

	"github.com/goulash/errs"
	"github.com/goulash/osutil"
	pu "github.com/goulash/pacman/pkgutil"
)

// Copy copies the given files into the repository if they do not already
// exist there and adds them to the database.
func (r *Repo) Copy(h errs.Handler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, osutil.CopyFileLazy, "copying")
}

// Move moves the given files into the repository if they do not already
// exist there and adds them to the database. If the files already exist
// there, then they are still deleted from where they were. Thus, the
// move always appears to have worked, even if no work was done.
//
// The exception is that when the source and destination files are the
// same; then no move or deletion is performed.
func (r *Repo) Move(h errs.Handler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, osutil.MoveFileLazy, "moving")
}

// add does the hard work of Move and Copy.
func (r *Repo) add(h errs.Handler, pkgfiles []string, ar func(string, string) error, lbl string) error {
	errs.Init(&h)
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
	return r.Dispatch(h, pu.Map(pkgs, pu.PkgFilename)...)
}

// Remove removes the given names from the database and dispatches
// the files.
func (r *Repo) Remove(h errs.Handler, pkgnames ...string) error {
	errs.Init(&h)
	if len(pkgnames) == 0 {
		r.debugf("repoctl.(Repo).Remove: pkgnames empty.\n")
		return nil
	}

	pkgs, err := r.ReadNames(h, pkgnames...)
	if err != nil {
		return err
	}
	err = h(r.DatabaseRemove(pu.Map(pkgs, pu.PkgName)...))
	if err != nil {
		return err
	}
	return r.Dispatch(h, pu.Map(pkgs, pu.PkgFilename)...)
}

// Dispatch either removes the given files or it backs them up.
func (r *Repo) Dispatch(h errs.Handler, pkgfiles ...string) error {
	errs.Init(&h)
	if len(pkgfiles) == 0 {
		r.debugf("repoctl.(Repo).Dispatch: pkgfiles empty.\n")
		return nil
	}

	if r.Backup {
		return r.backup(h, pkgfiles)
	}
	return r.unlink(h, pkgfiles)
}

func (r *Repo) backup(h errs.Handler, pkgfiles []string) error {
	for _, f := range pkgfiles {
		src := path.Base(f)
		r.printf("backing up: %s\n", f)
		dst := path.Join(r.Directory, r.BackupDir, src)
		err := osutil.MoveFileLazy(f, dst)
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

func (r *Repo) unlink(h errs.Handler, pkgfiles []string) error {
	for _, f := range pkgfiles {
		r.printf("deleting: %s\n", f)
		err := os.Remove(f)
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
func (r *Repo) Update(h errs.Handler, pkgnames ...string) error {
	errs.Init(&h)

	pkgs, err := r.ReadMeta(h, pkgnames...)
	if err != nil {
		return err
	}

	var updates []string
	var obsolete []string
	var missing []string
	for _, p := range pkgs {
		if !p.HasFiles() {
			missing = append(missing, p.Name)
			continue
		}
		if p.HasUpdate() || len(pkgnames) > 0 {
			updates = append(updates, p.Pkg().Filename)
		}
		if p.HasObsolete() {
			obsolete = append(obsolete, pu.Map(p.Obsolete(), pu.PkgFilename)...)
		}
	}

	err = r.DatabaseRemove(missing...)
	if err != nil {
		return err
	}

	err = r.DatabaseAdd(updates...)
	if err != nil {
		return err
	}

	return r.Dispatch(h, obsolete...)
}

// Delete the repository database and readd all the packages.
// This is the same as unlinking the database and then running Update.
func (r *Repo) Reset(h errs.Handler) error {
	errs.Init(&h)

	err := r.DeleteDatabase()
	if err != nil {
		return err
	}

	return r.Update(h)
}

// Delete the repository database (but not the files).
func (r *Repo) DeleteDatabase() error {
	db := path.Join(r.Directory, r.Database)
	r.printf("deleting database: %s\n", db)
	return os.Remove(db)
}
