// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/goulash/osutil"
)

// PackageOrigin exists to document which fields in the Package type can be
// expected to be filled with data. Note that some fields may be blank
// because there is nothing to specify, such as MakeDepends.
type PackageOrigin int

const (
	// UnknownOrigin specifies unknown origin. No assumption may be made as
	// to what fields are filled.
	UnknownOrigin PackageOrigin = iota

	// FileOrigin specifies package file origin. All fields are filled in as
	// available.
	FileOrigin

	// DatabaseOrigin specifies database origin. All fields are filled in as
	// available.
	DatabaseOrigin

	// AUROrigin specifies AUR search origin. Only the following fields are
	// touched:
	//
	// 	Name
	// 	Version
	// 	Description
	// 	URL
	// 	License
	AUROrigin
)

// The Package datatype represents all the information that encompasses a Pacman
// package, including the filename of the package.
//
// Note: While we could include information from the database or an AUR search,
// we have decided against it for now. If you feel that this is important,
// please contact us.
type Package struct {
	// Filename is the file that the package is either read from, or that
	// the package refers to (for example from the database). There is no
	// guarantee over the format of the filename! It could be a partial
	// path or an absolute path.
	Filename string
	Origin   PackageOrigin

	Name            string    // pkgname
	Version         string    // pkgver
	Description     string    // pkgdesc
	Base            string    // pkgbase
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

// Check if one package is the same as another.
//
// The equality comparisons for the []string attributes
// are set comparisons.
func (p *Package) Equal(a *Package) bool {
	// If the pointer is the same, we are wasting time.
	if p == a {
		return true
	}

	if p.Filename != a.Filename {
		return false
	}
	if p.Origin != a.Origin {
		return false
	}
	if p.Name != a.Name {
		return false
	}
	if p.Version != a.Version {
		return false
	}
	if p.Description != a.Description {
		return false
	}
	if p.Base != a.Base {
		return false
	}
	if p.URL != a.URL {
		return false
	}
	if p.BuildDate != a.BuildDate {
		return false
	}
	if p.Packager != a.Packager {
		return false
	}
	if p.Size != a.Size {
		return false
	}
	if p.Arch != a.Arch {
		return false
	}
	if p.License != a.License {
		return false
	}
	if !isequalset(p.Backups, a.Backups) {
		return false
	}
	if !isequalset(p.Replaces, a.Replaces) {
		return false
	}
	if !isequalset(p.Provides, a.Provides) {
		return false
	}
	if !isequalset(p.Conflicts, a.Conflicts) {
		return false
	}
	if !isequalset(p.Groups, a.Groups) {
		return false
	}
	if !isequalset(p.Depends, a.Depends) {
		return false
	}
	if !isequalset(p.OptionalDepends, a.OptionalDepends) {
		return false
	}
	if !isequalset(p.MakeDepends, a.MakeDepends) {
		return false
	}
	if !isequalset(p.CheckDepends, a.CheckDepends) {
		return false
	}
	if !isequalset(p.MakeOptions, a.MakeOptions) {
		return false
	}

	return true
}

// OlderThan returns true if pkg's version is older than alt's.
// It takes the Epoch value into account.
//
// If alt is nil, then false is returned.
func (pkg *Package) OlderThan(alt *Package) bool {
	if pkg == nil {
		panic("pkg is nil")
	}
	if alt == nil {
		return false
	}
	return pkg.CompareVersion(alt) == -1
}

// NewerThan returns true if pkg's version is newer than alt's.
// It takes the Epoch value into account.
//
// If alt is nil, then true is returned.
func (pkg *Package) NewerThan(alt *Package) bool {
	if pkg == nil {
		panic("pkg is nil")
	}
	if alt == nil {
		return true
	}
	return pkg.CompareVersion(alt) == 1
}

// CompareVersion compares the versions of two packages, taking the Epoch
// value into account.
func (pkg *Package) CompareVersion(alt *Package) int {
	return VerCmp(pkg.Version, alt.Version)
}

// HasPackageFormat returns true if the filename matches a pacman package
// format that we can do anything with.
//
// Currently, only the following formats are supported:
//  .pkg.tar
//	.pkg.tar.xz
//	.pkg.tar.gz
//	.pkg.tar.bz2
//
func HasPackageFormat(filename string) bool {
	for _, ext := range []string{".pkg.tar", ".pkg.tar.xz", ".pkg.tar.gz", ".pkg.tar.bz2"} {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

// ReadPackage reads the package information from a pacman package
// and returns it in the Package datatype.
func ReadPackage(filename string) (*Package, error) {
	bs, err := osutil.ReadFileFromArchive(filename, ".PKGINFO")
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(bs)
	info, err := readFilePkgInfo(r)
	if err != nil {
		return nil, err
	}

	info.Filename = filename
	info.Origin = FileOrigin
	return info, nil
}

// readFilePkgInfo reads the package information from a pacman package.
//
// We don't do any specific controlling for you, so you should use
// HasPackageFormat on a path string before using this function on it.
// Even if you don't, nothing bad should happen, but just in case.
func readFilePkgInfo(r io.Reader) (*Package, error) {
	var (
		info  Package
		err   error
		epoch int
	)

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
			epoch, err = strconv.Atoi(kv[1])
			if err != nil {
				return nil, fmt.Errorf("cannot parse epoch value '%s'", kv[1])
			}
		case "url":
			info.URL = kv[1]
		case "builddate":
			n, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse build time '%s'", kv[1])
			}
			info.BuildDate = time.Unix(n, 0)
		case "packager":
			info.Packager = kv[1]
		case "size":
			info.Size, err = strconv.ParseUint(kv[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse size value '%s'", kv[1])
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
			return nil, fmt.Errorf("unknown field '%s' in .PKGINFO", kv[0])
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	// The Version field must include the epoch value. If it already has one,
	// we compare and take the maximum value.
	if epoch > 0 {
		if i := strings.IndexByte(info.Version, ':'); i != -1 {
			e, err := strconv.Atoi(info.Version[:i])
			if err != nil {
				return nil, fmt.Errorf("unable to read epoch from version '%s'", info.Version)
			}
			epoch = max(epoch, e)
			info.Version = info.Version[i+1:]
		}
		info.Version = fmt.Sprintf("%d:%s", epoch, info.Version)
	}

	return &info, nil
}
