// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/cassava/repoctl"
	"github.com/cassava/repoctl/conf"
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
	Long: `Create a new repository or configuration file.

  For repoctl to work, there must be at least one repository and one
  configuration file. These can be created manually, or repoctl can
  create them for you.

  If you already have a repository, creating a new config is sufficient:

    repoctl new config /path/to/repository

  If you do not have a repository, you can create both repository and
  configuration file in one step:

    repoctl new repo /path/to/repository

  See the respective commands for more information.
`,
}

var (
	nConf string // path to configuration
)

func init() {
	NewCmd.PersistentFlags().StringVarP(&nConf, "config", "c", conf.HomeConf(), "path to configuration file")
}

func init() {
	NewCmd.AddCommand(newRepoCmd)
	NewCmd.AddCommand(newConfigCmd)
}

var newRepoCmd = &cobra.Command{
	Use:   "repo </path/to/repo/database>",
	Short: "create a new repository and configuration file",
	Long: `Create a new repository with configuration file.

  FIXME: This function still needs to be implemented.
`,
	Run: func(cmd *cobra.Command, args []string) {
		panic("not implemented")
	},
}

var newConfigCmd = &cobra.Command{
	Use:   "config </path/to/repo/database>",
	Short: "create a new configuration file",
	Long: `Create a new initial configuration file.

  The minimal configuration of repoctl is read from a configuration
  file, which tells repoctl where your repositories are. The absolute
  path to the repository database must be given as the only argument.
  If the suffix "db.tar.gz" is omitted, it is appended automatically.

  There are several places that repoctl reads its configuration from.
  If $REPOCTL_CONFIG is set, then only this path is loaded. Otherwise,
  the following paths are checked for repoctl/config.toml:

    1. All the paths in $XDG_CONFIG_DIRS, where a colon ":" acts as
       the separator. If $XDG_CONFIG_DIRS is not set or empty, then
       it defaults to /etc/xdg.
    2. The path given by $XDG_CONFIG_HOME. If $XDG_CONFIG_HOME is not
       set, it defaults to $HOME/.config.

  In most systems then, repoctl will read:

    /etc/xdg/repoctl/config.toml
    /home/you/.config/repoctl/config.toml

  The default location to create a repoctl configuration file is in
  your $XDG_CONFIG_HOME directory.

  When creating a configuration file, repoctl will overwrite any
  existing files. You have been warned.
`,
	Example: `  repoctl new config /srv/abs/atlas.db.tar.gz
  repoctl new config /srv/abs/atlas
  repoctl new config -c /etc/xdg/repoctl/config.toml /srv/abs/atlas
  REPOCTL_CONFIG=/etc/repoctl.conf repoctl new config /srv/abs/atlas`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}

		err := NewConfig(nConf, args[0])
		dieOnError(err)
	},
}

func NewConfig(confpath, repo string) error {
	if confpath == "" {
		return ErrInvalidConfPath
	}

	if !path.IsAbs(repo) {
		return repoctl.ErrRepoDirRelative
	}
	if !strings.HasSuffix(repo, ".db.tar.gz") {
		repo += ".db.tar.gz"
	}

	dir := path.Dir(confpath)
	if ex, _ := osutil.DirExists(dir); !ex {
		if Conf.Debug {
			fmt.Println("Creating directory structure", dir, "...")
		}
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s.\n", err)
		}
	}

	fmt.Println("writing new configuration file at", confpath, "...")
	Conf.Repository = repo
	Conf.Unconfigured = false
	return Conf.WriteFile(confpath)
}
