// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package alpm works with parts of Arch Linux packages.
package alpm

import "strings"

// PackageGlob is a glob that should only find packages.
const PackageGlob = "-*.pkg.tar*"

// PackageRegex is a regex that matches MOST well-named packages.
//
// The following groups are available:
//
//		Group 0 is the entire match
//		Group 1 is the name
//		Group 2 is the version-release
//		Group 3 is the arch
const PackageRegex = `^([a-zA-Z.0-9_+-]+)-(\d.*-\d+)-(.*)\.pkg\.tar\..*$`

// PackageExtensions is a list of filename extensions that should only match
// for packages.
var PackageExtensions = []string{
	"pkg.tar",
	"pkg.tar.zst",
	"pkg.tar.xz",
	"pkg.tar.gz",
	"pkg.tar.bz2",
}

// DatabaseExtensions is a list of filename extensions that should only match
// for valid databases.
var DatabaseExtensions = []string{
	"db.tar",
	"db.tar.zst",
	"db.tar.xz",
	"db.tar.gz",
	"db.tar.bz2",
}

// HasDatabaseFormat returns true if the filename matches a pacman package
// format that we can do anything with.
//
// Currently, only the following formats are supported:
//	.db.tar.gz
//  .db.tar.xz
//  .db.tar.gz2
//  .db.tar.zst
//
func HasDatabaseFormat(filename string) bool {
	for _, ext := range DatabaseExtensions {
		if strings.HasSuffix(filename, "."+ext) {
			return true
		}
	}
	return false
}

// HasPackageFormat returns true if the filename matches a pacman package
// format that we can do anything with.
//
// Currently, only the following formats are supported:
//	.pkg.tar
//	.pkg.tar.xz
//	.pkg.tar.gz
//	.pkg.tar.bz2
//	.pkg.tar.zst
//
func HasPackageFormat(filename string) bool {
	for _, ext := range PackageExtensions {
		if strings.HasSuffix(filename, "."+ext) {
			return true
		}
	}
	return false
}
