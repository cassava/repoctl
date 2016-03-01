// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"

	"github.com/goulash/osutil"
	"github.com/spf13/cobra"
)

var hostListen string

func init() {
	MainCmd.AddCommand(hostCmd)

	hostCmd.Flags().StringVar(&hostListen, "listen", ":8080", "which address and port to listen on")
}

var hostCmd = &cobra.Command{
	Use:     "host",
	Aliases: []string{"serve"},
	Short:   "host repository on a network",
	Long: `Host the repository on a network.

  This is essentially static file serving the repository on a specific
  address and port, and is only meant for temporary use. If you want
  to run something like this for longer, consider using darkhttpd.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if ok, _ := osutil.DirExists(Repo.Directory); !ok {
			return fmt.Errorf("repo directory %q does not exist", Repo.Directory)
		}
		fmt.Printf("Serving %s on %s...\n", Repo.Directory, hostListen)
		return http.ListenAndServe(hostListen, http.FileServer(http.Dir(Repo.Directory)))
	},
}
