// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/cassava/repoctl"
	"github.com/cassava/repoctl/conf"
	"github.com/goulash/color"
	"github.com/spf13/cobra"
)

// Repo lets us use the repoctl library to do the most of the work.
var Repo *repoctl.Repo

// Conf loads and stores the configuration (apart from command line
// configuration) of this program, including where the repository is.
var Conf = conf.Default()

// col lets us print in colors.
var col = color.New()

var RepoctlCmd = &cobra.Command{
	Use:   "repoctl",
	Short: "manage local Pacman repositories",
	Long: `Repoctl helps manage local Pacman repositories, and acts in particular as
a supplement to the repo-add and repo-remove tools that come with Pacman.

Whether compiling and installing from AUR every time is not what you want,
or if you want to host your own repository, repoctl is right for you.

Note that in all of these commands, the following terminology is used:

    pkgname: is the name of the package, e.g. pacman
    pkgfile: is the path to a package file, e.g. pacman-3.5.3-i686.pkg.tar.xz
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This function can be overriden if it's not necessary for a command.
		if Conf.Unconfigured {
			fmt.Fprintln(os.Stderr, "Error: repoctl is unconfigured, please create configuration.")
			os.Exit(1)
		}
		Repo = Conf.Repo()
	},
}

func addConfFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&Conf.Backup, "backup", "b", Conf.Backup, "backup obsolete files instead of deleting")
	cmd.PersistentFlags().StringVarP(&Conf.BackupDir, "backup-dir", "B", Conf.BackupDir, "backup directory relative to repository path")
	cmd.PersistentFlags().BoolVarP(&Conf.Columnate, "columns", "s", Conf.Columnate, "show items in columns rather than lines")
	cmd.PersistentFlags().BoolVarP(&Conf.Quiet, "quiet", "q", Conf.Quiet, "show minimal amount of information")
	cmd.PersistentFlags().BoolVar(&Conf.Debug, "debug", Conf.Debug, "show unnecessary debugging information")
	col.Set(Conf.Color) // set default, which will be auto if Conf.Color is empty or invalid
	cmd.PersistentFlags().Var(col, "color", "when to use color (auto|never|always)")
}

// main loads the configuration and executes the primary command.
func main() {
	err := Conf.MergeAll()
	if err != nil {
		// We didn't manage to load any configuration, which means that repoctl
		// is unconfigured. There are some commands that work nonetheless, so
		// we can't stop now -- which is why we don't os.Exit(1).
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
	}

	// Arguments from the command line override the configuration file,
	// so we have to add the flags after loading the configuration.
	//
	// TODO: Maybe in the future we will make it possible to specify the
	// configuration file via the command line; right now it is not a priority.
	addConfFlags(RepoctlCmd)

	RepoctlCmd.Execute()
}
