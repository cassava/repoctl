// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman/aur"
	"github.com/spf13/cobra"
)

var (
	statusAUR     bool
	statusMissing bool
	statusCached  bool
)

func init() {
	MainCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&statusAUR, "aur", "a", false, "check AUR for upgrades")
	statusCmd.Flags().BoolVarP(&statusMissing, "missing", "m", false, "highlight packages missing in AUR")
	statusCmd.Flags().BoolVarP(&statusCached, "cached", "c", false, "show how many old package files are cached")
}

var statusCmd = &cobra.Command{
	Use:   "status [--aur]",
	Short: "Show pending changes and packages that can be upgraded",
	Long: `Show pending changes to the database and packages that can be updated.

  In particular, the following is shown:

    "updated":  database entries that can be updated/added (new package files)
    "obsolete": package files that can be deleted (or backed up)
    "cached":   package files that are cached (contrary to obsolete)
    "removal":  database entries that should be deleted (no package files)
    "upgrade":  packages with updates in AUR (only with -a)
    "!aur":     packages unavailable in AUR (only with -m)
`,
	Args:              cobra.ExactArgs(0),
	ValidArgsFunction: completeNoFiles,
	PreRunE:           ProfileInit,
	PostRunE:          ProfileTeardown,
	RunE: func(cmd *cobra.Command, args []string) error {
		exceptQuiet()
		term.Printf("On repo @{!y}%s\n\n", Repo.Name())

		pkgs, err := Repo.ReadMeta(nil)
		if err != nil {
			return err
		}
		ignore := Repo.IgnoreMap()
		if statusAUR || statusMissing {
			err = pkgs.ReadAUR()
			if err != nil && !aur.IsNotFound(err) {
				return err
			}
		}

		// We assume that there is nothing to do, and if there is,
		// then this is set to false.
		var nothing = true

		for _, p := range pkgs {
			var flags []string
			if p.HasUpgrade() && !ignore[p.Name] {
				flags = append(flags, term.Formatter.Sprintf("@gupgrade(@|%s -> %s@g)", p.Version(), p.AUR.Version))
			}
			if p.HasUpdate() {
				flags = append(flags, term.Formatter.Sprintf("@gupdated(@|%s -> %s@g)", p.VersionRegistered(), p.Version()))
			}
			if !p.HasFiles() {
				flags = append(flags, term.Formatter.Sprint("@rremoval"))
			}
			if o := p.Obsolete(); len(o) > 0 {
				if Repo.IsObsoleteCached() {
					if statusCached {
						flags = append(flags, term.Formatter.Sprintf("@ycached(@|%d@y)", len(o)))
					}
				} else {
					flags = append(flags, term.Formatter.Sprintf("@yobsolete(@|%d@y)", len(o)))
				}
			}
			if statusMissing && p.AUR == nil && !ignore[p.Name] {
				flags = append(flags, term.Formatter.Sprint("@y!aur"))
			}

			if len(flags) > 0 {
				nothing = false
				term.Printf("    %s:", p.Name)
				for _, f := range flags {
					term.Printf(" %s", f)
				}
				term.Println()
			}
		}

		if nothing {
			term.Printf("Everything up-to-date.\n")
		}
		return nil
	},
}
