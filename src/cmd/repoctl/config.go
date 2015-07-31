// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/goulash/osutil"
)

var (
	ErrNoConfig   = errors.New("no configuration files found in path")
	ErrUnsetHOME  = errors.New("HOME environment variable unset")
	ErrRepoNotAbs = errors.New("repository path must be absolute")
	ErrRepoUnset  = errors.New("repository path must be set in configuration")
)

const configurationPostfix = "repoctl/config.toml"

var configurationTmpl = template.Must(template.New("config").Parse(`# repoctl configuration
{{ if .Unconfigured }}
# When repoctl is unconfigured, nothing makes sense.
# Remove this line when you are done, or set it to false.
unconfigured = {{printt .Unconfigured}}
{{ end }}

# repo is the full path to the repository that will be managed by repoctl.
# The packages that belong to the repository are assumed to lie in the
# same folder.
#repo = {{printt .Repo}}

# add_params is the set of parameters that will be passed to repo-add
# when it is called. Specify one time for each parameter.
#add_params = {{printt .AddParameters}}

# rm_params is the set of parameters that will be passed to repo-remove
# when it is called. Specify one time for each parameter.
#rm_params = {{printt .RemoveParameters}}

# ignore_aur is a set of package names that are ignored in conjunction
# with AUR related tasks, such as determining if there is an update or not.
#ignore_aur = {{printt .IgnoreAUR}}

# backup specifies whether package files should be backed up or deleted.
# If it is set to false, then obsolete package files are deleted.
#backup = {{printt .Backup}}

# backup_dir specifies which directory backups are stored in.
# If a relative path is given, then 
#backup_dir = {{printt .BackupDir}}

# interactive specifies that repoctl should ask before doing anything
# destructive.
#interactive = {{printt .Interactive}}

# columnate specifies that listings should be in columns rather than
# in lines. This only applies to the list command.
#columnate = {{printt .Columnate}}

# quiet specifies whether repoctl should print more information or less.
# I prefer to know what happens, but if you don't like it, you can change it.
#quiet = {{printt .Quiet}}
`)).Funcs(template.FuncMap{
	"printt": printt,
})

// Configuration doubles as configuration file format and the global configuration set.
type Configuration struct {
	// Repository is the absolute path to the database. We assume that this is
	// also where the packages are. The variables database and path are constructed
	// from this.
	Repository string `toml:"repo"`
	database   string
	repodir    string

	// AddParameters are parameters to add to the repo-add command line.
	AddParameters []string `toml:"add_param"`
	// RemoveParameters are parameters to add to the repo-remove command line.
	RemoveParameters []string `toml:"rm_param"`
	// Packages to ignore when doing AUR related tasks.
	IgnoreAUR []string `toml:"ignore_aur`

	// Backup causes older packages to be backed up rather than deleted.
	Backup bool `toml:"backup"`
	// BackupDir specifies where old packages are backed up to.
	BackupDir string `toml:"backup_dir"`
	// Interactive requires confirmation before deleting and changing the
	// repository database.
	Interactive bool `toml:"interactive"`

	// Columnate causes items to be printed in columns rather than lines.
	Columnate bool `toml:"columnate"`
	// Quiet causes less information to be printed than usual.
	Quiet bool `toml:"quiet"`

	// When Debug is specified, it presides over Quiet.
	// This allows it to override a possible default value of Quiet.
	Debug bool `toml:"-"`
	// When Unconfigured is true, the program should fail.
	Unconfigured bool `toml:"unconfigured"`
}

// Conf acts as the global storage for the program configuration.
// It also contains the default values.
var Conf = &Configuration{
	BackupDir:    "backup/",
	Unconfigured: true,
}

// HomeConf is the path to the home configuration file.
//
// If REPOCTL_CONFIG is set, then HomeConf points to that.
func HomeConf() string {
	if p := os.Getenv("REPOCTL_CONFIG"); p != "" {
		return p
	} else {
		return xdgConfigHome(configurationPostfix)
	}
}

// NewConfiguration creates a new default configuration with repo as
// the repository database.
//
// If repo is invalid (because it is absolute), nil is returned.
// We don't check for existance, because at this point, it might
// not exist yet.
func NewConfiguration(repo string) *Configuration {
	if !path.IsAbs(repo) {
		return nil
	}
	return &Configuration{
		Repository: repo,
		database:   path.Base(repo),
		repodir:    path.Dir(repo),

		AddParameters:    make([]string, 0),
		RemoveParameters: make([]string, 0),
		IgnoreAUR:        make([]string, 0),
	}
}

