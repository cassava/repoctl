// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

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

// getAURPkgs retrieves the packages listed from AUR.
// Packages that are not found are stored in missed.
func getAURPkgs(pkgnames []string) (aur map[string]*pacman.Package, missed []string) {
	aps, err := pacman.ReadAllAUR(pkgnames)
	if err != nil {
		if nf, ok := err.(*pacman.NotFoundError); ok {
			if Conf.Debug {
				fmt.Fprintf(os.Stderr, "warning: %s.\n", err)
			}
			missed = nf.Names
		} else {
			fmt.Fprintf(os.Stderr, "error: %s.\n", err)
			return nil, nil
		}
	}

	aur = make(map[string]*pacman.Package)
	for _, ap := range aps {
		aur[ap.Name] = ap.Package()
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

// dieOnError prints error to stderr and dies if err != nil.
func dieOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		os.Exit(1)
	}
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
