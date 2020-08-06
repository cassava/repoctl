// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

func init() {
	MainCmd.AddCommand(resetCmd)
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Recreate repository database",
	Long: `Delete the repository database and re-add all packages in repository.

  Essentially, this command deletes the repository database and
  recreates it by running the update command.
`,
	DisableFlagsInUseLine: true,
	Args:                  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return Repo.Reset(nil)
	},
}
