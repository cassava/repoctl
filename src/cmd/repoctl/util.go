// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path"
	"sort"

	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
	"github.com/goulash/pr"
)

// getRepoPkgs retrieves the most up-to-date packages from the repository path,
// and returns all older packages in outdated.
func getRepoPkgs(repopath string) (pkgs []*pacman.Package, outdated []*pacman.Package) {
	// TODO/FUTURE: Handle errors in a more intelligent fashion. If a single package could not
	// be read, it's not the end of the world. If the entire path could not be read, then
	// there isn't much point continuing. How do you figure the difference out though?
	dirPkgs, _ := pacman.ReadDir(repopath, func(err error) error {
		fmt.Fprintln(os.Stderr, "error:", err)
		return nil
	})
	return pacman.SplitOld(dirPkgs)
}

// getDatabasePkgs retrieves the packages stored in the database.
// Any package that is referenced but does not exist is stored in missed.
func getDatabasePkgs(dbpath string) (db map[string]*pacman.Package, missed []*pacman.Package) {
	db = make(map[string]*pacman.Package)
	missed = make([]*pacman.Package, 0)
	pkgs, _ := pacman.ReadDatabase(dbpath)
	for _, p := range pkgs {
		if ex, _ := osutil.FileExists(p.Filename); !ex {
			missed = append(missed, p)
			continue
		}
		db[p.Name] = p
	}
	return db, missed
}

// getAURMap returns a map of the results of all the AUR searches pkgnames.
func getAURMap(pkgnames []string) (aur map[string]*pacman.Package) {
	ch := make(chan error)
	go handleErrors("warning: %s\n", ch)
	aur = pacman.ConcurrentlyReadAUR(pkgnames, 16, ch)
	close(ch)
	return aur
}

// getAURPkgs retrieves the packages listed from AUR.
// Packages that are not found are stored in missed.
func getAURPkgs(pkgnames []string) (aur map[string]*pacman.Package, missed []string) {
	aur = getAURMap(pkgnames)
	for k, v := range aur {
		if v == nil {
			missed = append(missed, k)
			delete(aur, k)
		}
	}

	return aur, missed
}

// handleErrors is meant to be launched as a separate goroutine to handle
// errors coming from ReadDir and the likes.
func handleErrors(format string, ch <-chan error) {
	for err := range ch {
		fmt.Fprintf(os.Stderr, format, err)
	}
}

// mapPkgs maps Packages to some string characteristic of a Package.
func mapPkgs(pkgs []*pacman.Package, f func(*pacman.Package) string) []string {
	results := make([]string, len(pkgs))
	for i, p := range pkgs {
		results[i] = f(p)
	}
	return results
}

func pkgFilename(p *pacman.Package) string {
	return p.Filename
}

func pkgBasename(p *pacman.Package) string {
	return path.Base(p.Filename)
}

func pkgName(p *pacman.Package) string {
	return p.Name
}

// printSet prints a set of items and optionally a header.
func printSet(list []string, h string, cols bool) {
	sort.Strings(list)
	if h != "" {
		fmt.Printf("\n%s\n", h)
	}
	if cols {
		pr.PrintFlex(list)
	} else if h != "" {
		for _, j := range list {
			fmt.Println(" ", j)
		}
	} else {
		for _, j := range list {
			fmt.Println(j)
		}
	}
}
