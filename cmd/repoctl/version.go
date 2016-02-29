// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

func init() {
	RepoctlCmd.AddCommand(versionCmd)
}

type programInfo struct {
	Name      string
	Author    string
	Email     string
	Version   string
	Date      string
	Homepage  string
	Copyright string
	License   string
}

const versionTmpl = `{{.Name}} version {{.Version}} ({{.Date}})
Copyright {{.Copyright}}, {{.Author}} <{{.Email}}>

You may find {{.Name}} on the Internet at
    {{.Homepage}}
Please report any bugs you may encounter.

The source code of {{.Name}} is licensed under the {{.License}} license.
`

var progInfo = programInfo{
	Name:      "repoctl",
	Author:    "Ben Morgan",
	Email:     "neembi@gmail.com",
	Version:   "0.14",
	Date:      "6 October 2015",
	Copyright: "2015",
	Homepage:  "https://github.com/cassava/repoctl",
	License:   "MIT",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version and date information",
	Long:  "Show the official version number of repoctl, as well as the release date.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Don't try to load repoctl configuration
	},
	Run: func(cmd *cobra.Command, args []string) {
		template.Must(template.New("version").Parse(versionTmpl)).Execute(os.Stdout, progInfo)
	},
}
