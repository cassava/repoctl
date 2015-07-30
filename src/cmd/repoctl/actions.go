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
)

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
