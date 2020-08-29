// Copyright (c) 2020, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package conf

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/goulash/osutil"
	"github.com/goulash/xdg"
)

var (
	ErrUnsetHOME       = errors.New("HOME environment variable unset")
	ErrRepoNotAbs      = errors.New("repository path must be absolute")
	ErrRepoUnset       = errors.New("repository path must be set in configuration")
	ErrProfileRequired = errors.New("profile must be specified or default profile set")
	ErrProfileUnknown  = errors.New("specified profile not found")
)

const configurationFile = "repoctl/config.toml"
const configurationEnv = "REPOCTL_CONFIG"

// HomeConf is the path to the home configuration file, which we can write to.
//
// If REPOCTL_CONFIG is set, then HomeConf points to that.
func HomeConf() string {
	if p := os.Getenv(configurationEnv); p != "" {
		return p
	} else {
		return xdg.UserConfig(configurationFile)
	}
}

type Configuration struct {
	// Columnate causes items to be printed in columns rather than lines.
	Columnate bool `toml:"columnate"`

	// Color causes repoctl output to be colorized.
	Color string `toml:"color"`

	// Quiet causes less information to be printed than usual.
	Quiet bool `toml:"quiet"`

	// When Debug is specified, it presides over Quiet.
	// This allows it to override a possible default value of Quiet.
	Debug bool `toml:"-"`

	// When CurrentProfile is specified, it presides over DefaultProfile.
	// This allows it to override the default, and is what we use for
	// profile selection from the command line.
	CurrentProfile string `toml:"-"`

	// DefaultProfile to use when no other profile is specified.
	DefaultProfile string `toml:"default_profile"`

	Profiles map[string]*Profile `toml:"profiles"`
}

// Default acts as the default configuraton and contains no profiles.
func Default() *Configuration {
	return &Configuration{
		Color:          "auto",
		DefaultProfile: "default",
		Profiles:       map[string]*Profile{},
	}
}

// New creates a new configuration with a default profile.
func New(repo string) *Configuration {
	return &Configuration{
		Color:          "auto",
		DefaultProfile: "default",
		Profiles: map[string]*Profile{
			"default": NewProfile(repo),
		},
	}
}

// Read creates a new configuration by reading one from the specified path.
func Read(filepath string) (*Configuration, error) {
	c := Default()
	err := c.MergeFile(filepath)
	return c, err
}

// FindAll loads all configuration files it finds in the search path.
//
// - If REPOCTL_CONFIG is set, this function will only use the path
//   specified in that variable.
func FindAll() (*Configuration, error) {
	// Normally, the configuration file is loaded according to the XDG
	// specification. This environment variable lets us override that.
	confpath := os.Getenv(configurationEnv)
	if confpath != "" {
		return Read(confpath)
	}

	// Create a default configuration and merge any files we find into that.
	c := Default()
	err := xdg.MergeConfigR(configurationFile, func(p string) error {
		err := c.MergeFile(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %s.\n", err)
		}
		return nil
	})
	return c, err
}

// SelectProfile returns the current "selected" profile.
//
// - CurrentProfile has priority over DefaultProfile.
// - If no profile is selected, then (nil, "", nil) is returned.
// - An error is only returned if a profile is selected but is not available.
//
// Before using the profile for the first time, make sure you call the
// profile Init() method.
func (c *Configuration) SelectProfile() (*Profile, string, error) {
	name := ""
	if c.CurrentProfile != "" {
		name = c.CurrentProfile
	} else if c.DefaultProfile != "" {
		name = c.DefaultProfile
	}
	if name == "" {
		return nil, "", nil
	}
	p, err := c.GetProfile(name)
	return p, name, err
}

func (c *Configuration) GetProfile(name string) (*Profile, error) {
	p, ok := c.Profiles[name]
	if !ok {
		return nil, ErrProfileUnknown
	}
	return p, nil
}

// MergeFile merges the contents of a file into the configuration here.
//
// It will warn the user on the use of deprecated fields, and it will
// also auto-migrate old configuration files.
func (c *Configuration) MergeFile(filepath string) error {
	ex, err := osutil.FileExists(filepath)
	if err != nil {
		return err
	}
	if !ex {
		return fmt.Errorf("cannot open %s: file not found", filepath)
	}

	// Deserialize the data into the file.
	bs, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", filepath, err)
	}
	md, err := toml.Decode(string(bs), c)
	if err != nil {
		return fmt.Errorf("cannot decode %s: %w", filepath, err)
	}

	// If there were extra fields that weren't decoded into c, then the TOML
	// decoder won't complain, so we need to check this ourselves.
	if len(md.Undecoded()) == 0 {
		return nil
	}

	// If there are undecoded values in the data, then this could either be
	// a user error or the use of the old configuration format.
	undecoded := make(map[string]bool)
	for _, f := range md.Undecoded() {
		undecoded[f.String()] = true
	}

	// Fortunately, the new configuration format is spread over two structs
	// now, and there is a good mapping between them. We shall try to
	// deserialize into the current/default profile.
	p, name, _ := c.SelectProfile()
	if p == nil {
		p = DefaultProfile()
		if name == "" {
			name = "default"
		}
		c.Profiles[name] = p
	}

	// Try to decode the data into the profile.
	md, err = toml.Decode(string(bs), p)
	if err != nil {
		return fmt.Errorf("cannot decode into profile %s: %w", filepath, err)
	}

	invalid := make([]string, 0)
	for _, f := range md.Undecoded() {
		s := f.String()

		// If it was undecoded in the original deserialization, then it hasn't
		// been deserialized at all, and is therefore invalid.
		if undecoded[s] {
			// There are two exceptions:
			if s == "action_on_completion" {
				fmt.Fprintf(os.Stderr, "Warning: option \"action_on_completion\" is deprecated; this is now always disabled.")
				continue
			}
			if s == "unconfigured" {
				fmt.Fprintf(os.Stderr, "Warning: option \"unconfigured\" is deprecated; this is now irrelevant.")
				continue
			}
			invalid = append(invalid, s)
		}
	}
	if len(invalid) != 0 {
		// We successfully migrated, so print a warning.
		return fmt.Errorf("cannot decode unknown fields: %q", invalid)
	}
	return nil
}

func (c *Configuration) WriteTemplate(w io.Writer) error {
	return ConfigurationTmpl.Execute(w, c)
}

func (c *Configuration) WriteProperties(w io.Writer) error {
	return PropertiesTmpl.Execute(w, c)
}

func (c *Configuration) WriteFile(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return c.WriteTemplate(file)
}
