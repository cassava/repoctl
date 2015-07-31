// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:   "update [pkgname...]",
	Short: "update database in repository",
	Long: `Update database in repository by adding pending packages and
deleting obsolete packages.`,
	Run: update,
}

func update(cmd *cobra.Command, args []string) {
	pkgs, outdated := getRepoPkgs(Conf.repodir)
	db, missed := getDatabasePkgs(Conf.Repository)
	pending := filterPkgs(pkgs, dbPendingFilter(db))

	if Conf.Interactive {
		info := "Delete following files:"
		if Conf.Backup {
			info = "Back following files up:"
		}
		proceed := confirmAll(
			[][]string{
				mapPkgs(missed, pkgName),
				mapPkgs(pending, pkgNameVersion(db)),
				mapPkgs(outdated, pkgBasename),
			},
			[]string{
				"Remove following entries from database:",
				"Update following entries in database:",
				info,
			},
			Conf.Columnate)
		if !proceed {
			os.Exit(0)
		}
	}

	var err error
	if len(missed) > 0 {
		err = removePkgs(mapPkgs(missed, pkgName))
		dieOnError(err)
	}
	if len(pending) > 0 {
		err = addPkgs(mapPkgs(pending, pkgFilename))
		dieOnError(err)
	}
	if len(outdated) > 0 {
		filenames := mapPkgs(outdated, pkgBasename)
		if Conf.Backup {
			err = backupPkgs(filenames)
		} else {
			err = deletePkgs(filenames)
		}
		dieOnError(err)
	}
}
