// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"github.com/cassava/repoctl/conf"
	"github.com/spf13/cobra"
)

func init() {
	MainCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:                   "version",
	Short:                 "Show version information and current configuration",
	Long:                  "Show version information of repoctl, as well as the current configuration.",
	Args:                  cobra.ExactArgs(0),
	DisableFlagsInUseLine: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Prevent errors that we print being printed a second time by cobra.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var progInfo = struct {
			Name      string
			Author    string
			Email     string
			Version   string
			Date      string
			Homepage  string
			Copyright string
			License   string
			Conf      *conf.Configuration
		}{
			Name:      "repoctl",
			Author:    "Ben Morgan",
			Email:     "neembi@gmail.com",
			Version:   "0.20",
			Date:      "24 July, 2020",
			Copyright: "2016-2020",
			Homepage:  "https://github.com/cassava/repoctl",
			License:   "MIT",
			Conf:      Conf,
		}
		versionTmpl.Execute(os.Stdout, progInfo)
	},
}

var versionTmpl = template.Must(template.New("version").Funcs(template.FuncMap{
	"printt": printt,
}).Parse(`{{.Name}} version {{.Version}} ({{.Date}})
Copyright {{.Copyright}}, {{.Author}} <{{.Email}}>

You may find {{.Name}} on the Internet at
    {{.Homepage}}
Please report any bugs you may encounter.

The source code of {{.Name}} is licensed under the {{.License}} license.

{{if .Conf.Unconfigured}}Default{{else}}Current{{end}} configuration:
    repo                  = {{printt .Conf.Repository}}
    add_params            = {{printt .Conf.AddParameters}}
    rm_params             = {{printt .Conf.RemoveParameters}}
    ignore_aur            = {{printt .Conf.IgnoreAUR}}
    require_signature     = {{printt .Conf.RequireSignature}}
    backup                = {{printt .Conf.Backup}}
    backup_dir            = {{printt .Conf.BackupDir}}
    interactive           = {{printt .Conf.Interactive}}
    columnate             = {{printt .Conf.Columnate}}
    color                 = {{printt .Conf.Color}}
    quiet                 = {{printt .Conf.Quiet}}
    debug                 = {{printt .Conf.Debug}}
    pre_action            = {{printt .Conf.PreAction}}
    post_action           = {{printt .Conf.PostAction}}
    action_on_completion  = {{printt .Conf.ActionOnCompletion}}
`))

// printt returns a TOML representation of the value.
//
// This function is used in printing TOML values in the template.
//
// NOTE: Copied from ../../conf/config.go
func printt(v interface{}) string {
	switch obj := v.(type) {
	case string:
		return fmt.Sprintf("%q", obj)
	case []string:
		if len(obj) == 0 {
			return "[]"
		}

		var buf bytes.Buffer
		buf.WriteRune('[')
		for _, k := range obj[:len(obj)-1] {
			buf.WriteString(fmt.Sprintf("%q", k))
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q", obj[len(obj)-1]))
		buf.WriteRune(']')
		return buf.String()
	default: // floats, ints, bools
		return fmt.Sprintf("%v", obj)
	}
}
