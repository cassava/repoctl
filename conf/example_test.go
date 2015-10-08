// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package conf_test

import (
	"fmt"
	"os"

	"github.com/cassava/repoctl"
	"github.com/cassava/repoctl/conf"
	"github.com/spf13/cobra"
)

var Repo *repoctl.Repo
var Conf *conf.Configuration

var repoctlCmd = &cobra.Command{
	Use:   "example",
	Short: "this command demonstrates how to use conf",
	Run: func(cmd *cobra.Command, args []string) {
		dieIfUnconfigured()
	},
}

func addConfFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&Conf.Backup, "backup", "b", Conf.Backup, "backup obsolete files instead of deleting")
	cmd.PersistentFlags().StringVarP(&Conf.BackupDir, "backup-dir", "B", Conf.BackupDir, "backup directory relative to repository path")
	cmd.PersistentFlags().BoolVarP(&Conf.Columnate, "columns", "s", Conf.Columnate, "show items in columns rather than lines")
	cmd.PersistentFlags().BoolVarP(&Conf.Quiet, "quiet", "q", Conf.Quiet, "show minimal amount of information")
	cmd.PersistentFlags().BoolVar(&Conf.Debug, "debug", Conf.Debug, "show unnecessary debugging information")
}

func dieIfUnconfigured() {
	if Conf.Unconfigured {
		fmt.Fprintln(os.Stderr, "Error: repoctl is unconfigured.")
		fmt.Fprintln(os.Stderr, "Please see: repoctl help new")
		os.Exit(1)
	}
}

// main loads the configuration and executes the primary command.
func Example() {
	Conf = conf.Default()
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
	addConfFlags(repoctlCmd)
	Repo = Conf.Repo()

	repoctlCmd.Execute()
}
