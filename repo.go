// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/goulash/osutil"
)

type Repo struct {
	// Directory is the absolute path to the directory where the
	// packages are stored.
	Directory string
	// Database is the relative path to the repository database,
	// relative from Directory.
	Database string

	// Backup specifies whether to backup old packages
	Backup bool
	// BackupDir specifies where old packages are backed up to,
	// relative to the repository directory.
	BackupDir string

	// AddParameters are parameters to add to the repo-add
	// command line.
	AddParameters []string
	// RemoveParameters are parameters to add to the repo-remove
	// command line.
	RemoveParameters []string

	// Error, Info, and Debug is where output is written to. If set to
	// nil, no output is written.
	Error io.Writer
	Info  io.Writer
	Debug io.Writer
}

// New creates a new default configuration with repo as the repository
// database. It is assumed that the database resides in the same
// directory as the packages.
//
// If repo is invalid (because it is absolute), nil is returned.
// We don't check for database existance, because at this point,
// it might not exist yet.
func New(repo string) *Repo {
	if !path.IsAbs(repo) {
		return nil
	}

	return &Repo{
		Directory: path.Dir(repo),
		Database:  path.Base(repo),
		BackupDir: `backup`,

		Error: os.Stderr,
		Info:  os.Stdout,
		Debug: nil,

		AddParameters:    make([]string, 0),
		RemoveParameters: make([]string, 0),
	}
}

// AssertSetup returns nil if a normal repository setup is present:
// the directory exists.
//
// While it would make sense to check for readability and writability,
// in modern systems there are so many ways to achieve this, that to
// test all of them is more effort than it is worth.
func (r *Repo) AssertSetup() error {
	if !path.IsAbs(r.Directory) {
		return ErrRepoDirRelative
	}

	ex, err := osutil.DirExists(r.Directory)
	if err != nil {
		return err
	}
	if !ex {
		return ErrRepoDirMissing
	}

	return nil
}

// Setup creates the directory and returns an error if not possible.
func (r *Repo) Setup() error {
	if err := r.AssertSetup(); err != ErrRepoDirMissing {
		return err
	}

	return os.MkdirAll(r.Directory, os.ModePerm)
}

func (r *Repo) printf(format string, obj ...interface{}) {
	if r.Info != nil {
		fmt.Fprintf(r.Info, format, obj...)
	}
}

func (r *Repo) errorf(format string, obj ...interface{}) {
	if r.Error != nil {
		fmt.Fprintf(r.Error, format, obj...)
	}
}

func (r *Repo) debugf(format string, obj ...interface{}) {
	if r.Debug != nil {
		fmt.Fprintf(r.Debug, format, obj...)
	}
}
