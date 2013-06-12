package actions

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/cassava/little/pr"
)

// List displays all the packages available for the database.
// Note that they don't need to be registered with the database.
func List() {
	var list []string
	// Get a list of all the files that match the pattern

	// Extract the package names from the files

	// Sort the names into a list

	// Print the list as columns

	// If GetSize fails, width is -1, and Printc prints single column.
	width, _, _ := terminal.GetSize()
	pr.Printc(list, width)
}
