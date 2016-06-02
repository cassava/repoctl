// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package meta

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/goulash/errs"
	"github.com/goulash/pacman"
	"github.com/goulash/pacman/alpm"
	"github.com/goulash/pacman/aur"
)

var ErrMultipleDB = errors.New("multiple database files found")

// ReadRepo tries to find a database in the specified directory.
// In any case, it loads all files it can. If there are multiple
// databases, it returns (nil, ErrMultipleDB). It does not recurse.
func ReadRepo(h errs.Handler, dirpath string) (Packages, error) {
	var dbpath string

	dirpath = filepath.Clean(dirpath)
	err := filepath.Walk(dirpath, func(filename string, info os.FileInfo, err error) error {
		if err != nil && h != nil {
			println(err)
			return h(err)
		}
		if info.Mode().IsDir() {
			if filename == dirpath {
				return nil
			}
			return filepath.SkipDir
		}
		if !info.Mode().IsDir() && alpm.HasDatabaseFormat(filename) {
			if dbpath != "" {
				return ErrMultipleDB
			}
			dbpath = filepath.Join(dirpath, filename)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return Read(h, dirpath, dbpath)
}

// Read reads meta packages in dirpath and using the database at dpbath.
// If dbpath == "", then no database is read. This function does not
// recurse.
//
// If no database can be read, the function still continues reading
// packages from the directory.
func Read(h errs.Handler, dirpath, dbpath string) (Packages, error) {
	errs.Init(&h)

	mps := make(map[string]*Package)

	// Read the database and start tracking the packages.
	if dbpath != "" {
		dbpkgs, err := pacman.ReadDatabase(dbpath)
		if err != nil {
			// Ignore this error, it's probably just a missing database.
			// It doesn't affect the results anyway.
			h(err)
		} else {
			for _, p := range dbpkgs {
				// This is the same as unlinking the database
				// and then running Update.
				mps[p.Name] = &Package{Name: p.Name, Database: p}
			}
		}
	}

	// Read the packages in the repository directory.
	fspkgs, err := pacman.ReadDir(h, dirpath)
	if err != nil {
		return nil, err
	}
	for _, p := range fspkgs {
		mp := mps[p.Name]
		if mp == nil {
			mps[p.Name] = &Package{Name: p.Name, Files: pacman.Packages{p}}
		} else {
			if mp.Files == nil {
				mp.Files = pacman.Packages{p}
			} else {
				mp.Files = append(mp.Files, p)
			}
		}
	}

	// Sort Files inside packages and prepare slice to return.
	pkgs := make(Packages, len(mps))
	i := 0
	for _, v := range mps {
		if v.Files != nil {
			sort.Sort(sort.Reverse(v.Files))
		}
		pkgs[i] = v
		i++
	}
	sort.Sort(pkgs)
	return pkgs, nil
}

// ReadAUR updates the list of packages by reading the AUR information for them.
// If any packages cannot be found, *aur.NotFoundError is returned. This can be
// ignored:
//
//  err := ps.ReadAUR()
//  if err != nil && !aur.IsNotFound(err) {
//      return err
//  }
//
func (ps Packages) ReadAUR() error {
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
	}
	aurpkgs, err := aur.ReadAll(names)
	if err != nil && !aur.IsNotFound(err) {
		return err
	}
	apm := make(map[string]*aur.Package)
	for _, ap := range aurpkgs {
		apm[ap.Name] = ap
	}
	for _, p := range ps {
		// If apm[p.Name] == nil, then no loss.
		p.AUR = apm[p.Name]
	}
	return err
}
