// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"os"
	"path"

	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman"
	pu "github.com/cassava/repoctl/pacman/pkgutil"
	"github.com/goulash/errs"
	"github.com/goulash/osutil"
)

// Link tries to hard link the file, and failing that, copies it over.
func (r *Repo) Link(h errs.Handler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, linkFile, "Linking")
}

// TODO: Get this in the osutil package.
func linkFile(src, dst string) error {
	err := os.Link(src, dst)
	if err != nil {
		// If we can't link it (not same filesystem, etc.), then try copying.
		return osutil.CopyFileLazy(src, dst)
	}
	return nil
}

// Copy copies the given files into the repository if they do not already
// exist there and adds them to the database.
func (r *Repo) Copy(h errs.Handler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, osutil.CopyFileLazy, "Copying")
}

// Move moves the given files into the repository if they do not already
// exist there and adds them to the database. If the files already exist
// there, then they are still deleted from where they were. Thus, the
// move always appears to have worked, even if no work was done.
//
// The exception is that when the source and destination files are the
// same; then no move or deletion is performed.
func (r *Repo) Move(h errs.Handler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, osutil.MoveFileLazy, "Moving")
}

// add does the hard work of Move and Copy.
func (r *Repo) add(h errs.Handler, pkgfiles []string, ar func(string, string) error, lbl string) error {
	errs.Init(&h)
	if len(pkgfiles) == 0 {
		return nil
	}

	added := make([]string, 0, len(pkgfiles))
	for _, f := range pkgfiles {
		pkg, err := NewSignedPkg(f)
		if err != nil {
			// This means that we are trying to add something that's
			// non-existant or corrupt.
			term.Errorf("Skipping %s: %s", f, err)
			continue
		}
		if r.RequireSignature && !pkg.HasSignature() {
			term.Errorf("Skipping %s: require signature but none available", f)
			continue
		}

		term.Printf("%s and adding to repository: %s\n", lbl, pkg.PathSet())
		err = pkg.Apply(func(src string, _ bool) error {
			dst := path.Join(r.Directory, path.Base(src))
			return ar(src, dst)
		})
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			continue
		}
		added = append(added, path.Join(r.Directory, path.Base(f)))
	}

	err := r.AddToDatabase(added...)
	if err != nil {
		return err
	}

	pkgs, err := r.FindSimilar(h, added...)
	if err != nil {
		return err
	}
	return r.Dispatch(h, pu.Map(pkgs, pu.PkgFilename)...)
}

// Remove removes the given names from the database and dispatches
// the files.
func (r *Repo) Remove(h errs.Handler, pkgnames ...string) error {
	errs.Init(&h)
	if len(pkgnames) == 0 {
		return nil
	}

	pkgs, err := r.ReadNames(h, pkgnames...)
	if err != nil {
		return err
	}
	err = h(r.RemoveFromDatabase(pu.Map(pkgs, pu.PkgName)...))
	if err != nil {
		return err
	}
	return r.Dispatch(h, pu.Map(pkgs, pu.PkgFilename)...)
}

// Dispatch either removes the given files or it backs them up.
func (r *Repo) Dispatch(h errs.Handler, pkgfiles ...string) error {
	errs.Init(&h)
	if len(pkgfiles) == 0 {
		return nil
	}

	if r.Backup {
		return r.backup(h, pkgfiles)
	}
	return r.unlink(h, pkgfiles)
}

// IsObsoleteCached returns true when obsolete package files should
// be cached instead of backed-up or deleted.
//
// If the backup directory is the directory where all the
// packages are, then the idea is that we leave them in place.
func (r *Repo) IsObsoleteCached() bool {
	// TODO: this assumes that r.Directory is absolute and clean.
	return r.backupDirAbs() == r.Directory
}

// backupDirAbs returns the absolute path to the backup directory.
// If r.BackupDir is relative, then it is relative to the repository
// path, otherwise it is as is.
func (r *Repo) backupDirAbs() string {
	if path.IsAbs(r.BackupDir) {
		return path.Clean(r.BackupDir)
	}
	return path.Join(r.Directory, r.BackupDir)
}

func (r *Repo) backup(h errs.Handler, pkgfiles []string) error {
	if r.IsObsoleteCached() {
		for _, f := range pkgfiles {
			term.Debugf("Caching: %s\n", f)
		}
		return nil
	}

	backupDir := r.backupDirAbs()
	if ex, _ := osutil.DirExists(backupDir); !ex {
		term.Debugf("Creating directory: %s\n", backupDir)
		err := os.MkdirAll(backupDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("cannot create backup directory %s: %w", backupDir, err)
		}
	}

	for _, f := range pkgfiles {
		pkg, err := NewSignedPkg(f)
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			if pkg == nil {
				continue
			}
		}

		term.Printf("Backing up: %s\n", pkg.NameSet())
		err = pkg.Apply(func(f string, _ bool) error {
			src := path.Base(f)
			dst := path.Join(backupDir, src)
			return osutil.MoveFileLazy(f, dst)
		})
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Repo) unlink(h errs.Handler, pkgfiles []string) error {
	for _, f := range pkgfiles {
		pkg, err := NewSignedPkg(f)
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			if pkg == nil {
				continue
			}
		}

		term.Printf("Deleting: %s\n", pkg.NameSet())

		err = pkg.Apply(func(f string, _ bool) error {
			return os.Remove(f)
		})
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
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

	dbpath := r.DatabasePath()
	if pacman.IsDatabaseLocked(dbpath) {
		return fmt.Errorf("database is locked: %s", dbpath+".lck")
	}

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
		if p.HasObsolete() {
			obsolete = append(obsolete, pu.Map(p.Obsolete(), pu.PkgFilename)...)
		}
		if p.HasUpdate() || len(pkgnames) > 0 {
			f := p.Pkg().Filename
			if r.RequireSignature {
				spkg, err := NewSignedPkg(f)
				if err != nil {
					term.Errorf("Skipping %s: %s", f, err)
					continue
				} else if !spkg.HasSignature() {
					term.Errorf("Skipping %s: require signature but none found", f)
					continue
				}
			}
			updates = append(updates, f)
		}
	}

	err = r.RemoveFromDatabase(missing...)
	if err != nil {
		return err
	}

	err = r.AddToDatabase(updates...)
	if err != nil {
		return err
	}

	return r.Dispatch(h, obsolete...)
}
