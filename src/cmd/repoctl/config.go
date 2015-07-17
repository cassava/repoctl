// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
)

var repoConfigTmpl = template.Must(template.New("repoctl").Parse(`# repoctl config

# repo is the full path to the repository that will be managed by repoctl.
# The packages that belong to the repository are assumed to lie in the
# same folder.
repo = "{{ .Repo }}"

# Remove the following line when you are done editing this file.
default = true

# add_params is the set of parameters that will be passed to repo-add
# when it is called. Specify one time for each parameter.
#add_params = []

# rm_params is the set of parameters that will be passed to repo-remove
# when it is called. Specify one time for each parameter.
#rm_params = []

# ignore_aur is a set of package names that are ignored in conjunction
# with AUR related tasks, such as determining if there is an update or not.
#ignore_aur = []
`))

type RepoConfig struct {
	Repo      string   `toml:"repo"`
	AddParam  []string `toml:"add_params"`
	RmParam   []string `toml:"rm_params"`
	IgnoreAUR []string `toml:"ignore_aur"`

	Default bool `toml:"default"`
}

func ReadRepoConfig(path string) (*RepoConfig, error) {
	rc := &RepoConfig{}
	_, err := toml.DecodeFile(path, rc)
	if err != nil {
		return nil, err
	}

	if rc.Repo == "" {
		return nil, errors.New("path to repository missing")
	}

	return rc, nil
}

func (rc RepoConfig) MergeIntoConfig(conf *Config) {
	if conf.Repository == "" {
		conf.Repository = rc.Repo
	}
	conf.AddParameters = rc.AddParam
	conf.RemoveParameters = rc.RmParam
	for _, k := range rc.IgnoreAUR {
		conf.IgnoreAUR[k] = true
	}
}

func (rc RepoConfig) OverwriteConfig(conf *Config) {
	conf.Repository = rc.Repo
	conf.AddParameters = rc.AddParam
	conf.RemoveParameters = rc.RmParam
	for _, k := range rc.IgnoreAUR {
		conf.IgnoreAUR[k] = true
	}
}

func (rc RepoConfig) WriteDefault(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return repoConfigTmpl.Execute(file, rc)
}
