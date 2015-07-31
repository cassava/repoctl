// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/goulash/pacman"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list packages that belong to the managed repository",
	Long: `List packages that belong to the managed repository.
Note that they don't need to be registered with the database.`,
}

var (
	// Versioned causes packages to be printed with version information.
	listVersioned bool
	// Mode can be either "count", "filter", or "mark" (which is the default
	// if no match is found.
	listMode string
	// Pending marks packages that need to be added to the database,
	// as well as packages that are in the database but are not available.
	listPending bool
	// Duplicates marks the number of obsolete packages for each package.
	listDuplicates bool
	// Installed marks whether packages are locally installed or not.
	listInstalled bool
	// Synchronize marks which packages have newer versions on AUR.
	listSynchronize bool
	// Same as all of the above.
	listAllOptions bool
)

func init() {
	ListCmd.Run = list

	ListCmd.Flags().BoolVarP(&listVersioned, "versioned", "v", false, "show package versions along with name")
	ListCmd.Flags().BoolVarP(&listPending, "pending", "p", false, "mark pending changes to the database")
	ListCmd.Flags().BoolVarP(&listDuplicates, "duplicates", "d", false, "mark packages with duplicate package files")
	ListCmd.Flags().BoolVarP(&listInstalled, "installed", "l", false, "mark packages that are locally installed")
	ListCmd.Flags().BoolVarP(&listSynchronize, "outdated", "o", false, "mark packages that are newer in AUR")
	ListCmd.Flags().BoolVarP(&listAllOptions, "all", "a", false, "all information; same as -vpdlo")
}

// list displays all the packages available for the database.
func list(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		ListCmd.Usage()
		os.Exit(1)
	}

	if listAllOptions {
		listVersioned = true
		listPending = true
		listDuplicates = true
		listInstalled = true
		listSynchronize = true
	}

	// Get packages in repo directory and database and find out duplicates.
	// Depinding on the configuration, some of this might be unecessary work,
	// so we have to declare the variabes here.
	var (
		db     map[string]*pacman.Package
		missed []*pacman.Package
		dups   map[string]int
		aur    map[string]*pacman.Package
	)
	pkgs, old := getRepoPkgs(Conf.repodir)
	if listPending {
		db, missed = getDatabasePkgs(Conf.Repository)
	}
	if listDuplicates {
		dups = make(map[string]int)
		for _, p := range old {
			dups[p.Name]++
		}
	}
	if listSynchronize {
		aur, _ = getAURPkgs(mapPkgs(pkgs, pkgName))
	}

	// Create a list.
	var pkgnames []string
	if listPending {
		for _, p := range missed {
			pkgnames = append(pkgnames, fmt.Sprintf("-%s-", p.Name))
		}
	}
	for _, p := range pkgs {
		buf := bytes.NewBufferString(p.Name)

		if listPending {
			dbp, ok := db[p.Name]
			if !ok || dbp.OlderThan(p) {
				buf.WriteRune('*')
			}
		}
		if listVersioned {
			buf.WriteRune(' ')
			buf.WriteString(p.Version)
		}
		if listSynchronize {
			ap := aur[p.Name]
			if ap == nil {
				buf.WriteString(" <?>")
			} else if ap.NewerThan(p) {
				if listVersioned {
					buf.WriteString(" -> ")
					buf.WriteString(ap.Version)
				} else {
					buf.WriteString(" <!>")
				}
			} else if ap.OlderThan(p) {
				if listVersioned {
					buf.WriteString(" <- ")
					buf.WriteString(ap.Version)
				} else {
					buf.WriteString(" <*>")
				}
			}
		}
		if listDuplicates && dups[p.Name] > 0 {
			buf.WriteString(fmt.Sprintf(" (%v)", dups[p.Name]))
		}

		pkgnames = append(pkgnames, buf.String())
	}

	// Print packages to stdout
	printSet(pkgnames, "", Conf.Columnate)
}
