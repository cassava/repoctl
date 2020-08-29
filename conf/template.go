// Copyright (c) 2020, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package conf

import (
	"bytes"
	"fmt"
	"text/template"
)

var PropertiesTmpl = template.Must(template.New("properties").Funcs(template.FuncMap{
	"printt": printt,
}).Parse(`Current configuration:
    columnate = {{ printt .Columnate }}
    color = {{ printt .Color }}
    quiet = {{ printt .Quiet }}

    current_profile = {{ printt .CurrentProfile }}
    default_profile = {{ printt .DefaultProfile }}
    {{ range $key, $value := .Profiles }}
    [profiles.{{  $key }}]
        repo = {{ printt $value.Repository }}
        add_params = {{ printt $value.AddParameters }}
        rm_params = {{ printt $value.RemoveParameters }}
        ignore_aur = {{ printt $value.IgnoreAUR }}
        require_signature = {{ printt $value.RequireSignature }}
        backup = {{ printt $value.Backup }}
        backup_dir = {{ printt $value.BackupDir }}
        interactive = {{ printt $value.Interactive }}
        pre_action = {{printt $value.PreAction}}
        post_action = {{ printt $value.PostAction }}
    {{ end }}
`))

var ConfigurationTmpl = template.Must(template.New("config").Funcs(template.FuncMap{
	"printt": printt,
}).Parse(`# repoctl configuration

# columnate specifies that listings should be in columns rather than
# in lines. This only applies to the list command.
columnate = {{ printt .Columnate }}

# color specifies when to use color. Can be one of auto, always, and never.
color = {{ printt .Color }}

# quiet specifies whether repoctl should print more information or less.
# I prefer to know what happens, but if you don't like it, you can change it.
quiet = {{ printt .Quiet }}

# default_profile specifies which profile should be used when none is
# specified on the command line.
default_profile = {{ printt .DefaultProfile }}

{{ range $key, $value := .Profiles }}[profiles.{{  $key }}]
  # repo is the full path to the repository that will be managed by repoctl.
  # The packages that belong to the repository are assumed to lie in the
  # same folder.
  repo = {{ printt $value.Repository }}

  # add_params is the set of parameters that will be passed to repo-add
  # when it is called. Specify one time for each parameter.
  add_params = {{ printt $value.AddParameters }}

  # rm_params is the set of parameters that will be passed to repo-remove
  # when it is called. Specify one time for each parameter.
  rm_params = {{ printt $value.RemoveParameters }}

  # ignore_aur is a set of package names that are ignored in conjunction
  # with AUR related tasks, such as determining if there is an update or not.
  ignore_aur = {{ printt $value.IgnoreAUR }}

  # require_signature prevents packages from being added that do not
  # also have a signature file.
  require_signature = {{ printt $value.RequireSignature }}

  # backup specifies whether package files should be backed up or deleted.
  # If it is set to false, then obsolete package files are deleted.
  backup = {{ printt $value.Backup }}

  # backup_dir specifies which directory backups are stored in.
  # - If a relative path is given, then it is interpreted as relative to
  #   the repository directory.
  # - If the path here resolves to the same as repo, then obsolete packages
  #   are effectively ignored by repoctl, if backup is true.
  backup_dir = {{ printt $value.BackupDir }}

  # interactive specifies that repoctl should ask before doing anything
  # destructive.
  interactive = {{ printt $value.Interactive }}

  # pre_action is a command that should be executed before doing anything
  # with the repository, like reading or modifying it. Useful for mounting
  # a remote filesystem.
  pre_action = {{printt $value.PreAction}}

  # post_action is a command that should be executed before exiting.
  post_action = {{ printt $value.PostAction }}
{{ end }}
`))

// printt returns a TOML representation of the value.
//
// This function is used in printing TOML values in the template.
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
