// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

func Update(c *Config) error {
	pkgs, outdated := getRepoPkgs(c.path)
	db, missed := getDatabasePkgs(c.Repository)
	pending := filterPkgs(pkgs, dbPendingFilter(db))

	if Interactive {
		backup := "Delete following files:"
		if Backup {
			backup = "Back following files up:"
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
				backup,
			},
			Columnate)
		if !proceed {
			return nil
		}
	}

	var err error
	if len(missed) > 0 {
		err = removePkgs(c, mapPkgs(missed, pkgName))
		if err != nil {
			return err
		}
	}
	if len(pending) > 0 {
		err = addPkgs(c, mapPkgs(pending, pkgFilename))
		if err != nil {
			return err
		}
	}
	if len(outdated) > 0 {
		filenames := mapPkgs(outdated, pkgBasename)
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
