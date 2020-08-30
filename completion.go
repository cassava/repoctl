// Copyright (c) 2020, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cassava/repoctl/pacman"
	"github.com/cassava/repoctl/pacman/alpm"
	"github.com/cassava/repoctl/pacman/aur"
	"github.com/cassava/repoctl/pacman/pkgutil"
	"github.com/cassava/repoctl/repo"
	"github.com/spf13/cobra"
)

func init() {
	MainCmd.AddCommand(completionCmd)
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion",
	Long: `Generate shell completion scripts for repoctl.

  If you installed repoctl via your package manager (probably Pacman), then
  shell completions should already be installed. If this is not the case, and
  you only have the repoctl binary, then this command can help you add
  completion to your shell.

  Completions are supported for three shells: Bash, Zsh, and Fish.
  If you don't provide the shell for which the completion script should be
  generated, it will generated one based on the SHELL environment variable.

  Bash:
    To load the completion for the current session:
        source <(repoctl completion bash)
    To install the completion for all sessions, execute once:
        repoctl completion bash > /etc/bash_completion.d/repoctl

  Zsh:
    To install completions for all sessions, execute once:
        repoctl completion zsh > "${fpath[1]}/_yourprogram"

    If shell completion is not already enabled in your environment you will need
    to enable it.  You can execute the following once:
        echo "autoload -U compinit; compinit" >> ~/.zshrc
    You will need to start a new shell for this setup to take effect.

  Fish:
    To load the completion for the current session:
        repoctl completion fish | source
    To install completions for all your sessions, execute once:
        repoctl completion fish > ~/.config/fish/completions/repoctl.fish
`,
	DisableFlagsInUseLine: true,
	Hidden:                true,
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var shell string
		if len(args) == 0 {
			val, ok := os.LookupEnv("SHELL")
			if !ok {
				return fmt.Errorf("cannot derive shell from SHELL environment variable")
			}
			shell = filepath.Base(val)
		} else if len(args) == 1 {
			shell = args[0]
		} else {
			return fmt.Errorf("the completion command expects a single argument")
		}

		switch shell {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		default:
			return fmt.Errorf("unknown shell %q", shell)
		}

		return nil
	},
}

func completeDirectory(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveFilterDirs
}

func completeProfiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	result := make([]string, 0, len(Conf.Profiles))
	for k := range Conf.Profiles {
		result = append(result, k)
	}
	return result, cobra.ShellCompDirectiveNoFileComp
}

func completeLocalPackageFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return alpm.PackageExtensions, cobra.ShellCompDirectiveFilterFileExt
}

func completeRepoPackageNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// FIXME: Take profiles into account.

	r, err := repo.NewFromConf(Conf)
	if err != nil || r == nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Get the names of the packages in the repository
	var names []string
	names, err = pacman.ReadDirApproxOnlyNames(nil, r.Directory)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return filterCompletionResults(names, args, toComplete)
}

func completeAURPackageNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// We don't complete when the argument is too small, because otherwise the
	// AUR will probably be overloaded.
	if len(toComplete) < 4 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	pkgs, err := aur.SearchByName(toComplete)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return filterCompletionResults(pkgutil.Map(pkgs, pkgutil.PkgName), args, toComplete)
}

func filterCompletionResults(results []string, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Filter out all names that we already have or that don't match.
	alreadyHas := make(map[string]bool)
	for _, x := range args {
		alreadyHas[x] = true
	}

	filteredResults := make([]string, 0, len(results))
	for _, x := range results {
		if strings.HasPrefix(x, toComplete) && !alreadyHas[x] {
			filteredResults = append(filteredResults, x)
		}
	}

	return filteredResults, cobra.ShellCompDirectiveNoFileComp
}
