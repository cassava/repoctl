// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	MainCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:     "remove <pkgname...>",
	Aliases: []string{"rm"},
	Short:   "Remove and delete packages from the database",
	Long: `Remove and delete the package files from the repository.

  This command removes the specified package from the repository
  database, and deletes any associated package files, unless the backup
  option is given, in which case the package files are moved to the
  backup directory.

  If the backup directory resolves to the repository directory,
  then package files are ignored; repoctl update will add them again.
  In this case, you probably want to use --backup=false to force
  them to be deleted.
`,
	Example: `  repoctl rm fairsplit
  repoctl rm --backup=false fairsplit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if Repo.Backup && Repo.IsObsoleteCached() {
			fmt.Fprintf(os.Stderr, "warning: removing only database entries, use --backup=false to delete package files.\n")
		}
		return Repo.Remove(nil, args...)
	},
}
