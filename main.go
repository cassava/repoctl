// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "os"

const (
	progName    = "repo"
	progVersion = "2.0.0"
	progDate    = "22. May 2014"
	progString  = progName + " " + progVersion + " (" + progDate + ")"

	configPath = "~/.repo.conf"

	sysRepoAdd    = "/usr/bin/repo-add"
	sysRepoRemove = "/usr/bin/repo-remove"

	pkgStrictExt  = "-[0-9]+-(any|i686|x86_64).pkg.tar.(gz|bz2|xz)$"
	pkgLenientExt = ".pkg.tar.(gz|bz2|xz)$"
	pkgExt        = pkgLenientExt
)

const usage = `
Manage local pacman repositories.

Commands available:
  add <pkgname>    Add the package(s) with <pkgname> to the database by
                   finding in the same directory of the database the latest
                   file for that package (by file modification date),
                   deleting the others, and updating the database.
  list             List all the packages that are currently available.
  (ls)             Note that this has nothing to do with the database.
  remove <pkgname> Remove the package with <pkgname> from the database, by
  (rm)             removing its entry from the database and deleting the files
                   that belong to it.
  update           Same as add, except scan and add changed packages.
  synchronize      Compare packages in the database to AUR for new versions.
  (sync)

NOTE: In all of these cases, <pkgname> is the name of the package, without
anything else. For example: pacman, and not pacman-3.5.3-1-i686.pkg.tar.xz
`

type Configuration struct {
	Verbose bool
	Confirm bool

	ConfigFile   string
	DatabaseName string
	DatabaseDir  string
	DatabasePath string

	PackageNames []string
}

func main() {
	args := os.Args[1:]
	for _, dir := range args {
		List(dir)
	}
}
