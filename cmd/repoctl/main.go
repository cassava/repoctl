// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

type UsageError struct {
	Cmd   string
	Msg   string
	Usage func() error
}

func (e *UsageError) Error() string {
	return fmt.Sprintf("%s", e.Msg)
}

type ExecError struct {
	Err     error
	Output  string
	Command string
}

func (err *ExecError) Error() string { return err.Err.Error() }

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

var MainCmd = &cobra.Command{
	Use:   "repoctl",
	Short: "manage local Pacman repositories",
	Long: `Repoctl helps manage local Pacman repositories, and acts in particular as
a supplement to the repo-add and repo-remove tools that come with Pacman.

If compiling and installing from AUR every time is not what you want,
or if you want to host your own repository, then repoctl might be for you.

Note that in all of these commands, the following terminology is used:

    pkgname: is the name of the package, e.g. pacman
    pkgfile: is the path to a package file, e.g. pacman-3.5.3-i686.pkg.tar.xz
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Prevent errors that we print being printed a second time by cobra.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		if strings.HasPrefix(cmd.Use, "help") {
			return nil
		}

		// Make sure that repoctl is configured and that the configuration
		// makes sense. This function can be overriden if it's not necessary
		// for a command.
		if Conf.Unconfigured {
			return errors.New("repoctl is unconfigured, please create configuration")
		} else if Conf.Repository == "" {
			return conf.ErrRepoUnset
		}
		Repo = Conf.Repo()

		// Run preaction if defined.
		if Conf.PreAction != "" {
			return runShellCommand(Conf.PreAction)
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		// If PersistentPreRunE was overridden, then don't execute this step.
		// We can determine this by looking to see if Repo was set.
		if Conf.PostAction != "" && Repo != nil {
			return runShellCommand(Conf.PostAction)
		}
		return nil
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
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}

	// Arguments from the command line override the configuration file,
	// so we have to add the flags after loading the configuration.
	//
	// TODO: Maybe in the future we will make it possible to specify the
	// configuration file via the command line; right now it is not a priority.
	addConfFlags(MainCmd)

	err = MainCmd.Execute()
	if err != nil {
		// If this is an ExecError, we deal with it specially:
		if e, ok := err.(*ExecError); ok {
			fmt.Fprintf(os.Stderr, "error: command %q failed: %s\n", e.Command, e.Err)
			fmt.Fprintf(os.Stderr, "command output:\n%s", e.Output)
			os.Exit(1)
		}

		// All other errors:
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		if e, ok := err.(*UsageError); ok {
			e.Usage()
		}
		os.Exit(1)
	}
}
