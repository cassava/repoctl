// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"

	flag "github.com/ogier/pflag"
)

const (
	progName    = "repoctl"
	progVersion = "1.9.9"
	progDate    = "26. May 2014"

	configPath = "~/.repo.conf"
)

// Config contains all the configuration flags, variables, and arguments that
// are needed for the various actions.
type Config struct {
	// ConfigFile stores the name of the configuration file from which this
	// configuration was loaded from, otherwise it is empty.
	ConfigFile string

	// RepoPath is the path to where all the package files and the database
	// reside. It doesn't matter whether it ends in a "/" or not.
	RepoPath string
	// Database stores the name of the repository database. This is the file
	// that is usually has the ".db.tar.gz" suffix.
	Database string
	// AddParameters are parameters to add to the repo-add command line.
	AddParameters []string
	// RemoveParameters are parameters to add to the repo-remove command line.
	RemoveParameters []string

	// Verbose causes more information to be printed than usual.
	// Default is false.
	Verbose bool

	// Columnated causes packages to be printed in columns.
	// Default is true.
	Columnated bool
	// Versioned causes packages to be printed with version information.
	// Default is false.
	Versioned bool
	// Duplicates causes the number of outdated packages to be printed along
	// with the packages.
	// Default is false.
	Duplicates bool

	// Confirm requires confirmation before deleting and changing the
	// repository database.
	// Default is false.
	Confirm bool
	// Delete causes older packages to be deleted if there is a newer one.
	// Default is true.
	Delete bool
	// UpdateByAge causes the update function to only consider the modification
	// times of the packages in reference to the database. Newer packages are
	// added to the database; older aren't. This is faster than a more thorough
	// update, but may miss some packages in certain situations.
	// Default is false.
	UpdateByAge bool

	// Arguments contains the argumetns given on the commandline.
	Args []string
}

// Action is the type that all action functions need to satisfy.
type Action func(*Config) error

// actions is a map from names to action functions.
var actions map[string]Action = map[string]Action{
	"list":        List,
	"ls":          List,
	"update":      Update,
	"add":         Add,
	"remove":      Remove,
	"rm":          Remove,
	"synchronize": Sync,
	"sync":        Sync,
	"help":        Usage,
	"usage":       Usage,
}

// NewConfig creates a minimal configuration.
func NewConfig(repoPath, db string) *Config {
	return &Config{
		RepoPath: repoPath,
		Database: db,

		// Set the default values as documented in Config.
		Columnated: true,
		Delete:     true,
	}
}

// NewConfigFromFile reads a configuration from a file.
func NewConfigFromFile(path string) (conf *Config, err error) {
	return nil, nil
}

// NewConfigFromFlags reads a configuration from the command line arguments.
//
// TODO: Implement Config file reading and merging
func NewConfigFromFlags() (conf *Config, cmd Action, err error) {
	conf = &Config{}

	flag.StringVarP(&conf.ConfigFile, "config", "c", configPath, "configuration file to load settings from")
	flag.StringVar(&conf.RepoPath, "repo", "/srv/abs", "the path to where the packages and database reside")
	flag.StringVar(&conf.Database, "db", "atlas.db.tar.gz", "the name of the database")

	flag.BoolVarP(&conf.Verbose, "verbose", "v", false, "print more information")

	flag.BoolVarP(&conf.Columnated, "columns", "s", true, "print packages in columns like ls")
	flag.BoolVarP(&conf.Versioned, "versioned", "V", false, "print the version of each package when listing")
	flag.BoolVarP(&conf.Duplicates, "duplicates", "d", false, "mark the number of duplicate (outdated) packages")

	flag.BoolVarP(&conf.Confirm, "confirm", "i", false, "confirm before deleting and changing the repo db")
	flag.BoolVarP(&conf.Delete, "delete", "r", true, "delete outdated packages")
	flag.BoolVarP(&conf.UpdateByAge, "fast-update", "f", false, "determine which packages to update by age")

	flag.Usage = func() { Usage(nil) }
	flag.Parse()
	if len(flag.Args()) == 0 {
		return nil, Usage, errors.New("no action specified on command line")
	}
	conf.Args = flag.Args()[1:]
	cmd, ok := actions[flag.Arg(0)]
	if !ok {
		return nil, Usage, errors.New("unrecognized action " + flag.Arg(0))
	}

	return conf, cmd, nil
}

// Usage prints the help message for the program.
// TODO: Print option usage too!
func Usage(*Config) error {
	fmt.Printf("%s %s (%s)\n", progName, progVersion, progDate)
	fmt.Print(`
Manage local pacman repositories.

Commands available:
  add <pkgname>    Add the package(s) with <pkgname> to the database by
                   finding in the same directory of the database the latest
                   file for that package (by file modification date),
                   deleting the others, and updating the database.
  list             List all the packages that are currently available.
  (ls)             Note that this has nothing to do with the database.
  remove <pkgname> Remove the package with <pkgname> from the database, by
  (rm)             removing its entry from the database and deleting the files
                   that belong to it.
  update           Same as add, except scan and add changed packages.
  synchronize      Compare packages in the database to AUR for new versions.
  (sync)

NOTE: In all of these cases, <pkgname> is the name of the package, without
anything else. For example: pacman, and not pacman-3.5.3-1-i686.pkg.tar.xz

Options:
`)
	flag.PrintDefaults()
	fmt.Println()

	return nil
}
