// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/goulash/pacman/aur"
	"github.com/spf13/cobra"
)

var (
	statusAUR     bool
	statusMissing bool
)

func init() {
	MainCmd.AddCommand(statusCmd)

	statusCmd.Flags().BoolVarP(&statusAUR, "aur", "a", false, "check AUR for upgrades")
	statusCmd.Flags().BoolVarP(&statusMissing, "missing", "m", false, "highlight packages missing in AUR")
}

var statusCmd = &cobra.Command{
	Use:   "status [--aur]",
	Short: "show pending changes and packages that can be upgraded",
	Long: `Show pending changes to the database and packages that can be updated.

  In particular, the following is shown:

    - obsolete package files that can be deleted (or backed up)
    - database entries that should be deleted (no package files)
    - database entries that can be updated/added (new package files)
    - packages unavailable in AUR
    - packages with updates in AUR
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return &UsageError{"status", "status command takes no arguments", cmd.Usage}
		}

		col.Printf("On repo @{!y}%s\n\n", Repo.Name())

		pkgs, err := Repo.ReadMeta(nil)
		if err != nil {
			return err
		}
		ignore := Repo.IgnoreMap()
		if statusAUR {
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
				flags = append(flags, col.Sprintf("@gupgrade(@|%s->%s@g)", p.Version(), p.AUR.Version))
			}
			if p.HasUpdate() {
				flags = append(flags, col.Sprintf("@gupdate(@|%s->%s@g)", p.VersionRegistered(), p.Version()))
			}
			if !p.HasFiles() {
				flags = append(flags, col.Sprint("@rremoval"))
			}
			if o := p.Obsolete(); len(o) > 0 {
				flags = append(flags, col.Sprintf("@yobsolete(@|%d@y)", len(o)))
			}
			if statusMissing && p.AUR == nil && !ignore[p.Name] {
				flags = append(flags, col.Sprint("@y!aur"))
			}

			if len(flags) > 0 {
				nothing = false
				fmt.Printf("\t%s:", p.Name)
				for _, f := range flags {
					fmt.Printf(" %s", f)
				}
				fmt.Println()
			}
		}

		if nothing {
			fmt.Println("Everything up-to-date.")
		}
		return nil
	},
}
