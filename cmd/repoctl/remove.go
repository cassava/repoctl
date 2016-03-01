// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

func init() {
	MainCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:     "remove <pkgname...>",
	Aliases: []string{"rm"},
	Short:   "remove and delete packages from the database",
	Long: `Remove and delete the package files from the repository.

  This command removes the specified package from the repository
  database, and deletes any associated package files, unless the backup
  option is given, in which case the package files are moved to the
  backup directory.
`,
	Example: `  repoctl rm fairsplit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Repo.Remove(nil, args...)
	},
}
