// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"github.com/goulash/pacman"
	"github.com/spf13/cobra"
)

var RemoveCmd = &cobra.Command{
	Use:     "remove <pkgfile...>",
	Aliases: []string{"rm"},
	Short:   "remove and dele tepackages from the database",
	Long: `Add (and copy if necessary) the package file to the repository.
All obsolete package files in the repository are deleted.
If the backup option is given, obsolete package files are backed up
in a separate (specified) directory instead of being deleted.
`,
	Run: remove,
}

func remove(cmd *cobra.Command, args []string) {
	// TODO: handle the errors here correctly!
	pkgs, _ := pacman.ReadMatchingNames(c.path, c.Args, nil)
	db, _ := getDatabasePkgs(c.Repository)

	rmmap := make(map[string]bool)
	for _, p := range pkgs {
		rmmap[p.Name] = true
	}
	dbpkgs := make([]string, 0, len(rmmap))
	for k := range rmmap {
		if _, ok := db[k]; ok {
			dbpkgs = append(dbpkgs, k)
		}
	}

	if c.Interactive {
		backup := "Delete following files:"
		if c.Backup {
			backup = "Back following files up:"
		}
		proceed := confirmAll(
			[][]string{
				dbpkgs,
				mapPkgs(pkgs, pkgBasename),
			},
			[]string{
				"Remove following entries from database:",
				backup,
			},
			c.Columnate)
		if !proceed {
			return nil
		}
	}

	var err error
	if len(dbpkgs) > 0 {
		err = removePkgs(c, dbpkgs)
		if err != nil {
			return err
		}
	}
	if len(pkgs) > 0 {
		files := mapPkgs(pkgs, pkgFilename)
		if c.Backup {
			err = backupPkgs(c, files)
		} else {
			err = deletePkgs(c, files)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
