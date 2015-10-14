// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goulash/errs"
)

// ReadDir reads all the packages it finds in a directory.
func ReadDir(h errs.Handler, dirpath string) (Packages, error) {
	errs.Init(&h)

	var pkgs Packages
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
		if !info.Mode().IsDir() && HasPackageFormat(filename) {
			p, err := Read(filename)
			if err != nil && h != nil {
				println(err)
				return h(err)
			}

			pkgs = append(pkgs, p)
		}

		return nil
	})

	return pkgs, err
}

// ReadMatchingNames reads all packages with one of the given names in a directory.
func ReadNames(h errs.Handler, dirpath string, pkgnames ...string) (Packages, error) {
	errs.Init(&h)

	var pkgs Packages
	var err error

	for _, n := range pkgnames {
		var matches []string
		matches, err = filepath.Glob(filepath.Join(dirpath, n+pkgGlob))
		if err != nil {
			err = h(fmt.Errorf("cannot find package %q", n))
			if err != nil {
				break
			}
			continue
		}
		for _, fp := range matches {
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
