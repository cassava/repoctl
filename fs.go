// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadDir reads all the packages it finds in a directory, recursing into
// subdirectories.
//
// Errors are passed to errHandler if errHandler is not nil. If errHandler
// returns an error, then the function aborts and returns the already read
// packages and the error.
//
// Example:
//
//	// Our errHandler never returns an error, so neither does ReadDir.
//	pkgs := ReadDir(dirpath, func(err error) error {
//		fmt.Println("error:", err)
//		return nil
//	})
//
func ReadDir(dirpath string, errHandler func(error) error) (Packages, error) {
	var pkgs Packages
	err := filepath.Walk(dirpath, func(filename string, info os.FileInfo, err error) error {
		if err != nil && errHandler != nil {
			return errHandler(err)
		}
		if !info.Mode().IsDir() && HasPackageFormat(filename) {
			p, err := ReadPackage(filename)
			if err != nil && errHandler != nil {
				return errHandler(err)
			}

			pkgs = append(pkgs, p)
		}

		return nil
	})

	return pkgs, err
}

// ReadMatchingName reads all packages with the given name in a directory,
// recursing into subdirectories.
//
// Error handling is managed in the same way as in ReadDir.
func ReadMatchingName(dirpath, pkgname string, errHandler func(error) error) (Packages, error) {
	var pkgs Packages
	err := filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if err != nil && errHandler != nil {
			return errHandler(err)
		}
		if !info.Mode().IsDir() && strings.HasPrefix(path, pkgname) && HasPackageFormat(path) {
			p, err := ReadPackage(path)
			if err != nil && errHandler != nil {
				return errHandler(err)
			}

			if p.Name == pkgname {
				pkgs = append(pkgs, p)
			}
		}

		return nil
	})

	return pkgs, err
}

// ReadMatchingNames reads all packages with one of the given names in a directory,
// at the moment it does not recurse into subdirectories.
//
// Error handling is managed the same as in ReadDir.
func ReadMatchingNames(dirpath string, pkgnames []string, errHandler func(error) error) (Packages, error) {
	var pkgs Packages
	var err error

	for _, n := range pkgnames {
		var matches []string
		matches, err = filepath.Glob(filepath.Join(dirpath, n+"-*.pkg.tar.*"))
		if err != nil && errHandler != nil {
			err = errHandler(fmt.Errorf("cannot find package '%s'", n))
			if err != nil {
				break
			}
			continue
		}
		for _, fp := range matches {
			p, err := ReadPackage(fp)
			if err != nil && errHandler != nil {
				err = errHandler(err)
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
