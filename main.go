// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
)

func Help() {
	fmt.Printf("%s %s (%s)\n", progName, progVersion, progDate)
	fmt.Println(`
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
`)

}

func main() {
	if len(os.Args) == 1 {
		Help()
		os.Exit(1)
	}

	config := NewConfig("/srv/abs", "atlas.db.tar.gz")
	config.Args = os.Args[2:]
	cmd := os.Args[1]

	switch cmd {
	case "list", "ls":
		List(config)
	case "update":
		Update(config)
	case "add":
		Add(config)
	case "remove", "rm":
		Remove(config)
	case "synchronize", "sync":
		Sync(config)
	case "help":
		Help()
	default:
		fmt.Printf("Error: unrecognized command '%s'\n", cmd)
		Help()
		os.Exit(1)
	}
}
