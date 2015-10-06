// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "os/exec"

var (
	SystemRepoAdd    = "/usr/bin/repo-add"
	SystemRepoRemove = "/usr/bin/repo-remove"
)

// DatabaseAdd adds the given packages to the repository database.
func (r *Repo) DatabaseAdd(pkgfiles ...string) error {
	return in(r.Directory, func() error {
		for _, p := range pkgfiles {
			r.printf("adding package to database: %s\n", p)
		}

		args := joinArgs(r.AddParameters, r.Database, pkgfiles)
		cmd := exec.Command(SystemRepoAdd, args...)
		return system(cmd, r.Error)
	})
}

func (r *Repo) DatabaseRemove(pkgfiles ...string) error {
	return in(r.Directory, func() error {
		for _, p := range pkgfiles {
			r.printf("removing package from database %s\n", p)
		}

		args := joinArgs(r.RemoveParameters, r.Database, pkgnames)
		cmd := exec.Command(sysRepoRemove, args...)
		return system(cmd, r.Error)
	})
}
