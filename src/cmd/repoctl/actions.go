// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
)

const (
	sysRepoAdd    = "/usr/bin/repo-add"
	sysRepoRemove = "/usr/bin/repo-remove"

	backupDir = "backup/"
)

// Add finds the newest packages given in pkgs and adds them, removing the old
// packages.
func Add(c *Config) error {
	// TODO: handle the errors here correctly!
	pkgs, _ := pacman.ReadMatchingNames(c.path, c.Args, nil)
	pkgs, outdated := pacman.SplitOld(pkgs)
	db, _ := getDatabasePkgs(c.Repository)
	pending := filterPkgs(pkgs, dbPendingFilter(db))

	if c.Interactive {
		backup := "Delete following files:"
		if c.Backup {
			backup = "Back following files up:"
		}
		proceed := confirmAll(
			[][]string{
				mapPkgs(pending, pkgNameVersion(db)),
				mapPkgs(outdated, pkgBasename),
			},
			[]string{
				"Add following entries to database:",
				backup,
			},
			c.Columnate)
		if !proceed {
			return nil
		}
	}

	var err error
	if len(pending) > 0 {
		err = addPkgs(c, mapPkgs(pending, pkgFilename))
		if err != nil {
			return err
		}
	}
	if len(outdated) > 0 {
		filenames := mapPkgs(outdated, pkgFilename)
		if c.Backup {
			err = backupPkgs(c, filenames)
		} else {
			err = deletePkgs(c, filenames)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func Remove(c *Config) error {
	// TODO: handle the errors here correctly!
	pkgs, _ := pacman.ReadMatchingNames(c.path, c.Args, nil)
	db, _ := getDatabasePkgs(c.Repository)

	rmmap := make(map[string]bool)
	for _, p := range pkgs {
		rmmap[p.Name] = true
	}
	dbpkgs := make([]string, 0, len(rmmap))
	for k := range rmmap {
		if _, ok := db[k]; ok {
			dbpkgs = append(dbpkgs, k)
		}
	}

	if c.Interactive {
		backup := "Delete following files:"
		if c.Backup {
			backup = "Back following files up:"
		}
		proceed := confirmAll(
			[][]string{
				dbpkgs,
				mapPkgs(pkgs, pkgBasename),
			},
			[]string{
				"Remove following entries from database:",
				backup,
			},
			c.Columnate)
		if !proceed {
			return nil
		}
	}

	var err error
	if len(dbpkgs) > 0 {
		err = removePkgs(c, dbpkgs)
		if err != nil {
			return err
		}
	}
	if len(pkgs) > 0 {
		files := mapPkgs(pkgs, pkgFilename)
		if c.Backup {
			err = backupPkgs(c, files)
		} else {
			err = deletePkgs(c, files)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func Update(c *Config) error {
	pkgs, outdated := getRepoPkgs(c.path)
	db, missed := getDatabasePkgs(c.Repository)
	pending := filterPkgs(pkgs, dbPendingFilter(db))

	if c.Interactive {
		backup := "Delete following files:"
		if c.Backup {
			backup = "Back following files up:"
		}
		proceed := confirmAll(
			[][]string{
				mapPkgs(missed, pkgName),
				mapPkgs(pending, pkgNameVersion(db)),
				mapPkgs(outdated, pkgBasename),
			},
			[]string{
				"Remove following entries from database:",
				"Update following entries in database:",
				backup,
			},
			c.Columnate)
		if !proceed {
			return nil
		}
	}

	var err error
	if len(missed) > 0 {
		err = removePkgs(c, mapPkgs(missed, pkgName))
		if err != nil {
			return err
		}
	}
	if len(pending) > 0 {
		err = addPkgs(c, mapPkgs(pending, pkgFilename))
		if err != nil {
			return err
		}
	}
	if len(outdated) > 0 {
		filenames := mapPkgs(outdated, pkgBasename)
		if c.Backup {
			err = backupPkgs(c, filenames)
		} else {
			err = deletePkgs(c, filenames)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// addPkgs adds all the packages listed from the database.
func addPkgs(c *Config, pkgfiles []string) error {
	args := joinArgs(c.AddParameters, c.Repository, pkgfiles)

	if !c.Quiet {
		forallPrintf("adding package to database: %s\n", pkgfiles)
	}

	cmd := exec.Command(sysRepoAdd, args...)
	return runStderr(cmd)
}

// removePkgs removes all the packages listed from the database.
func removePkgs(c *Config, pkgnames []string) error {
	args := joinArgs(c.RemoveParameters, c.Repository, pkgnames)

	if !c.Quiet {
		forallPrintf("removing package from database: %s\n", pkgnames)
	}

	cmd := exec.Command(sysRepoRemove, args...)
	return runStderr(cmd)
}

// deletePkgs deletes the given files.
func deletePkgs(c *Config, pkgfiles []string) error {
	os.Chdir(c.path)
	for _, p := range pkgfiles {
		if !c.Quiet {
			fmt.Println("deleting package file:", p)
		}
		err := os.Remove(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}
	}

	return nil
}

// backupPkgs backs up the given files.
func backupPkgs(c *Config, pkgfiles []string) error {
	backup := path.Join(c.path, backupDir)
	ex, err := osutil.DirExists(backup)
	if err != nil {
		return err
	} else if !ex {
		if !c.Quiet {
			fmt.Println("creating backup directory:", backup)
		}
		err = os.Mkdir(backup, os.ModePerm)
		if err != nil {
			return err
		}
	}

	for _, p := range pkgfiles {
		dest := path.Join(backup, fmt.Sprintf("%s.bak", p))
		src := path.Join(c.path, p)
		if !c.Quiet {
			fmt.Println("backing up file:", p)
		}
		err = os.Rename(src, dest)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}
	}

	return nil
}

// forallPrintf prints for each item according to the format.
func forallPrintf(format string, set []string) {
	for _, s := range set {
		fmt.Printf(format, s)
	}
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

func pkgNameVersion(db map[string]*pacman.Package) func(*pacman.Package) string {
	return func(p *pacman.Package) string {
		dp, ok := db[p.Name]
		if ok {
			return fmt.Sprintf("%s %s -> %s", p.Name, dp.Version, p.Version)
		}
		return fmt.Sprintf("%s -> %s", p.Name, p.Version)
	}
}

// confirmAll prints all the headings, items, and then returns the user's decision.
func confirmAll(sets [][]string, hs []string, cols bool) bool {
	nothing := true
	for i, s := range sets {
		if len(s) > 0 {
			printSet(s, hs[i], cols)
			nothing = false
		}
	}
	if nothing {
		return false
	}
	fmt.Println()
	return confirm()
}

// confirm gets the user's decision.
func confirm() bool {
	fmt.Print("Proceed? [Yn] ")
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.TrimSpace(strings.ToLower(line))
	return !(line == "no" || line == "n")
}
