// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/goulash/pacman"
	"github.com/goulash/pr"
	"github.com/goulash/util"
)

// List displays all the packages available for the database.
// Note that they don't need to be registered with the database.
func List(c *Config) error {
	// Get packages in repo directory and database and find out duplicates.
	// Depinding on the configuration, some of this might be unecessary work,
	// so we have to declare the variabes here.
	var (
		db     map[string]*pacman.Package
		missed []*pacman.Package
		dups   map[string]int
	)
	pkgs, old := pacman.SplitOld(getRepoPkgs(c))
	if c.Pending {
		db, missed = getDatabasePkgs(c)
	}
	if c.Duplicates {
		dups = make(map[string]int)
		for _, p := range old {
			dups[p.Name]++
		}
	}

	// Create a list.
	var pkgnames []string
	if c.Pending {
		for _, p := range missed {
			pkgnames = append(pkgnames, fmt.Sprintf("-%s-", p.Name))
		}
	}
	for _, p := range pkgs {
		buf := bytes.NewBufferString(p.Name)

		if c.Pending {
			dbp, ok := db[p.Name]
			if !ok || dbp.OlderThan(p) {
				buf.WriteRune('*')
			}
		}
		if c.Versioned {
			buf.WriteRune(' ')
			buf.WriteString(p.Version)
		}
		if c.Synchronize {
			aurp, err := pacman.ReadAUR(p.Name)
			if err != nil {
				c.inform(fmt.Sprintf("searching for %s: %s", p.Name, err))
				buf.WriteString(" <?>")
			} else if aurp.NewerThan(p) {
				if c.Versioned {
					buf.WriteString(" -> ")
					buf.WriteString(aurp.Version)
				}
				buf.WriteString(" <!>")
			}
		}
		if c.Duplicates && dups[p.Name] > 0 {
			buf.WriteString(fmt.Sprintf(" (%v)", dups[p.Name]))
		}

		pkgnames = append(pkgnames, buf.String())
	}

	// Print packages to stdout
	printList(c, pkgnames)
	return nil
}

func Status(c *Config) error {
	pkgs, old := pacman.SplitOld(getRepoPkgs(c))
	db, missed := getDatabasePkgs(c)
	name := c.Database[:strings.IndexByte(c.Database, '.')]
	fmt.Printf("On repo %s\n", name)

	// We assume that there is nothing to do, and if there is,
	// then this is set to false.
	var nothing = true

	if len(missed) > 0 {
		filenames := make([]string, len(missed))
		for i, p := range missed {
			filenames[i] = p.Name
		}
		printStatus("Database entries to be removed:", filenames)
		nothing = false
	}

	if len(old) > 0 {
		filenames := make([]string, len(old))
		for i, p := range old {
			filenames[i] = p.Filename
		}
		printStatus("Obsolete packages to be removed:", filenames)
		nothing = false
	}

	{
		var pending []string
		for _, p := range pkgs {
			dbp, ok := db[p.Name]
			if !ok || dbp.OlderThan(p) {
				pending = append(pending, p.Name)
			}
		}

		if len(pending) > 0 {
			printStatus("Packages pending database addition:", pending)
			nothing = false
		}
	}

	{
		var updates []string
		for _, p := range pkgs {
			aurp, err := pacman.ReadAUR(p.Name)
			if err != nil {
				c.inform(fmt.Sprintf("error searching for %s: %s", p.Name, err))
			} else if aurp.NewerThan(p) {
				updates = append(updates, p.Name)
			}
		}

		if len(updates) > 0 {
			printStatus("Outdated packages:", updates)
			nothing = false
		}
	}

	if nothing {
		fmt.Println("everything up-to-date.")
	}
	return nil
}

func printStatus(heading string, items []string) {
	sort.Strings(items)
	fmt.Println()
	fmt.Println(heading)
	for _, i := range items {
		fmt.Println(" ", i)
	}
}

func getRepoPkgs(c *Config) []*pacman.Package {
	ch := make(chan error)
	go handleErrors("warning: %s\n", ch)
	dirPkgs := pacman.ReadDir(c.RepoPath, ch)
	close(ch)
	return dirPkgs
}

// handleErrors is meant to be launched as a separate goroutine to handle
// errors coming from ReadDir and the likes.
func handleErrors(format string, ch <-chan error) {
	for err := range ch {
		fmt.Fprintf(os.Stderr, format, err)
	}
}

func getDatabasePkgs(c *Config) (db map[string]*pacman.Package, missed []*pacman.Package) {
	db = make(map[string]*pacman.Package)
	pkgs, _ := pacman.ReadDatabase(path.Join(c.RepoPath, c.Database))
	for _, p := range pkgs {
		if ex, _ := util.FileExists(p.Filename); !ex {
			missed = append(missed, p)
			continue
		}
		db[p.Name] = p
	}
	return db, missed
}

func printList(c *Config, items []string) {
	sort.Strings(items)
	if c.OnePerLine {
		for _, i := range items {
			fmt.Println(i)
		}
	} else {
		pr.PrintFlex(items)
	}
}
