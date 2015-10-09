// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var (
	SystemRepoAdd    = "/usr/bin/repo-add"
	SystemRepoRemove = "/usr/bin/repo-remove"
)

// DatabaseAdd adds the given packages to the repository database.
func (r *Repo) DatabaseAdd(pkgfiles ...string) error {
	if len(pkgfiles) == 0 {
		return nil
	}
	return in(r.Directory, func() error {
		for _, p := range pkgfiles {
			r.printf("adding package to database: %s\n", p)
		}

		args := joinArgs(r.AddParameters, r.Database, pkgfiles)
		cmd := exec.Command(SystemRepoAdd, args...)
		return system(cmd, r.Error)
	})
}

func (r *Repo) DatabaseRemove(pkgnames ...string) error {
	if len(pkgnames) == 0 {
		return nil
	}
	return in(r.Directory, func() error {
		for _, p := range pkgnames {
			r.printf("removing package from database %s\n", p)
		}

		args := joinArgs(r.RemoveParameters, r.Database, pkgnames)
		cmd := exec.Command(SystemRepoRemove, args...)
		return system(cmd, r.Error)
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
func system(cmd *exec.Cmd, ew io.Writer) error {
	if ew != nil {
		return cmd.Run()
	}

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
		fmt.Fprintln(ew, line)
	}

	return cmd.Wait()
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
