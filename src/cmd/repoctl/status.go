// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "show pending changes and packages that can be updated",
	Long:  `Show pending changes to the database and packages that can be updated.`,
}

func init() {
	StatusCmd.Run = status
}

// Status prints a thorough status of the current repository.
func status(cmd *cobra.Command, args []string) {
	if Conf.Unconfigured {
		dieOnError(ErrUnconfigured)
	}

	if len(args) > 0 {
		StatusCmd.Usage()
		os.Exit(1)
	}

	pkgs, outdated := getRepoPkgs(Conf.repodir)
	db, missed := getDatabasePkgs(Conf.Repository)

	name := Conf.database[:strings.IndexByte(Conf.database, '.')]
	fmt.Printf("On repo %s\n", name)

	// We assume that there is nothing to do, and if there is,
	// then this is set to false.
	var nothing = true

	if len(outdated) > 0 {
		printSet(mapPkgs(outdated, pkgFilename), "Obsolete packages to be removed/backed up:", Conf.Columnate)
		nothing = false
	}

	if len(missed) > 0 {
		printSet(mapPkgs(missed, pkgName), "Database entries pending removal:", Conf.Columnate)
		nothing = false
	}

	pending := filterPkgs(pkgs, dbPendingFilter(db))
	if len(pending) > 0 {
		printSet(mapPkgs(pending, pkgName), "Database entries pending addition:", Conf.Columnate)
		nothing = false
	}

	pkgs = removeIgnored(pkgs)

	aur, unavailable := getAURPkgs(mapPkgs(pkgs, pkgName))
	if len(unavailable) > 0 {
		printSet(unavailable, "Packages unavailable in AUR:", Conf.Columnate)
		nothing = false
	}

	updates := filterPkgs(pkgs, aurNewerFilter(aur))
	if len(updates) > 0 {
		printSet(mapPkgs(updates, pkgName), "Packages with updates in AUR:", Conf.Columnate)
		nothing = false
	}

	if nothing {
		fmt.Println("Everything up-to-date.")
	}
}
