// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/goulash/util"
)

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

// OlderThan returns true if pkg's version is older than alt's.
// It takes the Epoch value into account.
func (pkg *Package) OlderThan(alt *Package) bool {
	return pkg.CompareVersion(alt) == -1
}

// NewerThan returns true if pkg's version is newer than alt's.
// It takes the Epoch value into account.
func (pkg *Package) NewerThan(alt *Package) bool {
	return pkg.CompareVersion(alt) == 1
}

// CompareVersion compares the versions of two packages, taking the Epoch
// value into account.
func (pkg *Package) CompareVersion(alt *Package) int {
	// If the Epoch values are different, the package with the higher Epoch
	// value is always more recent.
	if pkg.Epoch != alt.Epoch {
		if pkg.Epoch < alt.Epoch {
			return -1
		}
		return 1
	}

	return VerCmp(pkg.Version, alt.Version)
}

// HasPackageFormat returns true if the filename matches a pacman package
// format that we can do anything with.
//
// Currently, only the following formats are supported:
//	.pkg.tar.xz
//	.pkg.tar.gz
//	.pkg.tar.bz2
//
func HasPackageFormat(filename string) bool {
	for _, ext := range []string{".pkg.tar.xz", ".pkg.tar.gz", ".pkg.tar.bz2"} {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// ReadPackage reads the package information from a pacman package
// and returns it in the Package datatype.
func ReadPackage(filename string) (*Package, error) {
	bs, err := util.ReadFileFromArchive(filename, ".PKGINFO")
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(bs)
	info, err := readFilePkgInfo(r)
	if err != nil {
		return nil, err
	}

	info.Filename = filename
	return info, nil
}

// readFilePkgInfo reads the package information from a pacman package.
//
// We don't do any specific controlling for you, so you should use
// HasPackageFormat on a path string before using this function on it.
// Even if you don't, nothing bad should happen, but just in case.
func readFilePkgInfo(r io.Reader) (*Package, error) {
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
