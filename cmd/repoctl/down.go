// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"github.com/goulash/pacman/pkgutil"
	"github.com/spf13/cobra"
)

var (
	downDest     string
	downClobber  bool
	downExtract  bool
	downUpgrades bool
	downAll      bool
)

func init() {
	RepoctlCmd.AddCommand(downCmd)

	downCmd.Flags().StringVarP(&downDest, "dest", "d", "", "output directory for tarballs")
	downCmd.Flags().BoolVarP(&downClobber, "clobber", "l", false, "delete conflicting files and folders")
	downCmd.Flags().BoolVarP(&downExtract, "extract", "e", true, "extract the downloaded tarballs")
	downCmd.Flags().BoolVarP(&downUpgrades, "upgrades", "u", false, "download tarballs for all upgrades")
	downCmd.Flags().BoolVarP(&downAll, "all", "a", false, "download tarballs for all packages in database")
}

var downCmd = &cobra.Command{
	Use:     "down [pkgname...]",
	Aliases: []string{"download"},
	Short:   "download and extract tarballs from AUR",
	Long: `Download and extract tarballs from AUR for given packages.
Alternatively, all packages, or those with updates can be downloaded.
Options specified are additive, not exclusive.

By default, tarballs are deleted after being extracted, and are placed
in the current directory.
`,
	Example: `  repoctl down -u`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if downAll {
			names, err := Repo.ReadNames(nil)
			dieOnError(err)
			err = Repo.Download(nil, downDest, downExtract, downClobber, pkgutil.Map(names, pkgutil.PkgName)...)
		} else if downUpgrades {
			err = Repo.DownloadUpgrades(nil, downDest, downExtract, downClobber, args...)
		} else {
			err = Repo.Download(nil, downDest, downExtract, downClobber, args...)
		}
		dieOnError(err)
	},
}
