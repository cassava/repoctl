// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/goulash/osutil"
	"github.com/spf13/cobra"
)

var (
	ErrInvalidConfPath = errors.New("invalid configuration path")
)

// TODO: I'm not sure I like the interface of this command. Do I really
// need subcommands? Would it not make more sense to have configuration
// creation happen with repository creation?
//
// Also, profile creation?
var NewCmd = &cobra.Command{
	Use:   "new [command] [flags]",
	Short: "create a new repository or configuration file",
	Long: `Create either a new repository or configuration file.
Paths will be created as necessary.
`,
}

var (
	nConf string // path to configuration
)

func init() {
	NewCmd.PersistentFlags().StringVarP(&nConf, "config", "c", HomeConf(), "path to configuration file")
}

func init() {
	NewCmd.AddCommand(newRepoCmd)
	NewCmd.AddCommand(newConfigCmd)
}

var newRepoCmd = &cobra.Command{
	Use:   "repo </path/to/repo/database>",
	Short: "create a new repository",
	Long:  `Create a new repository with configuration file.`,
	Run:   newRepo,
}

func newRepo(cmd *cobra.Command, args []string) {
	panic("not implemented")
}

var newConfigCmd = &cobra.Command{
	Use:   "config </path/to/repo/database>",
	Short: "create a new configuration file",
	Long: `create a new configuration file.

The path to the repository database need not exist, but it must be absolute.

The configuration file will be created at $XDG_CONFIG_HOME/repoctl/config.toml.
If neither $XDG_CONFIG_HOME nor $HOME are defined, then you need to tell us
where you want the configuration file to be placed. Note that it won't be
found automatically. You will have to set $REPOCTL_CONFIG.
`,
	Run: newConfig,
}

func newConfig(cmd *cobra.Command, args []string) {
	if nConf == "" {
		dieOnError(ErrInvalidConfPath)
	}
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(1)
	}

	repo := args[0]
	if !path.IsAbs(repo) {
		dieOnError(ErrRepoNotAbs)
	}

	dir := path.Dir(nConf)
	if ex, _ := osutil.DirExists(dir); !ex {
		if Conf.Debug {
			fmt.Println("Creating directory structure", dir, "...")
		}
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s.\n", err)
		}
	}

	fmt.Println("writing new configuration file at", nConf, "...")
	Conf.Repository = repo
	Conf.Unconfigured = false
	err := Conf.WriteFile(nConf)
	if err != nil {
		dieOnError(err)
	}
}
