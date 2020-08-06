// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cassava/repoctl/conf"
	"github.com/goulash/osutil"
	"github.com/spf13/cobra"
)

func init() {
	MainCmd.AddCommand(newCmd)
}

var (
	ErrInvalidConfPath = errors.New("invalid configuration path")
)

// TODO: I'm not sure I like the interface of this command. Do I really
// need subcommands? Would it not make more sense to have configuration
// creation happen with repository creation?
//
// Also, profile creation?
var newCmd = &cobra.Command{
	Use:   "new [command] [flags]",
	Short: "Create a new repository or configuration file",
	Long: `Create a new repository or configuration file.

  For repoctl to work, there must be at least one repository and one
  configuration file. These can be created manually, or repoctl can
  create them for you.

  If you already have a repository, creating a new config is sufficient:

    repoctl new config /path/to/repository/database.db.tar.gz

  If you do not have a repository, you can create both repository and
  configuration file in one step:

    repoctl new repo /path/to/repository

  See the respective commands for more information.
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Prevent errors that we print being printed a second time by cobra.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		return nil
	},
}

var (
	nConf string // path to configuration
)

func init() {
	newCmd.PersistentFlags().StringVarP(&nConf, "config", "c", conf.HomeConf(), "path to configuration file")
}

func init() {
	newCmd.AddCommand(newRepoCmd)
	newCmd.AddCommand(newConfigCmd)
}

var newRepoCmd = &cobra.Command{
	Use:   "repo </path/to/repo/database>",
	Short: "Create a new repository and configuration file",
	Long: `Create a new repository with configuration file.

  FIXME: This function still needs to be implemented.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("not implemented")
	},
}

var newConfigCmd = &cobra.Command{
	Use:   "config </path/to/repo/database>",
	Short: "Create a new configuration file",
	Long: `Create a new initial configuration file.

  The minimal configuration of repoctl is read from a configuration
  file, which tells repoctl where your repositories are. The absolute
  path to the repository database must be given as the only argument.
  The database file specified must have an extension of one of:

    - ".db.tar"             - ".db.tar.gz"
    - ".db.tar.xz"          - ".db.tar.bz2"
	- ".db.tar.zst"

  The recommended database extension to use is ".db.tar.gz".

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

  When creating a configuration file, repoctl will create a backup of
  any existing files.
`,
	Example: `  repoctl new config /srv/abs/atlas.db.tar.gz
  repoctl new config -c /etc/xdg/repoctl/config.toml /srv/abs/atlas.db.tar.gz
  REPOCTL_CONFIG=/etc/repoctl.conf repoctl new config /srv/abs/atlas.db.tar.gz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return &UsageError{"new config", "new config command takes only one argument", cmd.Usage}
		}

		return newConfig(nConf, args[0])
	},
}

func newConfig(confpath, repo string) error {
	if confpath == "" {
		return ErrInvalidConfPath
	}

	if !filepath.IsAbs(repo) {
		var err error
		repo, err = filepath.Abs(repo)
		if err != nil {
			return err
		}
	}

	var extOk bool
	for _, ext := range []string{".db.tar", ".db.tar.gz", ".db.tar.xz", ".db.tar.bz2"} {
		if strings.HasSuffix(repo, ext) {
			extOk = true
			break
		}
	}
	if !extOk {
		fmt.Fprintf(os.Stderr, "Warning: specified repository database %q has an unexpected extension.\n", repo)
		fmt.Fprintf(os.Stderr, "Should be one of \".db.tar\", \".db.tar.gz\", \".db.tar.xz\", or \"db.tar.bz2\".\n")
	}

	dir := filepath.Dir(confpath)
	if ex, _ := osutil.DirExists(dir); !ex {
		if Conf.Debug {
			fmt.Println("Creating directory structure", dir, "...")
		}
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s.\n", err)
		}
	}

	if ex, _ := osutil.FileExists(confpath); ex {
		backup := generateBackupPath(confpath, ".bak")
		fmt.Fprintf(os.Stderr, "Backing up current configuration to: %s\n", backup)
		os.Rename(confpath, backup)
	}

	fmt.Println("Writing new configuration file at", confpath, "...")
	Conf.Repository = repo
	Conf.Unconfigured = false
	return Conf.WriteFile(confpath)
}

func generateBackupPath(filepath string, suffix string) string {
	backupPath := filepath + suffix
	ex, _ := osutil.FileExists(backupPath)

	if ex {
		for i := 1; ex; i++ {
			backupPath = filepath + suffix + "." + strconv.Itoa(i)
			ex, _ = osutil.FileExists(backupPath)
		}
	}

	return backupPath
}
