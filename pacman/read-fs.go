// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cassava/repoctl/pacman/alpm"
	"github.com/goulash/errs"
)

// ReadDir reads all packages that are found in the repository
// directory.
func ReadDir(h errs.Handler, dirpath, dbpath string) (Packages, error) {
	errs.Init(&h)

	// 1. First make sure we can open the database before we do some
	// complicated logic.
	dirpath = filepath.Clean(dirpath)
	dbinfo, err := os.Stat(dbpath)
	if err != nil {
		return ReadEveryFileInDir(h, dirpath)
	}

	// 2. Read a list of packages from the database.
	// Fix the filenames of each of packages and then put them in a set.
	dbPkgs, err := ReadDatabase(dbpath)
	if err != nil {
		return ReadEveryFileInDir(h, dirpath)
	}
	pkgs := make(map[string]*Package)
	for _, p := range dbPkgs {
		pkgpath := filepath.Join(dirpath, filepath.Base(p.Filename))
		pkgs[pkgpath] = p
	}

	// 3. Get the list of packages in the directory, and cross-check with the
	// entries from the database to see which ones we need to read. At each
	// entry we add it to the results list. This gives us the ordering we
	// would have anyway from reading the directory, but we don't have to
	// open each file.
	dbtime := dbinfo.ModTime()
	results := make(Packages, 0, len(pkgs))
	err = filepath.Walk(dirpath, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return h(err)
		}
		if info.Mode().IsDir() {
			if filename == dirpath {
				return nil
			}
			return filepath.SkipDir
		}
		if alpm.HasPackageFormat(filename) {
			// If there is an entry in the database AND the database is newer
			// than the file, then we use the database entry and continue to
			// the next file.
			if p, ok := pkgs[filename]; ok {
				if dbtime.After(info.ModTime()) {
					results = append(results, p)
					return nil
				}
			}

			// Either the file doesn't exist or dbtime is older than this file.
			p, err := Read(filename)
			if err != nil {
				return h(err)
			}
			results = append(results, p)
		}
		return nil
	})

	// Much faster.
	return results, err
}

// ReadEveryFileInDir reads all the packages it finds in a directory.
//
// Note: this will try to read each package, and currently it does this
// in an inefficient manner, decompressing the entire package in memory.
// Even a directory with only a few files will take a while to process.
func ReadEveryFileInDir(h errs.Handler, dirpath string) (Packages, error) {
	errs.Init(&h)

	var pkgs Packages
	dirpath = filepath.Clean(dirpath)
	err := filepath.Walk(dirpath, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return h(fmt.Errorf("read file %s: %w", filename, err))
		}
		if info.Mode().IsDir() {
			if filename == dirpath {
				return nil
			}
			return filepath.SkipDir
		}
		if alpm.HasPackageFormat(filename) {
			p, err := Read(filename)
			if err != nil {
				return h(err)
			}

			pkgs = append(pkgs, p)
		}

		return nil
	})

	return pkgs, err
}

// ReadDirApproxOnlyNames returns the names of all packages it finds
// in a directory.
//
// This does not have the performance issues that ReadEveryFileInDir has.
func ReadDirApproxOnlyNames(h errs.Handler, dirpath string) ([]string, error) {
	errs.Init(&h)
	re := regexp.MustCompile(alpm.PackageRegex)

	var pkgs []string
	dirpath = filepath.Clean(dirpath)
	err := filepath.Walk(dirpath, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			return h(err)
		}
		if info.Mode().IsDir() {
			if filename == dirpath {
				return nil
			}
			return filepath.SkipDir
		}

		matches := re.FindStringSubmatch(filepath.Base(filename))
		if matches != nil {
			pkgs = append(pkgs, matches[1])
		}

		return nil
	})

	return pkgs, err
}

// ReadFiles reads all the given package files.
func ReadFiles(h errs.Handler, pkgfiles ...string) (Packages, error) {
	errs.Init(&h)
	if len(pkgfiles) == 0 {
		return nil, nil
	}

	pkgs := make(Packages, 0, len(pkgfiles))
	for _, pf := range pkgfiles {
		p, err := Read(pf)
		if err != nil {
			err = h(err)
			if err != nil {
				return pkgs, err
			}
		} else {
			pkgs = append(pkgs, p)
		}
	}
	return pkgs, nil
}

// ReadNames reads all packages with one of the given names in a directory.
func ReadNames(h errs.Handler, dirpath string, pkgnames ...string) (Packages, error) {
	errs.Init(&h)

	var pkgs Packages
	var err error

	for _, n := range pkgnames {
		var matches []string
		matches, err = filepath.Glob(filepath.Join(dirpath, n+alpm.PackageGlob))
		if err != nil {
			err = h(fmt.Errorf("cannot find package %q", n))
			if err != nil {
				break
			}
			continue
		}
		for _, fp := range matches {
			if !alpm.HasPackageFormat(fp) {
				// Globbing also finds signatures, which we currently ignore
				continue
			}
			p, err := Read(fp)
			if err != nil {
				err = h(err)
				if err != nil {
					break
				}
				continue
			}

			if p.Name == n {
				pkgs = append(pkgs, p)
			}
		}
	}

	return pkgs, err
}
