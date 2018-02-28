// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

func init() {
	MainCmd.AddCommand(resetCmd)

	resetCmd.Flags().BoolVarP(&keepCacheFiles, "keep-cache", "k", false, "keep cache files untouched")
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "recreate repository database",
	Long: `Delete the repository database and re-add all packages in repository.

  Essentially, this command deletes the repository database and
  recreates it by running the update command.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Repo.Reset(nil, keepCacheFiles)
	},
}
