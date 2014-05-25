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

type ListFlag int

const (
	// ListDefault effects no change to the way packages are printed. The
	// default is to print all package names in columns.
	ListDefault ListFlag = 0

	// OnePackagePerLine causes each package to be printed on it's own line
	// instead of printing them in columns.
	OnePackagePerLine ListFlag = 1 << iota

	// ShowVersion causes the (highest) package version to be shown:
	ShowVersion

	// ShowDuplicates causes the number of duplicate packages to be shown.
	ShowDuplicates
)

func (f ListFlag) Is(o ListFlag) bool {
	return f&o != 0
}

// List displays all the packages available for the database.
// Note that they don't need to be registered with the database.
func List(dbdir string, flags ListFlag) {
	pkgs := GetAllPackages(dbdir)
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
		if flags.Is(ShowVersion) {
			name += fmt.Sprintf(" %s", p.Version)
		}
		if flags.Is(ShowDuplicates) && dups[p.Name] > 0 {
			name += fmt.Sprintf(" [%v]", dups[p.Name])
		}
		pkgnames = append(pkgnames, name)
	}
	// While GetAllPackages
	sort.Strings(pkgnames)

	// Print packages to stdout
	if flags.Is(OnePackagePerLine) {
		for _, pkg := range pkgnames {
			fmt.Println(pkg)
		}
	} else {
		pr.PrintAutoGrid(pkgnames)
	}
}

type ModFlag int

const (
	ModDefault ModFlag = 0
	Confirm    ModFlag = 1 << iota
	NoDelete
	Verbose
)

func (f ModFlag) Is(o ModFlag) bool {
	return f&o != 0
}

// Add finds the newest packages given in pkgs and adds them, removing the old
// packages.
func Add(dbdir, dbname string, pkgs []string, flags ModFlag) {
	pkgs := GetAllMatchingPackages(dbdir, pkgs)
	updated, old := SplitOldPackages(pkgs)

	// Find out how many old duplicates each package has.
	dups := make(map[string]int)
	for _, p := range old {
		dups[p.Name]++
	}

	addPackages(updated)

	if !flags.Is(NoDelete) {
		for _, p := range old {
			if flags.Is(Verbose) {
				fmt.Printf("removing %s...", p.Filepath)
				// TODO: continue here...
			}
			err := os.Remove(p.Filepath)
		}
	}
}

func Remove(dbdir, dbname string, pkgs []string, flags ModFlag) {

}

func Update(dbdir, dbname string, flags ModFlag) {

}

func Sync(dbdir string) {

}
