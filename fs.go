// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ReadDir reads all the packages it finds in a directory, recursing into
// subdirectories.
//
// Errors are passed through the channel if the channel is not nil, otherwise
// they are ignored. Make sure you handle the errors right away, like so:
//
//	ch := make(chan error)
//	go func() {
//		for err := range ch {
//			fmt.Println("error:", err)
//		}
//	}()
//	pkgs := ReadDir(dirpath, ch)
//	close(ch)
//
// Because if you don't, the program will probably run into a deadlock when
// there is an error. Note that ReadDir does not close the channel, you have to
// do that yourself.
func ReadDir(dirpath string, ch chan<- error) []*Package {
	var pkgs []*Package

	filepath.Walk(dirpath, func(filename string, info os.FileInfo, err error) error {
		if err != nil {
			if ch != nil {
				ch <- err
			}
			return nil
		}
		if !info.Mode().IsDir() && HasPackageFormat(filename) {
			p, err := ReadPackage(filename)
			if err != nil {
				if ch != nil {
					ch <- err
				}
				return nil
			}

			pkgs = append(pkgs, p)
		}

		return nil
	})

	return pkgs
}

// ReadMatchingName reads all packages with the given name in a directory,
// recursing into subdirectories.
//
// Error handling is managed in the same way as in ReadDir.
func ReadMatchingName(dirpath, pkgname string, ch chan<- error) []*Package {
	var pkgs []*Package

	filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if ch != nil {
				ch <- err
			}
			return nil
		}
		if !info.Mode().IsDir() && strings.HasPrefix(path, pkgname) && HasPackageFormat(path) {
			p, err := ReadPackage(path)
			if err != nil {
				if ch != nil {
					ch <- err
				}
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

// ReadMatchingNames reads all packages with one of the given names in a directory,
// at the moment it does not recurse into subdirectories.
//
// Error handling is managed the same as in ReadDir.
func ReadMatchingNames(dirpath string, pkgnames []string, ch chan<- error) []*Package {
	var pkgs []*Package

	for _, n := range pkgnames {
		matches, err := filepath.Glob(filepath.Join(dirpath, n+"-*.pkg.tar.*"))
		if err != nil {
			if ch != nil {
				ch <- errors.New("cannot find package " + n)
			}
			continue
		}

		for _, fp := range matches {
			p, err := ReadPackage(fp)
			if err != nil {
				if ch != nil {
					ch <- err
				}
				continue
			}

			if p.Name == n {
				pkgs = append(pkgs, p)
			}
		}
	}

	return pkgs
}
