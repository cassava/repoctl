// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/spf13/cobra"

var ResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "recreate repository database",
	Long: `Delete the repository database and re-add all packages in repository.
    
  Essentially, this command deletes the repository database and
  recreates it by running the update command.
`,
	Run: reset,
}

func reset(cmd *cobra.Command, args []string) {
	panic("not implemented")
}
