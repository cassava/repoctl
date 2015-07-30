// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"github.com/goulash/pacman"
	"github.com/spf13/cobra"
)

var AddCmd = &cobra.Command{
	Use:     "add <pkgfile...>",
	Aliases: []string{"copy"},
	Short:   "copy and add packages to the database",
	Long: `Add (and copy if necessary) the package file to the repository.
All obsolete package files in the repository are deleted.
If the backup option is given, obsolete package files are backed up
in a separate (specified) directory instead of being deleted.
`,
	Example: `Copy and add the fairsplit-1.0.pkg.tar.gz file to database:
    repoctl add ./fairsplit-1.0.pkg.tar.gz`,
	Run: add,
}

// FIXME: This will not compile, because we haven't fixed the global arguments
// problem yet!
//
// Note that the semantics of this function have changed! I figured that add
// implies adding a (new) package file to the database, which is not what it
// does. The old behavior of add will be covered by update.
func add(cmd *cobra.Command, args []string) {
	// TODO: handle the errors here correctly!
	pkgs, _ := pacman.ReadMatchingNames(c.path, c.Args, nil)
	pkgs, outdated := pacman.SplitOld(pkgs)
	db, _ := getDatabasePkgs(c.Repository)
	pending := filterPkgs(pkgs, dbPendingFilter(db))

	if c.Interactive {
		backup := "Delete following files:"
		if c.Backup {
			backup = "Back following files up:"
		}
		proceed := confirmAll(
			[][]string{
				mapPkgs(pending, pkgNameVersion(db)),
				mapPkgs(outdated, pkgBasename),
			},
			[]string{
				"Add following entries to database:",
				backup,
			},
			c.Columnate)
		if !proceed {
			return nil
		}
	}

	var err error
	if len(pending) > 0 {
		err = addPkgs(c, mapPkgs(pending, pkgFilename))
		if err != nil {
			return err
		}
	}
	if len(outdated) > 0 {
		filenames := mapPkgs(outdated, pkgFilename)
		if c.Backup {
			err = backupPkgs(c, filenames)
		} else {
			err = deletePkgs(c, filenames)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
