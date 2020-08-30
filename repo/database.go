// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman"
	"github.com/goulash/osutil"
)

var (
	SystemRepoAdd    = "/usr/bin/repo-add"
	SystemRepoRemove = "/usr/bin/repo-remove"
)

// DeleteDatabase deletes the repository database (but not the files).
func (r *Repo) DeleteDatabase() error {
	dbpath := r.DatabasePath()
	if ex, _ := osutil.FileExists(dbpath); ex {
		term.Printf("Deleting database: %s\n", dbpath)
		return os.Remove(dbpath)
	}
	return nil
}

// CreateDatabase creates the repository database (but does nothing else).
func (r *Repo) CreateDatabase() error {
	dbpath := r.DatabasePath()
	if ex, _ := osutil.FileExists(dbpath); ex {
		return nil
	}
	if ex, _ := osutil.DirExists(r.Directory); !ex {
		r.Setup()
	}

	term.Printf("Creating database: %s\n", dbpath)
	args := joinArgs(r.AddParameters, dbpath)
	cmd := exec.Command(SystemRepoAdd, args...)
	return r.system(cmd)
}

// AddToDatabase adds the given packages to the repository database.
func (r *Repo) AddToDatabase(pkgfiles ...string) error {
	if len(pkgfiles) == 0 {
		return nil
	}

	dbpath := r.DatabasePath()
	if pacman.IsDatabaseLocked(dbpath) {
		return fmt.Errorf("database is locked: %s.lck", dbpath)
	}

	return in(r.Directory, func() error {
		for _, p := range pkgfiles {
			term.Printf("Adding package to database: %s\n", p)
		}

		args := joinArgs(r.AddParameters, r.Database, pkgfiles)
		cmd := exec.Command(SystemRepoAdd, args...)
		return r.system(cmd)
	})
}

// RemoveFromDatabase removes the given packages from the repository database.
func (r *Repo) RemoveFromDatabase(pkgnames ...string) error {
	if len(pkgnames) == 0 {
		return nil
	}

	dbpath := r.DatabasePath()
	if pacman.IsDatabaseLocked(dbpath) {
		return fmt.Errorf("database is locked: %s.lck", dbpath)
	}

	return in(r.Directory, func() error {
		for _, p := range pkgnames {
			term.Printf("Removing package from database: %s\n", p)
		}

		args := joinArgs(r.RemoveParameters, r.Database, pkgnames)
		cmd := exec.Command(SystemRepoRemove, args...)
		return r.system(cmd)
	})
}

// joinArgs joins strings and arrays of strings together into one array.
func joinArgs(args ...interface{}) []string {
	var final []string
	for _, a := range args {
		switch a.(type) {
		case string:
			final = append(final, a.(string))
		case []string:
			final = append(final, a.([]string)...)
		default:
			final = append(final, fmt.Sprint(a))
		}
	}
	return final
}

// system runs cmd, and prints the stderr output to ew, if ew is not nil.
func (r *Repo) system(cmd *exec.Cmd) error {
	command := strings.Join(cmd.Args, " ")
	term.Debugf("Executing: %s\n", command)

	bs, err := cmd.CombinedOutput()
	if err != nil {
		term.Errorf("Error executing: %s\n", command)
		term.Errorff("---\n")
		term.Errorff("%s", bs)
		term.Errorff("...\n")
		return fmt.Errorf("command exited with non-zero return code: %s", command)
	}
	return nil
}

// in performs a function in a directory, and then returns to the
// previous directory.
func in(dir string, f func() error) (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	err = os.Chdir(dir)
	if err != nil {
		return err
	}
	defer func() {
		cerr := os.Chdir(cwd)
		if err == nil {
			err = cerr
		}
	}()
	err = f()
	return
}
