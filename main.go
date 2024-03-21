// Copyright (c) 2020, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cassava/repoctl/conf"
	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman"
	"github.com/cassava/repoctl/repo"
	"github.com/spf13/cobra"
)

var (
	// Conf loads and stores the configuration (apart from command line
	// configuration) of this program, including where the repository is.
	Conf *conf.Configuration

	// Profile is what is used to configure the Repo. For many commands
	// it is optional.
	Profile *conf.Profile

	// Repo lets us use the repoctl library to do the most of the work.
	Repo *repo.Repo
)

func init() {
	// Arguments from the command line override the configuration file,
	// so we have to add the flags after loading the configuration.
	c, err := conf.FindAll()
	if err != nil {
		// We didn't manage to load any configuration, which means that repoctl
		// is unconfigured. There are some commands that work nonetheless, so
		// we can't stop now -- which is why we don't os.Exit(1).
		term.Errorf("Error: %s\n", err)
	}
	Conf = c

	// Set default terminal output options.
	term.SetMode(Conf.Color)
	configureTerm()

	MainCmd.PersistentFlags().StringVarP(&Conf.CurrentProfile, "profile", "P", c.DefaultProfile, "configuration profile to use")
	MainCmd.RegisterFlagCompletionFunc("profile", completeProfiles)

	MainCmd.PersistentFlags().BoolVarP(&Conf.Columnate, "columns", "s", c.Columnate, "show items in columns rather than lines")
	MainCmd.PersistentFlags().BoolVarP(&Conf.Quiet, "quiet", "q", c.Quiet, "show minimal amount of information")
	MainCmd.PersistentFlags().BoolVar(&Conf.Debug, "debug", c.Debug, "show unnecessary debugging information")
	MainCmd.PersistentFlags().Var(term.Formatter, "color", "when to use color (auto|never|always)")
}

var MainCmd = &cobra.Command{
	Use:   "repoctl",
	Short: "Manage local Pacman repositories",
	Long: `Repoctl helps manage local Pacman repositories, and acts in particular as
a supplement to the repo-add and repo-remove tools that come with Pacman.

It also comes with several commands for searching, querying, and downloading
packages from AUR.

Note that in all of these commands, the following terminology is used:

  PKGNAME: is the name of the package, e.g. pacman
  PKGFILE: is the path to a package file, e.g. pacman-3.5.3-i686.pkg.tar.xz
  DBPATH:  is the (absolute) path to your repo database, e.g. /srv/sirius.db.tar.zst

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

---

If you are new to repoctl, perform the following commands to get started:

  repoctl conf new DBPATH  # Create a new configuration file
  repoctl conf edit        # Edit the configuration file
  repoctl reset            # Initialize the repository

You can then continue by building and adding packages:

  repoctl search QUERY
  repoctl down -r PKGNAME
  # cd PKGNAME && makepkg
  repoctl add PKGFILE ...

And by viewing the status of your repository:

  repoctl status

`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Prevent errors that we print being printed a second time by cobra.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		configureTerm()

		return nil
	},
}

func configureTerm() {
	if Conf.Quiet {
		term.StdOut = nil
	} else {
		term.StdOut = os.Stdout
	}
	if Conf.Debug {
		term.StdOut = os.Stderr
		term.DebugOut = os.Stderr
		pacman.DebugWriter = term.NewDebugWriter(term.DebugOut)
	} else {
		term.DebugOut = nil
		pacman.DebugWriter = nil
	}
}

func exceptQuiet() {
	if Conf.Quiet {
		term.Debugf("Overriding quiet: doesn't make sense for this command.\n")
	}
	term.StdOut = os.Stdout
}

// main loads the configuration and executes the primary command.
func main() {
	err := MainCmd.Execute()
	if err != nil {
		// If this is an ExecError, we deal with it specially:
		if e, ok := err.(*ExecError); ok {
			term.Errorf("Error: command %q failed: %s.\n", e.Command, e.Err)
			term.Errorff("Command output:\n%s", e.Output)
			os.Exit(1)
		}

		// All other errors:
		term.Errorf("Error: %s.\n", err)
		os.Exit(1)
	}
}

// ProfileInit should be used as the PreRunE part of every command
// that needs to make use of the profile or the Repo.
//
// Make sure to use ProfileTeardown in the PostRunE if using this.
func ProfileInit(cmd *cobra.Command, args []string) error {
	// Try to load the profile.
	if len(Conf.Profiles) == 0 {
		return fmt.Errorf("please create a configuration profile to proceed")
	}

	p, name, err := Conf.SelectProfile()
	if err != nil {
		return fmt.Errorf("cannot select unknown profile %q", name)
	} else if p == nil {
		return fmt.Errorf("cannot load default profile")
	}

	// 1. Initialize selected profile
	err = p.Init()
	if err != nil {
		// This currently only happens if the repository is unset or relative.
		return fmt.Errorf("cannot load profile %q: %s", name, err)
	}

	// 2. Set the global profile variable.
	Profile = p

	// 3. Create a new Repo struct from the configuration.
	Repo, err = repo.NewFromConf(Conf)
	if err != nil {
		return fmt.Errorf("cannot load profile %q: %s", name, err)
	}

	// 4. Run pre-action if defined.
	if Profile.PreAction != "" {
		return runShellCommand(Profile.PreAction)
	}

	return nil
}

// ProfileTeardown should be used as the PostRunE part of every command
// that needs to make use of the profile or the Repo.
func ProfileTeardown(cmd *cobra.Command, args []string) error {
	if Profile != nil && Profile.PostAction != "" {
		return runShellCommand(Profile.PostAction)
	}
	return nil
}

// runShellCommand runs the cmd in a shell and returns whether an error occurred.
// If an error is returned, it is of type *ExecError, which contains the field
// `Output` that contains the commands stdout and stderr output.
func runShellCommand(cmd string) error {
	bs, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return &ExecError{
			Err:     err,
			Output:  string(bs),
			Command: cmd,
		}
	}
	return nil
}

type ExecError struct {
	Err     error
	Output  string
	Command string
}

func (err *ExecError) Error() string { return err.Err.Error() }
