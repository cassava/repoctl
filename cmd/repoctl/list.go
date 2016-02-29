// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/goulash/pacman"
	"github.com/goulash/pacman/meta"
	"github.com/spf13/cobra"
)

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
	MainCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&listVersioned, "versioned", "v", false, "show package versions along with name")
	listCmd.Flags().BoolVarP(&listPending, "pending", "p", false, "mark pending changes to the database")
	listCmd.Flags().BoolVarP(&listDuplicates, "duplicates", "d", false, "mark packages with duplicate package files")
	listCmd.Flags().BoolVarP(&listInstalled, "installed", "l", false, "mark packages that are locally installed")
	listCmd.Flags().BoolVarP(&listSynchronize, "outdated", "o", false, "mark packages that are newer in AUR")
	listCmd.Flags().BoolVarP(&listAllOptions, "all", "a", false, "all information; same as -vpdlo")
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "list packages that belong to the managed repository",
	Long: `List packages that belong to the managed repository.

  Note that they don't need to be registered with the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			cmd.Usage()
			os.Exit(1)
		}

		if listAllOptions {
			listVersioned = true
			listPending = true
			listDuplicates = true
			listInstalled = true
			listSynchronize = true
		}

		pkgs, err := Repo.ListMeta(nil, listSynchronize, func(mp pacman.AnyPackage) string {
			p := mp.(*meta.Package)
			if listPending && !p.HasFiles() {
				return fmt.Sprintf("-%s-", p.Name)
			}

			buf := bytes.NewBufferString(p.Name)
			if listPending && p.HasUpdate() {
				buf.WriteRune('*')
			}
			if listVersioned {
				buf.WriteRune(' ')
				buf.WriteString(p.Version())
			}
			if listSynchronize {
				ap := p.AUR
				if ap == nil {
					buf.WriteString(" <?>")
				} else if pacman.PkgNewer(ap, p) {
					if listVersioned {
						buf.WriteString(" -> ")
						buf.WriteString(ap.Version)
					} else {
						buf.WriteString(" <!>")
					}
				} else if pacman.PkgOlder(ap, p) {
					if listVersioned {
						buf.WriteString(" <- ")
						buf.WriteString(ap.Version)
					} else {
						buf.WriteString(" <*>")
					}
				}
			}
			if listDuplicates && len(p.Files)-1 > 0 {
				buf.WriteString(fmt.Sprintf(" (%v)", len(p.Files)-1))
			}

			return buf.String()
		})
		dieOnError(err)

		// Print packages to stdout
		printSet(pkgs, "", Conf.Columnate)
	},
}
