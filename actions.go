// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goulash/pacman"
)

const (
	sysRepoAdd    = "/usr/bin/repo-add"
	sysRepoRemove = "/usr/bin/repo-remove"
)

var implErr error = errors.New("unimplemented functionality")

// Add finds the newest packages given in pkgs and adds them, removing the old
// packages.
func Add(c *Config) error {
	pkgs := pacman.ReadMatchingNames(c.RepoPath, c.Args, nil)
	return updatePackages(c, pkgs)
}

func Remove(c *Config) error {
	return implErr
}

func Update(c *Config) error {
	return implErr
}

func updatePackages(c *Config, pkgs []*pacman.Package) error {
	updated, old := pacman.SplitOld(pkgs)

	if c.Interactive {
		fmt.Println("The following packages will be added to the database:")
		return implErr
	}
	addPackages(c, updated)

	if c.Interactive {
		fmt.Println("The following outdated packages will be deleted:")
		return implErr
	}
	for _, p := range old {
		c.inform(fmt.Sprintf("removing %s...", p.Filename))
		err := os.Remove(p.Filename)
		if err != nil {
			fmt.Printf("error:", err)
		}
	}

	return nil
}

// addPackages adds all the packages listed from the database.
func addPackages(c *Config, pkgs []*pacman.Package) error {
	dbpath := filepath.Join(c.RepoPath, c.Database)
	args := joinArgs(c.AddParameters, dbpath, extractFilenames(pkgs))

	cmd := exec.Command(sysRepoAdd, args...)
	return runStderr(cmd)
}

// removePackage removes all the packages listed from the database.
func removePackages(c *Config, pkgs []*pacman.Package) error {
	dbpath := filepath.Join(c.RepoPath, c.Database)
	args := joinArgs(c.RemoveParameters, dbpath, extractFilenames(pkgs))

	cmd := exec.Command(sysRepoRemove, args...)
	return runStderr(cmd)
}

// extractFilenames maps the filenames of the packages into an array.
func extractFilenames(pkgs []*pacman.Package) []string {
	names := make([]string, len(pkgs))
	for i := range names {
		names[i] = pkgs[i].Filename
	}
	return names
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

// runStderr runs the given command and routes all error messages from the
// program out the stderr of this program.
func runStderr(cmd *exec.Cmd) error {
	out, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	rd := bufio.NewReader(out)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		fmt.Fprintln(os.Stderr, line)
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
