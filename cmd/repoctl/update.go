// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

func init() {
	RepoctlCmd.AddCommand(updateCmd)
}

var updateCmd = &cobra.Command{
	Use:   "update [pkgname...]",
	Short: "update database in repository",
	Long: `Update database by adding packages and dispatching obsolete files.

  Package files that are newer than the registered versions in the
  database are added to the database; entries in the database which
  have no files are removed.

  If no package names are given, the entire repository is scanned for
  updates.
`,
	Example: `  repoctl update fairsplit`,
	Run: func(cmd *cobra.Command, args []string) {
		err := Repo.Update(nil, args...)
		dieOnError(err)
	},
}
