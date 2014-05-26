// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/goulash/pr"
)

const (
	sysRepoAdd    = "/usr/bin/repo-add"
	sysRepoRemove = "/usr/bin/repo-remove"
)

// List displays all the packages available for the database.
// Note that they don't need to be registered with the database.
func List(c *Config) {
	pkgs := GetAllPackages(c.RepoPath)
	updated, old := SplitOldPackages(pkgs)

	// Find out how many old duplicates each package has.
	dups := make(map[string]int)
	for _, p := range old {
		dups[p.Name]++
	}

	// Create a list.
	var pkgnames []string
	for _, p := range updated {
		name := p.Name
		if c.Versioned {
			name += fmt.Sprintf(" %s", p.Version)
		}
		if c.Duplicates && dups[p.Name] > 0 {
			name += fmt.Sprintf(" [%v]", dups[p.Name])
		}
		pkgnames = append(pkgnames, name)
	}
	// While GetAllPackages
	sort.Strings(pkgnames)

	// Print packages to stdout
	if c.Columnated {
		pr.PrintAutoGrid(pkgnames)
	} else {
		for _, pkg := range pkgnames {
			fmt.Println(pkg)
		}
	}
}

// Add finds the newest packages given in pkgs and adds them, removing the old
// packages.
func Add(c *Config) {
	pkgs := GetAllMatchingPackages(c.RepoPath, c.Args)
	updated, old := SplitOldPackages(pkgs)

	// Find out how many old duplicates each package has.
	dups := make(map[string]int)
	for _, p := range old {
		dups[p.Name]++
	}

	if c.Confirm {
		fmt.Println("The following packages will be added to the database:")
		// TODO
		return
	}
	addPackages(updated)

	if c.Delete {
		if c.Confirm {
			fmt.Println("The following outdated packages will be deleted:")
			// TODO
			return
		}
		for _, p := range old {
			if c.Verbose {
				fmt.Printf("removing %s...", p.Filename)
			}
			err := os.Remove(p.Filename)
			if err != nil {
				fmt.Printf("error:", err)
			}
		}
	}
}

func Remove(c *Config) {

}

func Update(c *Config) {

}

func Sync(c *Config) {

}

func addPackages(pkgs []*Package) {

}

func removePackages(pkgs []*Package) {

}
