// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "github.com/goulash/pr"

// List displays all the packages available for the database.
// Note that they don't need to be registered with the database.
func List(dbdir string) {
	pkgs := GetAllPackages(dbdir)
	var pkgnames []string
	for _, p := range pkgs {
		pkgnames = append(pkgnames, p.Name)
	}
	pkgnames = uniq(pkgnames)
	pr.PrintAutoGrid(pkgnames)
}

func Add() {

}

func Remove() {

}

func Update() {

}

func Sync() {

}
