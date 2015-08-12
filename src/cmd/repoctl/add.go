// Copyright (Conf) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"github.com/goulash/pacman"
	"github.com/spf13/cobra"
)

var movePackages bool

func init() {
	AddCmd.Flags().BoolVarP(&movePackages, "move", "m", false, "move packages into repository")
}

var AddCmd = &cobra.Command{
	Use:   "add <pkgfile...>",
	Short: "copy and add packages to the repository",
	Long: `Add (and copy if necessary) the package file to the repository.

  Similarly to the repo-add script, this command copies the package
  file to the repository (if not already there) and adds the package to
  the database.  Exactly this package is added to the database, this
  allows you to downgrade a package in the repository.

  Any other package files in the repository are deleted or backed up,
  depending on whether the backup option is given. If the backup option
  is given, the "obsolete" package files are moved to a backup
  directory of choice.

  Note: since version 0.14, the semantic meaning of this command has
        changed. See the update command for the old behavior.
`,
	Example: `  repoctl add ./fairsplit-1.0.pkg.tar.gz`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if movePackages {
			err = Move(args)
		} else {
			err = Add(args)
		}
		dieOnError(err)
	},
}

// FIXME: implement me!
func Move(pkgfiles []string) error {
	return nil
}

// FIXME: the semantic of this function is wrong.
func Add(pkgfiles []string) error {
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
