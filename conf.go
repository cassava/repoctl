// Copyright (c) 2020, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cassava/repoctl/conf"
	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman/alpm"
	"github.com/goulash/osutil"
	"github.com/spf13/cobra"
)

var (
	confVar          *conf.Configuration
	confPath         string
	confShowTemplate bool
	confEditorPath   string
)

func init() {
	MainCmd.AddCommand(confCmd)
	confCmd.PersistentFlags().StringVarP(&confPath, "config", "c", conf.HomeConf(), "path to configuration file")

	confCmd.AddCommand(confShowCmd)
	confShowCmd.Flags().BoolVar(&confShowTemplate, "template", false, "show configuration as a TOML template")

	confCmd.AddCommand(confEditCmd)
	confEditCmd.Flags().StringVarP(&confEditorPath, "editor", "e", os.Getenv("EDITOR"), "editor path")

	confCmd.AddCommand(confMigrateCmd)
	confCmd.AddCommand(confNewCmd)
}

var (
	// ErrInvalidConfPath is returned when an invalid path has been specified as
	// the configuration file.
	ErrInvalidConfPath = errors.New("invalid configuration path")
)

var confCmd = &cobra.Command{
	Use:   "conf {show | new | edit | migrate } [options]",
	Short: "Create, edit, or show the repoctl configuration",
	Long: `Create, edit, or show the repoctl configuration.

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

  For repoctl to work, there must be at least one repository and one
  configuration file. These can be created manually, or repoctl can
  create them for you.

  If you already have a repository, creating a new config is sufficient:

    repoctl conf new /path/to/repository/database.db.tar.gz

  Otherwise, make sure to run

    repoctl reset

  afterwards.

  Repoctl supports multiple repository configuration through profiles.
  You can add a profile by editing the configuration file manually:

    repoctl conf edit
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Prevent errors that we print being printed a second time by cobra.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		// This may end up reading the same data as Conf, but that's OK.
		if ex, _ := osutil.Exists(confPath); ex {
			var err error
			confVar, err = conf.Read(confPath)
			return err
		}

		// We are using the default configuration then, which is OK for
		// certain operations, such as show, but not for others.
		confVar = conf.Default()
		return nil
	},
}

var confShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Show current repoctl configuration.

  If the configuration path is set to the empty string "", then this command
  shows the default repoctl configuration. The default configuration is
  sufficient for several commands, such as down or version.
`,
	Args:              cobra.ExactArgs(0),
	ValidArgsFunction: completeNoFiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		exceptQuiet()
		if confShowTemplate {
			confVar.WriteTemplate(os.Stdout)
		} else {
			confVar.WriteProperties(os.Stdout)
		}
		return nil
	},
}

var confMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate current configuration format to the latest one",
	Long: `Migrate current configuration format to the latest format.

  The configuration format has changed in version 0.21 to support multiple
  repository profiles. This is a breaking change from previous versions, but
  repoctl is able to understand old format versions if you have no need for
  multiple profiles. If you want to take advantage of profiles, you can migrate
  your old configuration file. Your previous configuration will be backed up.

  Warning: This command will unconditionally rewrite the configuration file,
  erasing any comments you have made in the file. It will do so even if the
  configuration file does not need to be migrated!
`,
	Args:              cobra.ExactArgs(0),
	ValidArgsFunction: completeNoFiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ex, _ := osutil.Exists(confPath); !ex {
			return fmt.Errorf("cannot migrate %s: file does not exist", confPath)
		}

		if ex, _ := osutil.FileExists(confPath); ex {
			backup := generateBackupPath(confPath, ".bak")
			term.Printf("Backing up current configuration to: %s\n", backup)
			os.Rename(confPath, backup)
		}

		term.Printf("Writing new configuration file to: %s\n", confPath)
		return confVar.WriteFile(confPath)
	},
}

var confEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit current configuration",
	Long: `Edit repoctl configuration, creating it if necessary.

  This command will launch your preferred editor with the file given as its
  only argument. It will use the environment variable EDITOR as its default,
  unless you specify an editor explicitely with the --editor flag.

  If the configuration file does not exist, repoctl will create a template
  for you.
`,
	Args:              cobra.ExactArgs(0),
	ValidArgsFunction: completeNoFiles,
	RunE: func(cmd *cobra.Command, args []string) error {
		if confPath == "" {
			return ErrInvalidConfPath
		}
		if confEditorPath == "" {
			return fmt.Errorf("editor must be specified in $EDITOR or with -e flag")
		}

		// Create configuration if missing
		if ex, _ := osutil.Exists(confPath); !ex {
			term.Printf("Writing new configuration file to: %s\n", confPath)
			err := newConfig(confPath, "")
			if err != nil {
				return fmt.Errorf("cannot create default config: %w", err)
			}
		}

		// Launch editor
		term.Debugf("Executing: %s %s\n", confEditorPath, confPath)
		sys := exec.Command(confEditorPath, confPath)
		sys.Stdin = os.Stdin
		sys.Stdout = os.Stdout
		sys.Stderr = os.Stderr
		return sys.Run()
	},
}

var confNewCmd = &cobra.Command{
	Use:   "new DBPATH",
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

  The default location to create a repoctl configuration file is in
  your $XDG_CONFIG_HOME directory.

  When creating a configuration file, repoctl will create a backup of
  any existing files.
`,
	Example: `  repoctl conf new /srv/abs/atlas.db.tar.gz
  repoctl conf new -c /etc/xdg/repoctl/config.toml /srv/abs/atlas.db.tar.gz
  REPOCTL_CONFIG=/etc/repoctl.conf repoctl conf new /srv/abs/atlas.db.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if confPath == "" {
			return ErrInvalidConfPath
		}

		// Check that the repo specified is according to the specifications.
		r := args[0]
		if !filepath.IsAbs(r) {
			var err error
			r, err = filepath.Abs(r)
			if err != nil {
				return err
			}
		}

		if !alpm.HasDatabaseFormat(r) {
			fmt.Fprintf(os.Stderr, "Warning: Specified repository database %q has an unexpected extension.\n", r)
			fmt.Fprintf(os.Stderr, "         It should conform to this pattern: .db.tar.(zst|xz|gz|bz2).\n")
			base := filepath.Base(r)
			if i := strings.IndexRune(base, '.'); i != -1 {
				base = base[:i]
			}
			fmt.Fprintf(os.Stderr, "         For example: %s.db.tar.zst\n", filepath.Join(filepath.Dir(r), base))
			fmt.Fprintf(os.Stderr, "Warning: Continuing anyway.\n")
		}

		// Create a new configuration.
		return newConfig(confPath, r)
	},
}

func newConfig(file, repo string) error {
	dir := filepath.Dir(file)
	if ex, _ := osutil.DirExists(dir); !ex {
		term.Debugf("Creating directory structure", dir, "...\n")
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s.\n", err)
		}
	}

	if ex, _ := osutil.FileExists(file); ex {
		backup := generateBackupPath(file, ".bak")
		term.Printf("Backing up current configuration to: %s\n", backup)
		os.Rename(file, backup)
	}

	term.Printf("Writing new configuration file at: %s\n", file)
	if repo == "" {
		return conf.Default().WriteFile(file)
	}
	return conf.New(repo).WriteFile(file)
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
