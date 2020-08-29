// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

var updateRequireSignature bool

func init() {
	MainCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVarP(&updateRequireSignature, "require-signature", "r", false, "require package signatures")
}

var updateCmd = &cobra.Command{
	Use:   "update [pkgname...]",
	Short: "Update database in repository",
	Long: `Update database by adding packages and dispatching obsolete files.

  Package files that are newer than the registered versions in the
  database are added to the database; entries in the database which
  have no files are removed.

  If no package names are given, the entire repository is scanned for
  updates.

  If backup is true, obsolete files are backup up instead of deleted.
  If the backup directory resolves to the repository directory,
  then obsolete package files are ignored.
  You can specify --backup=false to force them to be deleted.
`,
	Example: `  repoctl update fairsplit
  repoctl update --backup=false`,
	ValidArgsFunction: completeRepoPackageNames,
	PreRunE:           ProfileInit,
	PostRunE:          ProfileTeardown,
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateRequireSignature {
			Repo.RequireSignature = true
		}

		return Repo.Update(nil, args...)
	},
}
