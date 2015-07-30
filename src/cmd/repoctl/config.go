// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Configuration
//
// So I stand in front of the classic problem: how to get configuration from various sources
// and make that it works nicely. I want a simple system that I can combine with Cobra.
//
// For that though, I need to figure out how to do things...
//
// Idea 1: create a configuration file that links to 

package main

import (
	"errors"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
)

type Config struct{
    Struct interface{}
    Template string
}

type RepoctlConfig struct{
	Repo        *string   `toml:"repo"`
	AddParam    *[]string `toml:"add_params"`
	RmParam     *[]string `toml:"rm_params"`
	IgnoreAUR   *[]string `toml:"ignore_aur"`
	Quiet       *bool     `toml:"quiet"`
	Interactive *bool     `toml:"interactive"`
	Backup      *bool     `toml:"backup"`
	BackupDir   *string   `toml:"backup_dir"`
}

var repoConfig Config{
    Struct: RepoctlConfig{
        Repo: &Repository,
        AddParam: &AddParameters,
        RmParam: &RemoveParameters,
        IgnoreAUR: 
    },
    Template: `# repoctl config

# repo is the full path to the repository that will be managed by repoctl.
# The packages that belong to the repository are assumed to lie in the
# same folder.
repo = "{{.Repo}}"

# add_params is the set of parameters that will be passed to repo-add
# when it is called. Specify one time for each parameter.
add_params = []

# rm_params is the set of parameters that will be passed to repo-remove
# when it is called. Specify one time for each parameter.
rm_params = []

# ignore_aur is a set of package names that are ignored in conjunction
# with AUR related tasks, such as determining if there is an update or not.
ignore_aur = []

# quiet specifies whether repoctl should print more information or less.
# I prefer to know what happens, but if you don't like it, you can change it.
quiet = {{.Quiet}}

# interactive specifies that repoctl should ask before doing anything
# destructive.
interactive = {{.Interactive}}

# backup specifies whether package files should be backed up or deleted.
# If it is set to false, then obsolete package files are deleted.
backup = {{.Backup}}

# backup_dir specifies which directory backups are stored in.
# If a relative path is given, then 
backup_dir = {{.BackupDir}}
`,
    

type RepoConfig struct {
}

var (
	ErrRequireRepository = errors.New("path to repository missing")
	ErrConfigUnmodified  = errors.New("configuration needs adjusting (default set)")
)

func ReadRepoConfig(path string) (*RepoConfig, error) {
	rc := &RepoConfig{}
	_, err := toml.DecodeFile(path, rc)
	if err != nil {
		return nil, err
	}

	if rc.Repo == "" {
		return nil, ErrRequireRepository
	} else if rc.Default == true {
		return nil, ErrConfigUnmodified
	}

	return rc, nil
}

func (rc RepoConfig) OverwriteGlobal(conf *Config) {
	Repository = rc.Repo
	AddParameters = rc.AddParam
	RemoveParameters = rc.RmParam
	for _, k := range rc.IgnoreAUR {
		IgnoreAUR[k] = true
	}
	Quiet = rc.Quiet
	Interactive = rc.Interactive
	Backup = rc.Backup
	BackupDir = rc.BackupDir
}

func (rc RepoConfig) WriteDefault(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return repoConfigTmpl.Execute(file, rc)
}
