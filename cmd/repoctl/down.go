// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/goulash/pacman/graph"
	"github.com/goulash/pacman/pkgutil"
	"github.com/spf13/cobra"
)

var (
	downDest     string
	downClobber  bool
	downExtract  bool
	downUpgrades bool
	downAll      bool
	downRecurse  bool
	downOrder    string
)

func init() {
	MainCmd.AddCommand(downCmd)

	downCmd.Flags().StringVarP(&downDest, "dest", "d", "", "output directory for tarballs")
	downCmd.Flags().BoolVarP(&downClobber, "clobber", "l", false, "delete conflicting files and folders")
	downCmd.Flags().BoolVarP(&downExtract, "extract", "e", true, "extract the downloaded tarballs")
	downCmd.Flags().BoolVarP(&downUpgrades, "upgrades", "u", false, "download tarballs for all upgrades")
	downCmd.Flags().BoolVarP(&downRecurse, "recursive", "r", false, "download any necessary dependencies")
	downCmd.Flags().StringVarP(&downOrder, "order", "o", "", "write the order of compilation based on dependency tree into a file, implies -r")
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// First, populate the initial list of packages to download.
		var list []string
		if downAll {
			names, err := Repo.ReadNames(nil)
			if err != nil {
				return err
			}
			list = pkgutil.Map(names, pkgutil.PkgName)
		} else if downUpgrades {
			upgrades, err := Repo.FindUpgrades(nil, args...)
			if err != nil {
				return err
			}
			for _, u := range upgrades {
				list = append(list, u.New.Name)
			}
		} else {
			list = args
		}

		// If no dependencies are wanted, then get to it right away:
		if !downRecurse && downOrder == "" {
			return Repo.Download(nil, downDest, downExtract, downClobber, list...)
		}

		// Otherwise, get the dependencies:
		g, err := Repo.DependencyGraph(nil, list...)
		if err != nil {
			return err
		}
		_, aps, ups := graph.Dependencies(g)
		if downOrder != "" {
			f, err := os.Create(downOrder)
			if err != nil {
				return err
			}
			for _, p := range aps {
				fmt.Fprintln(f, p.Name)
			}
			f.Close()
		}
		for _, u := range ups {
			fmt.Fprintf(os.Stderr, "warning: unknown package %s\n", u)
		}
		return Repo.DownloadPackages(nil, aps, downDest, downExtract, downClobber)
	},
}
