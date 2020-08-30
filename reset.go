// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"github.com/goulash/osutil"
	"github.com/spf13/cobra"
)

func init() {
	MainCmd.AddCommand(resetCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "(Re-)create repository database",
	Long: `Delete the repository database and re-add all packages in repository.

  Essentially, this command deletes the repository database and recreates it by
  running the update command.

  If the repository does not exist yet, then it is initialized.
`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(0),
	ValidArgsFunction:     completeNoFiles,
	PreRunE:               ProfileInit,
	PostRunE:              ProfileTeardown,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create directory if necessary
		if ex, _ := osutil.DirExists(Repo.Directory); !ex {
			err := Repo.Setup()
			if err != nil {
				return err
			}
		}

		// Delete repository if it exists
		if ex, _ := osutil.Exists(Repo.DatabasePath()); ex {
			err := Repo.DeleteDatabase()
			if err != nil {
				return err
			}
		}

		// Create an empty database
		err := Repo.CreateDatabase()
		if err != nil {
			return err
		}

		// Populate the database with packages
		return Repo.Update(nil)
	},
}