// ReadConfiguration creates a new configuration by reading one
// from the specified path.
func ReadConfiguration(filepath string) (*Configuration, error) {
	c := &Configuration{}
	err := c.MergeFile(filepath)
	return c, err
}

// Initialize should be called every time after changing c.Repository.
// This ensures that the configuration is both valid and consistent.
//
// If the repository path is not absolute, ErrRepoNotAbs is returned.
func (c *Configuration) Initialize() error {
	if !path.IsAbs(c.Repository) {
		return ErrRepoNotAbs
	}
	c.database = path.Base(c.Repository)
	c.repodir = path.Dir(c.Repository)
	return nil
}

func (c *Configuration) WriteTemplate(w io.Writer) error {
	return configurationTmpl.Execute(w, c)
}

func (c *Configuration) WriteFile(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return c.WriteTemplate(file)
}

func (c *Configuration) MergeFile(filepath string) error {
	ex, err := osutil.FileExists(p)
	if err != nil {
		return err
	}
	if !ex {
		return ErrNoConfig
	}
	_, err := toml.DecodeFile(filepath, c)
	return err
}

// MergeFiles merges configuration files one after the other.
// At least one needs to exist. Later files will override earlier files.
//
// If there is an error at any point, we fail.
func (c *Configuration) MergeFiles(filepaths []string) error {
	var ok bool

	for _, p := range filepaths {
		if ex, _ := osutil.FileExists(p); ex {
			// Only try to read files that exist.
			c.MergeFile(p)
			ok = true
		}
	}

	if !ok {
		return ErrNoConfig
	}
	return nil
}

// MergeXDG tries to merge configuration files according to the XDG specification.
// First we merge all found files in XDG_CONFIG_DIRS in reverse order.
// We finish off with XDG_CONFIG_HOME.
func (c *Configuration) MergeXDG() error {
	paths := xdgConfigDirs(configurationPostfix)
	home, err := xdgConfigHome(configurationPostfix)
	if err != nil {
		if !c.Quiet {
			fmt.Fprintf(os.Stderr, "Warning: %s.\n", err)
		}
	} else {
		paths = append(paths, home)
	}

	return c.MergeFiles(paths)
}

// The last configuration is most important.
func xdgConfigDirs(postfix string) []string {
	sys := strings.Split(os.Getenv("XDG_CONFIG_DIRS"), ":")
	if len(sys) == 0 {
		return []string{path.Join("/etc/xdg", postfix)}
	}
	n := len(sys)
	paths := make([]string, n)
	for i := n - 1; i >= 0; i-- {
		paths[i] = path.Join(sys[n-i-1], postfix)
	}
	return paths
}

func xdgConfigHome(postfix string) (string, error) {
	cfg := os.Getenv("XDG_CONFIG_HOME")
	if cfg == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return "", ErrUnsetHOME
		}
		cfg = path.Join(home, ".config")
	}
	return path.Join(cfg, postfix), nil
}

// MergeAll performs the default configuration loading procedure,
// and sets c.Unconfigured accordingly.
func (c *Configuration) MergeAll() error {
	var err error

	// We assume for now, that loading a configuration will succeed.
	// If it doesn't, we'll change this back to true.
	c.Unconfigured = false

	// Normally, the configuration file is loaded according to the XDG
	// specification. This environment variable lets us override that.
	confpath = os.Getenv("REPOCTL_CONFIG")
	if confpath != "" {
		err = c.MergeFile(confpath)
	} else {
		err = c.MergeXDG()
	}
	if err != nil {
		c.Unconfigured = true
		return err
	} else if Conf.Repository == "" {
		c.Unconfigured = true
		return ErrRepoUnset
	} else if err = Conf.Initialize(); err != nil {
		c.Unconfigured = true
		return err
	}
	return nil
}

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
			buf.WriteString(fmt.Sprintf("%q", obj))
			buf.WriteString(", ")
		}
		buf.WriteString(fmt.Sprintf("%q", obj[len(obj)-1]))
		buf.WriteRune(']')
		return buf.String()
	default: // floats, ints, bools
		return fmt.Sprintf("%v", obj)
	}
}
