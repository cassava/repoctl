package actions

import (
	"fmt"
	"sort"

	//"github.com/goulash/pr"
)

// List displays all the packages available for the database.
// Note that they don't need to be registered with the database.
func List() {
	files, err := readPackageFiles("/home/benmorgan/")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Extract the package names from the files.
	pkgs := readPackageNames(files)

	// Sort the names into a list.
	sort.StringSlice(pkgs).Sort()
	list := uniq(pkgs)

	// Print the list as a grid, like ls always does it.
	//pr.PrintAutoGrid(pkgs)
	fmt.Println(list)
}
