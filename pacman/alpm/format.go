// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package alpm works with parts of Arch Linux packages.
package alpm

import "strings"

// PackageGlob is a glob that should only find packages.
const PackageGlob = "-*.pkg.tar*"

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
	for _, ext := range []string{".db.tar", ".db.tar.xz", ".db.tar.gz", ".db.tar.bz2", ".db.tar.zst"} {
		if strings.HasSuffix(filename, ext) {
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
	for _, ext := range []string{".pkg.tar", ".pkg.tar.xz", ".pkg.tar.gz", ".pkg.tar.bz2", ".pkg.tar.zst"} {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}
