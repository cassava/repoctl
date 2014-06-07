// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GetAllPackages takes a directory path as an argument, and
// then reads all the package information into a list.
func ReadDir(dirpath string) []*Package {
	var pkgs []*Package

	filepath.Walk(dirpath, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: %s\n", err)
			return nil
		}
		if !info.Mode().IsDir() && HasPackageFormat(filename) {
			p, err := ReadPackage(filename)
			if err != nil {
				log.Printf("Warning: %s\n", err)
				return nil
			}

			pkgs = append(pkgs, p)
		}

		return nil
	})

	return pkgs
}

func ReadMatchingName(dirpath, pkgname string) []*Package {
	var pkgs []*Package

	filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: %s\n", err)
			return nil
		}
		if !info.Mode().IsDir() && strings.HasPrefix(path, pkgname) && HasPackageFormat(path) {
			p, err := ReadPackage(path)
			if err != nil {
				log.Printf("Warning: %s\n", err)
				return nil
			}

			if p.Name == pkgname {
				pkgs = append(pkgs, p)
			}
		}

		return nil
	})

	return pkgs
}

// Note that we do not recurse into subdirectories!
func ReadMatchingNames(dirpath string, pkgnames []string) []*Package {
	var pkgs []*Package

	for _, n := range pkgnames {
		matches, err := filepath.Glob(filepath.Join(dirpath, n+"-*.pkg.tar.*"))
		if err != nil {
			log.Printf("Warning: cannot find package %s.\n", n)
			continue
		}

		for _, fp := range matches {
			p, err := ReadPackage(fp)
			if err != nil {
				log.Printf("Warning: %s\n.", err)
				continue
			}

			if p.Name == n {
				pkgs = append(pkgs, p)
			}
		}
	}

	return pkgs
}
