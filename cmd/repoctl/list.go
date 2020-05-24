// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"

	"github.com/goulash/pacman"
	"github.com/goulash/pacman/meta"
	"github.com/goulash/pr"
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
	// Only show registered packages.
	filterRegistered bool

	searchPOSIX bool
)

func init() {
	MainCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&filterRegistered, "registered", "r", false, "only show packages that are in the database")
	listCmd.Flags().BoolVarP(&listVersioned, "versioned", "v", false, "show package versions along with name")
	listCmd.Flags().BoolVarP(&listPending, "pending", "p", false, "mark pending changes to the database")
	listCmd.Flags().BoolVarP(&listDuplicates, "duplicates", "d", false, "mark packages with duplicate package files")
	listCmd.Flags().BoolVarP(&listInstalled, "installed", "l", false, "mark packages that are locally installed")
	listCmd.Flags().BoolVarP(&listSynchronize, "outdated", "o", false, "mark packages that are newer in AUR")
	listCmd.Flags().BoolVarP(&listAllOptions, "all", "a", false, "all information; same as -vpdlo")
	listCmd.Flags().BoolVar(&searchPOSIX, "posix", false, "use POSIX-style regular expressions")
}

var listCmd = &cobra.Command{
	Use:     "list [regex]",
	Aliases: []string{"ls"},
	Short:   "list packages that belong to the managed repository",
	Long: `List packages that belong to the managed repository.

  All packages that are in the managed repository are listed,
  whether or not they are registered with the database.
  If you only want to show registered packages, use the -r flag.
  
  When marking entries, the following symbols are used:

    -package-           package will be deleted
    package <?>         no AUR information could be found
    package <!>         local package is out-of-date
    package <*>         local package is newer than AUR package
    package (n)         there are n extra versions of package

  When versions are shown, local version is adjacent to package name:

    package 1.0 -> 2.0  local package is out-of-date
    package 2.0 <- 1.0  local package is newer than AUR package

  If a valid regular expression is supplied, only packages that match
  the expression will be listed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return &UsageError{"list", "list command takes at most one argument", cmd.Usage}
		}

		if listAllOptions {
			listVersioned = true
			listPending = true
			listDuplicates = true
			listInstalled = true
			listSynchronize = true
		}

		var regex *regexp.Regexp
		if len(args) == 1 {
			var err error
			if searchPOSIX {
				regex, err = regexp.Compile(args[0])
			} else {
				regex, err = regexp.CompilePOSIX(args[0])
			}
			if err != nil {
				return err
			}
		}

		pkgs, err := Repo.ListMeta(nil, listSynchronize, func(mp pacman.AnyPackage) string {
			p := mp.(*meta.Package)
			if regex != nil && !regex.MatchString(p.PkgName()) {
				return ""
			}

			if filterRegistered && !p.IsRegistered() {
				return ""
			}

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
					buf.WriteString(" <?>") // no aur info
				} else if pacman.PkgNewer(ap, p) {
					if listVersioned {
						buf.WriteString(" -> ") // new version
						buf.WriteString(ap.Version)
					} else {
						buf.WriteString(" <!>") // local version older than aur
					}
				} else if pacman.PkgOlder(ap, p) {
					if listVersioned {
						buf.WriteString(" <- ") // old version
						buf.WriteString(ap.Version)
					} else {
						buf.WriteString(" <*>") // local version newer than aur
					}
				}
			}
			if listDuplicates && len(p.Files)-1 > 0 {
				buf.WriteString(fmt.Sprintf(" (%v)", len(p.Files)-1))
			}

			return buf.String()
		})
		if err != nil {
			return err
		}

		// Print packages to stdout
		sort.Strings(pkgs)
		printSet(pkgs, "", Conf.Columnate)
		return nil
	},
}

// printSet prints a set of items and optionally a header.
func printSet(list []string, h string, cols bool) {
	if h != "" {
		fmt.Printf("\n%s\n", h)
	}
	if cols {
		pr.PrintFlex(list)
	} else if h != "" {
		for _, j := range list {
			fmt.Println(" ", j)
		}
	} else {
		for _, j := range list {
			fmt.Println(j)
		}
	}
}
