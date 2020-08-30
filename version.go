// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"text/template"

	"github.com/cassava/repoctl/internal/term"
	"github.com/spf13/cobra"
)

func init() {
	MainCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:                   "version",
	Short:                 "Show version information and current configuration",
	Long:                  "Show version information of repoctl, as well as the current configuration.",
	ValidArgsFunction:     completeNoFiles,
	Args:                  cobra.ExactArgs(0),
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		exceptQuiet()

		var progInfo = struct {
			Name      string
			Author    string
			Email     string
			Version   string
			Date      string
			Homepage  string
			Copyright string
			License   string
		}{
			Name:      "repoctl",
			Author:    "Ben Morgan",
			Email:     "cassava@iexu.de",
			Version:   "0.21",
			Date:      "30 August, 2020",
			Copyright: "2016-2020",
			Homepage:  "https://github.com/cassava/repoctl",
			License:   "MIT",
		}
		versionTmpl.Execute(term.StdOut, progInfo)

		// Print the current configuration.
		term.Println()
		Conf.WriteProperties(term.StdOut)
	},
}

var versionTmpl = template.Must(template.New("version").Parse(
	`{{.Name}} version {{.Version}} ({{.Date}})
Copyright {{.Copyright}}, {{.Author}} <{{.Email}}>

You may find {{.Name}} on the Internet at
    {{.Homepage}}
Please report any bugs you may encounter.

The source code of {{.Name}} is licensed under the {{.License}} license.
`))
