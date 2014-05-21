package main

import "github.com/goulash/pr"

// List displays all the packages available for the database.
// Note that they don't need to be registered with the database.
func List() {
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
