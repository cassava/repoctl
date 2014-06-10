// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
)

// Status prints a thorough status of the current repository.
//
// Filters available are:
// 	duplicates      files to be deleted or backed up
// 	pending         packages to be added/removed from database
// 	installed       packages locally installed
// 	outdated        packages with newer versions in AUR
// 	missing         packages not found in AUR
func Status(c *Config) error {
	return status(c)
}

func status(c *Config) error {
	pkgs, outdated := getRepoPkgs(c.path)
	db, missed := getDatabasePkgs(c.Repository)

	name := c.database[:strings.IndexByte(c.database, '.')]
	fmt.Printf("On repo %s\n", name)

	// We assume that there is nothing to do, and if there is,
	// then this is set to false.
	var nothing = true

	if len(outdated) > 0 {
		printSet(mapPkgs(outdated, pkgFilename), "Obsolete packages to be removed/backed up:", c.Columnate)
		nothing = false
	}

	if len(missed) > 0 {
		printSet(missed, "Database entries pending removal:", c.Columnate)
		nothing = false
	}

	pending := filterPending(pkgs, db)
	if len(pending) > 0 {
		printSet(mapPkgs(pending, pkgName), "Database entries pending addition:", c.Columnate)
		nothing = false
	}

	aur, unavailable := getAURPkgs(mapPkgs(pkgs, pkgName))
	if len(unavailable) > 0 {
		printSet(unavailable, "Packages unavailable in AUR:", c.Columnate)
		nothing = false
	}

	updates := filterUpdates(pkgs, aur)
	if len(updates) > 0 {
		printSet(mapPkgs(updates, pkgName), "Packages with updates in AUR:", c.Columnate)
		nothing = false
	}

	if nothing {
		fmt.Println("Everything up-to-date.")
	}
	return nil
}
