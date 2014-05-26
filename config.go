// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

const (
	progName    = "repoctl"
	progVersion = "1.9.9"
	progDate    = "26. May 2014"

	configPath = "~/.repo.conf"
)

// Config contains all the configuration flags, variables, and arguments that
// are needed for the various actions.
//
// The following configuration variables are read by:
//
//	List:	RepoPath,Verbose,Columnated,Versioned,Duplicates
//	Add:	RepoPath,Database,Verbose,Confirm,
//  Remove: RepoPath,Database,Verbose,
//  Update: RepoPath,Database,Verbose,FastUpdate,Confirm,
//  Sync:	RepoPath,Verbose,
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

func NewConfig(repoPath, db string) *Config {
	return &Config{
		RepoPath: repoPath,
		Database: db,

		// Set the default values as documented in Config.
		Columnated: true,
		Delete:     true,
	}
}

func ReadConfig(path string) *Config {
	return nil
}
