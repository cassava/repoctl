// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/cassava/repoctl/conf"
	"github.com/goulash/errs"
	"github.com/goulash/pacman/pkgutil"
)

// main loads the configuration and executes the primary command.
func main() {
	conf := conf.Default()
	err := conf.MergeAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s.\n", err)
		os.Exit(1)
	}
	repo := conf.Repo()
	// TODO: Does this work? Because we're still testing:
	errs.Default = errs.Print(os.Stderr)
	repo.Debug = os.Stderr
	repo.Info = os.Stderr
	repo.Error = os.Stderr

	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println("Usage: repols <fs|db|files|database>")
		os.Exit(1)
	}

	switch args[0] {
	case "fs":
		names, err := repo.ListDirectory(nil, nil)
		dieOnError(err)
		printList(names)
	case "db":
		names, err := repo.ListDatabase(nil)
		dieOnError(err)
		printList(names)
	case "files", "filesystem":
		filenames, err := repo.ListDirectory(nil, pkgutil.PkgFilename)
		dieOnError(err)
		printList(filenames)
	case "database":
		filenames, err := repo.ListDatabase(pkgutil.PkgFilename)
		dieOnError(err)
		printList(filenames)
	default:
		fmt.Println("Usage: repols <fs|db|files|database>")
		os.Exit(1)
	}
}

func printList(ls []string) {
	for _, i := range ls {
		fmt.Println(i)
	}
}

func dieOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
