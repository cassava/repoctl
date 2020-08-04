// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"sort"

	"github.com/cassava/repoctl/pacman/aur"
	"github.com/spf13/cobra"
)

var (
	searchSortBy string
	searchQuiet  bool
)

func init() {
	MainCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchSortBy, "sort-by", "name", "which key to sort results by")
	searchCmd.Flags().BoolVarP(&searchQuiet, "quiet", "q", false, "show only the name")
}

var searchCmd = &cobra.Command{
	Use:   "search [pkgname...]",
	Short: "search for packages on AUR",
	Long: `Search for packages hosted on AUR.

  This command searches the specified arguments on AUR by the name property.
  The results are combined and sorted by one of the following methods:

    name
    votes
    popularity
    votes-reverse
    popularity-reverse

  The default is "name". Duplicate results are filtered from the output.
  Note: This command is currently somewhat experimental; the flags may be
  subject to change.
`,
	Example: `  repoctl search --sort-by=votes firefox
  repoctl search flir flirc flirc-bin`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Prevent errors that we print being printed a second time by cobra.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var pkgs aur.Packages
		for _, q := range args {
			aurpkgs, err := aur.SearchByName(q)
			if err != nil {
				return err
			}
			pkgs = append(pkgs, aurpkgs...)
		}

		// Sort the list
		if searchSortBy == "name" {
			sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })
		} else if searchSortBy == "votes" {
			sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].NumVotes < pkgs[j].NumVotes })
		} else if searchSortBy == "votes-reverse" {
			sort.Slice(pkgs, func(i, j int) bool { return pkgs[j].NumVotes < pkgs[i].NumVotes })
		} else if searchSortBy == "popularity" {
			sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Popularity < pkgs[j].Popularity })
		} else if searchSortBy == "popularity-reverse" {
			sort.Slice(pkgs, func(i, j int) bool { return pkgs[j].Popularity < pkgs[i].Popularity })
		} else {
			return fmt.Errorf("unknown sort-by key '%s'", searchSortBy)
		}

		// Print the list
		var pkgnames []string
		pkgset := make(map[string]bool)
		for _, p := range pkgs {
			// Only add unique names to the list of packages
			if pkgset[p.Name] {
				continue
			}
			pkgset[p.Name] = true

			var s string
			if searchQuiet {
				s = p.Name
			} else {
				s = col.Sprintf("@{!m}aur/@{!w}%s @{!g}%s @{r}(%d)\n@|    %s", p.Name, p.Version, p.NumVotes, p.Description)
			}
			pkgnames = append(pkgnames, s)
		}
		if searchQuiet {
			printSet(pkgnames, "", Conf.Columnate)
		} else {
			for _, p := range pkgnames {
				fmt.Println(p)
			}
		}

		return nil
	},
}
