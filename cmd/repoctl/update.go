// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"github.com/goulash/pacman"
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:   "update [pkgname...]",
	Short: "update database in repository",
	Long: `Update database in repository by adding pending packages and
deleting obsolete packages.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// Update adds pending packages to the database and removes obsolete
// packages. If pkgnames is empty, the repository is scanned for
// whatever can be found.
//
// TODO: Update documentation and help text.
// TODO: Consolidate duplicate code.
func Update(pkgnames []string) error {
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
			return nil
		}
	}

	var err error
	if len(missed) > 0 {
		err = removePkgs(mapPkgs(missed, pkgName))
		if err != nil {
			return err
		}
	}
	if len(pending) > 0 {
		err = addPkgs(mapPkgs(pending, pkgFilename))
		if err != nil {
			return err
		}
	}
	if len(outdated) > 0 {
		filenames := mapPkgs(outdated, pkgBasename)
		if Conf.Backup {
			err = backupPkgs(filenames)
		} else {
			err = deletePkgs(filenames)
		}
		return err
	}
	return nil
}

// FIXME: the semantic of this function is wrong.
func UpdateAdd(pkgnames []string) error {
	// TODO: handle the errors here correctly!
	pkgs, _ := pacman.ReadMatchingNames(Conf.repodir, pkgfiles, nil)
	pkgs, outdated := pacman.SplitOld(pkgs)
	db, _ := getDatabasePkgs(Conf.Repository)
	pending := filterPkgs(pkgs, dbPendingFilter(db))

	if Conf.Interactive {
		info := "Delete following files:"
		if Conf.Backup {
			info = "Back following files up:"
		}
		proceed := confirmAll(
			[][]string{
				mapPkgs(pending, pkgNameVersion(db)),
				mapPkgs(outdated, pkgBasename),
			},
			[]string{
				"Add following entries to database:",
				info,
			},
			Conf.Columnate)
		if !proceed {
			return nil
		}
	}

	if len(pending) > 0 {
		err := addPkgs(mapPkgs(pending, pkgFilename))
		if err != nil {
			return err
		}
	}
	if len(outdated) > 0 {
		filenames := mapPkgs(outdated, pkgFilename)
		var err error
		if Conf.Backup {
			err = backupPkgs(filenames)
		} else {
			err = deletePkgs(filenames)
		}
		return err
	}

	return nil
}
