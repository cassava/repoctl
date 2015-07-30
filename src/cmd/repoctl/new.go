// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

// TODO: I'm not sure I like the interface of this command. Do I really
// need subcommands? Would it not make more sense to have configuration
// creation happen with repository creation?
//
// Also, profile creation?
var NewCmd = &cobra.Command{
	Use:   "new [command] [flags]",
	Short: "create a new repository or configuration file",
	Long: `Create either a new repository or configuration file.
If any flags are missing, you are prompted for their values interactively.`,
}

var newRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "create a new repository",
}

var newConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "create a new configuration file",
}

var newProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "create a new profile",
}
