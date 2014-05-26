// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/goulash/util"
)

// packageExtensions lists the file extensions by which we recognize packages.
// It is only used by the function HasPackageFormat, which is useful for quickly
// separating the wheat from the chaff (e.g. filter out all the packages in a
// directory.)
var packageExtensions = []string{".pkg.tar.xz", ".pkg.tar.gz", ".pkg.tar.bz2"}

// The Package datatype represents all the information that encompasses a Pacman
// package, including the filename of the package.
type Package struct {
	Filename string

	Name            string    // pkgname
	Version         string    // pkgver
	Description     string    // pkgdesc
	Base            string    // pkgbase
	Epoch           uint64    // epoch
	URL             string    // url
	BuildDate       time.Time // builddate
	Packager        string    // packager
	Size            uint64    // size
	Arch            string    // arch: one of any, i686, or x86_64
	License         string    // license
	Backups         []string  // backup
	Replaces        []string  // replaces
	Provides        []string  // provides
	Conflicts       []string  // conflict
	Groups          []string  // group
	Depends         []string  // depend
	OptionalDepends []string  // optdepend
	MakeDepends     []string  // makedepend
	CheckDepends    []string  // checkdepend
	MakeOptions     []string  // makepkgopt
}

// VersionLess compares the version of pi to pkg, and returns true if pi is
// older. It takes the Epoch value in account.
func (pkg *Package) VersionLess(alt *Package) bool {
	// If the Epoch values are different, the package with the higher
	// Epoch value is always more recent.
	if pkg.Epoch != alt.Epoch {
		return pkg.Epoch < alt.Epoch
	}
	return pkg.Version < alt.Version
}

// HasPackageFormat returns true if the filename matches a known
// pacman package format.
func HasPackageFormat(path string) bool {
	for _, ext := range packageExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// ReadPackage reads the package information from a pacman package
// and returns it in the Package datatype.
func ReadPackage(path string) (*Package, error) {
	bs, err := util.ReadFileFromArchive(path, ".PKGINFO")
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(bs)
	info, err := readPackageInfo(r)
	if err != nil {
		return nil, err
	}

	info.Filename = path
	return info, nil
}

// readPackageInfo reads the package information from a pacman package.
//
// We don't do any specific controlling for you, so you should use
// HasPackageFormat on a path string before using this function on it.
// Even if you don't, nothing bad should happen, but just in case.
func readPackageInfo(r io.Reader) (*Package, error) {
	var info Package
	var err error

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix("#", line) {
			continue
		}

		kv := strings.Split(line, " = ")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "pkgname":
			info.Name = kv[1]
		case "pkgver":
			info.Version = kv[1]
		case "pkgdesc":
			info.Description = kv[1]
		case "pkgbase":
			info.Base = kv[1]
		case "epoch":
			info.Epoch, err = strconv.ParseUint(kv[1], 10, 64)
			if err != nil {
				log.Printf("Warning: cannot parse epoch value '%s'\n", kv[1])
			}
		case "url":
			info.URL = kv[1]
		case "builddate":
			n, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				log.Printf("Warning: cannot parse build time '%s'\n", kv[1])
			}
			info.BuildDate = time.Unix(n, 0)
		case "packager":
			info.Packager = kv[1]
		case "size":
			info.Size, err = strconv.ParseUint(kv[1], 10, 64)
			if err != nil {
				log.Printf("Warning: cannot parse size value '%s'\n", kv[1])
			}
		case "arch":
			info.Arch = kv[1]
		case "license":
			info.License = kv[1]
		case "depend":
			info.Depends = append(info.Depends, kv[1])
		case "optdepend":
			info.OptionalDepends = append(info.OptionalDepends, kv[1])
		case "makedepend":
			info.MakeDepends = append(info.MakeDepends, kv[1])
		case "checkdepend":
			info.CheckDepends = append(info.CheckDepends, kv[1])
		case "makepkgopt":
			info.MakeOptions = append(info.MakeOptions, kv[1])
		case "backup":
			info.Backups = append(info.Backups, kv[1])
		case "replaces":
			info.Replaces = append(info.Replaces, kv[1])
		case "provides":
			info.Provides = append(info.Provides, kv[1])
		case "conflict":
			info.Conflicts = append(info.Conflicts, kv[1])
		case "group":
			info.Groups = append(info.Groups, kv[1])
		default:
			log.Printf("Warning: unknown field '%s' in .PKGINFO\n", kv[0])
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetAllPackages takes a directory path as an argument, and
// then reads all the package information into a list.
func GetAllPackages(path string) []*Package {
	var pkgs []*Package

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Warning: %s\n", err)
			return nil
		}
		if !info.Mode().IsDir() && HasPackageFormat(path) {
			p, err := ReadPackage(path)
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

// Note that we recurse into subdirectories.
func GetMatchingPackages(path, pkgname string) []*Package {
	var pkgs []*Package

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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
func GetAllMatchingPackages(path string, pkgnames []string) []*Package {
	var pkgs []*Package

	for _, n := range pkgnames {
		matches, err := filepath.Glob(filepath.Join(path, n+"-*.pkg.tar.*"))
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

// SplitOldPackages splits the input array into one containing the newest
// packages and another containing the outdated packages.
func SplitOldPackages(pkgs []*Package) (updated []*Package, old []*Package) {
	var m = make(map[string]*Package)

	// Find out which packages are newest and put the others in the old array.
	for _, p := range pkgs {
		if cur, ok := m[p.Name]; ok {
			if cur.VersionLess(p) {
				old = append(old, cur)
			} else {
				old = append(old, p)
				continue
			}
		}
		m[p.Name] = p
	}

	// Add the newest packages to the updated array and return.
	updated = make([]*Package, 0, len(m))
	for _, v := range m {
		updated = append(updated, v)
	}

	return updated, old
}

func SearchAUR(pkgname string) []*Package {
	//https://aur.archlinux.org/rpc.php?type=info&arg=dropbox
}
