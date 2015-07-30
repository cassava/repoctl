// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

var DownCmd = &cobra.Command{
	Use:     "down [pkgname...]",
	Aliases: []string{"download"},
	Short:   "download and extract tarballs from AUR",
	Long: `Download and extract tarballs from AUR for given packages.
Alternatively, all packages, or those with updates can be downloaded.
Options specified are additive, not exclusive.

By default, tarballs are deleted after being extracted, and are placed
in the current directory.
`,
	Example: `down -u gets all updates that are found in AUR`,
	Run:     down,
}

var (
	downUpdates   bool
	downAll       bool
	downNoExtract bool
	downRewrite   bool
	downOutput    string
)

func init() {
	downCmd.Flags().BoolVarP(&downUpdates, "updates", "u", false, "download tarballs for all updates")
	downCmd.Flags().BoolVarP(&downAll, "all", "a", false, "download tarballs for all packages in database")
	downCmd.Flags().BoolVarP(&downNoExtract, "no-extract", "n", false, "do not extract the tarballs")
	downCmd.Flags().BoolVarP(&downRewrite, "rewrite", "t", false, "delete conflicting folders")
	downCmd.Flags().StringVarP(&downOutput, "output", "o", "", "output directory for tarballs")
}

func down(cmd *cobra.Command, args []string) {
	panic("not implemented")
}
