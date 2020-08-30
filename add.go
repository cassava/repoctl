// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

var movePackages bool
var linkPackages bool
var addRequireSignature bool

func init() {
	MainCmd.AddCommand(addCmd)

	addCmd.Flags().BoolVarP(&movePackages, "move", "m", false, "move packages into repository")
	addCmd.Flags().BoolVarP(&linkPackages, "link", "l", false, "link packages instead of copying")
	addCmd.Flags().BoolVarP(&addRequireSignature, "require-signature", "r", false, "require package signatures")
}

var addCmd = &cobra.Command{
	Use:   "add PKGFILE ...",
	Short: "Copy and add packages to the repository",
	Long: `Add (and copy if necessary) the package file to the repository.

  Similarly to the repo-add script, this command copies the package
  file to the repository (if not already there) and adds the package to
  the database. Exactly this package is added to the database, this
  allows you to downgrade a package in the repository.

  Any other package files in the repository are deleted or backed up,
  depending on whether the backup option is set. If the backup option
  is set, the "obsolete" package files are moved to a backup
  directory of choice.

  If the backup directory resolves to the repository directory,
  then obsolete package files are ignored.
  You can specify --backup=false to force them to be deleted.
`,
	Example:           `  repoctl add -m ./fairsplit-1.0.pkg.tar.gz`,
	ValidArgsFunction: completeLocalPackageFiles,
	PreRunE:           ProfileInit,
	PostRunE:          ProfileTeardown,
	RunE: func(cmd *cobra.Command, args []string) error {
		if addRequireSignature {
			Repo.RequireSignature = true
		}

		if movePackages {
			return Repo.Move(nil, args...)
		}
		if linkPackages {
			return Repo.Link(nil, args...)
		}
		return Repo.Copy(nil, args...)
	},
}
