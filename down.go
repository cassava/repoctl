// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman/aur"
	"github.com/cassava/repoctl/pacman/graph"
	"github.com/cassava/repoctl/pacman/pkgutil"
	"github.com/cassava/repoctl/repo"
	"github.com/spf13/cobra"
)

var (
	downDest     string
	downDryRun   bool
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
	downCmd.RegisterFlagCompletionFunc("dest", completeDirectory)

	downCmd.Flags().BoolVarP(&downDryRun, "dry-run", "n", false, "don't download any packages")
	downCmd.Flags().BoolVarP(&downClobber, "clobber", "l", false, "delete conflicting files and folders")
	downCmd.Flags().BoolVarP(&downExtract, "extract", "e", true, "extract the downloaded tarballs")
	downCmd.Flags().BoolVarP(&downUpgrades, "upgrades", "u", false, "download tarballs for all upgrades")
	downCmd.Flags().BoolVarP(&downRecurse, "recursive", "r", false, "download any necessary dependencies")
	downCmd.Flags().StringVarP(&downOrder, "order", "o", "", "write the order of compilation based on dependency tree into a file, implies -r")
	downCmd.Flags().BoolVarP(&downAll, "all", "a", false, "download tarballs for all packages in database")
}

var downCmd = &cobra.Command{
	Use:     "down [PKGNAME ...]",
	Aliases: []string{"download"},
	Short:   "Download and extract tarballs from AUR",
	Long: `Download and extract tarballs from AUR for given packages.

  Alternatively, all packages, or those with updates can be downloaded.
  Options specified are additive, not exclusive.

  By default, tarballs are deleted after being extracted, and are placed
  in the current directory.

  Packages can also be downloaded recursively, and the list that these
  dependencies should be built can be saved. For example, to download
  all updates to the repository and build them in approximately the
  correct order:

    repoctl down -o build-order.txt -u
    for pkg in $(cat build-order.txt); do
        (
            cd $pkg
            makepkg -si
            ok=$?
            if $ok; then
                repoctl add -m *.pkg.tar*
                cd ..
                rm -rf $pkg
            fi
        )
    done

  You can just output the correct build order by adding the -n flag to
  prevent downloading of tarballs.

  Caveats:

  1. Automatic dependency resolution does not currently handle version
     resolution or library specifications, as noted in the Arch wiki at:
       https://wiki.archlinux.org/index.php/PKGBUILD#Dependencies

  2. Package dependencies are not resolved that are only "provided"
     by other packages. Here, we currently print an "unknown package" warning.

	 For example, at the time of writing firefox56 requires mime-types.
	 This package does not exist, but is provided by other packages.
	 We can check this with:
	   repoctl query $(repoctl search -q mime-types)
	 Which leads us to see that mailcap-mime-types provides mime-types.
	 This caveat will be resolved in the future.
`,
	Example: `  repoctl down -u
  repoctl down -o build-order.txt -u`,
	ValidArgsFunction: completeAURPackageNames,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if downAll || downUpgrades {
			return ProfileInit(cmd, args)
		}
		return nil
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		if downAll || downUpgrades {
			return ProfileTeardown(cmd, args)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// First, populate the initial list of packages to download.
		var list []string
		if downAll {
			names, err := Repo.ReadAllNames(nil)
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
			// There's not much point to a try run here, but we should respect
			// the option nevertheless.
			if downDryRun {
				return nil
			}
			return repo.Download(downDest, downExtract, downClobber, list)
		}

		// Otherwise, get the dependency list and download the packages:
		aps, err := downDependencies(list)
		if err != nil {
			return err
		}
		// Don't download any packages if dry run is activated.
		if downDryRun {
			return nil
		}
		return repo.DownloadPackages(aps, downDest, downExtract, downClobber)
	},
}

func downDependencies(packages []string) (aur.Packages, error) {
	g, err := repo.DependencyGraph(packages)
	if err != nil {
		return nil, err
	}
	_, aps, ups := graph.Dependencies(g)
	if downOrder != "" {
		term.Debugf("Writing build-order to: %s", downOrder)
		f, err := os.Create(downOrder)
		if err != nil {
			return nil, fmt.Errorf("cannot write build-order to %s: %w", downOrder, err)
		}

		for i := len(aps); i != 0; i-- {
			fmt.Fprintln(f, aps[i-1].Name)
		}
		f.Close()
	}
	for _, u := range ups {
		term.Warnf("Warning: unknown package %s\n", u)
		iter := g.To(g.NodeWithName(u).ID())
		for iter.Next() {
			node := iter.Node().(*graph.Node)
			term.Warnff("         Required by: %s\n", node.PkgName())
		}
	}
	return aps, nil
}
