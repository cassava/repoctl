// Copyright (c) 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package pr

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// columnPadding is the number of normal spaces that are between two columns.
var columnPadding = 2

// SetColumnPadding sets the number of spaces that are between two columns.
// The default value is two, which comes from the GNU utility "ls".
func SetColumnPadding(padding int) {
	if padding < 0 {
		columnPadding = 2 // default value
	} else {
		columnPadding = padding
	}
}

// FprintGrid tries to print the items in the given list in the given
// number of columns.
//
// If it is not possible to fit the items in that number of columns,
// then the items are printed in as many columns as possible.
func FprintGrid(w io.Writer, hspace, columns int, list []string) {
	writeGrid(w, hspace, columns, list)
}

// FprintFlex prints the items in the given list in as many columns as
// makes sense, given the horizontal space available.
//
// It will not print items in more columns than necessary: the minimum
// number of columns is used to attain the minimum row count.
func FprintFlex(w io.Writer, hspace int, list []string) {
	writeGrid(w, hspace, 0, list)
}

// writeGrid is the basis for all the other column or grid functions.
func writeGrid(w io.Writer, hspace, maxcols int, list []string) {
	n := len(list)
	rc := runes(list)
	span := columns(rc, columnPadding, maxcols, hspace)
	cols := len(span)

	if cols <= 1 {
		for _, s := range list {
			fmt.Println(s)
		}
		return
	}

	g := newGridFromCols(n, cols)
	for i := range g.IterRows() {
		if i.Ok {
			fmt.Print(list[i.Idx])
			fmt.Print(strings.Repeat(" ", span[i.Col]-rc[i.Idx]))
		}

		// At the last column, print a newline
		if i.Col == cols-1 {
			fmt.Print("\n")
		}
	}
}

// columns returns the amount of columns that can fit in space, taken the list
// and the padding into consideration.
//
// If the return value is nil, then only one column is supported (if at all),
// otherwise it contains the length of each column, including the padding.
func columns(list []int, padding, maxcols, hspace int) []int {
	if hspace <= 0 {
		return nil
	}

	n := len(list)
	rows := n
	cols := []int(nil)
	max := n
	if maxcols > 0 {
		max = maxcols
	}

trial:
	for c := 2; c <= max; c++ {
		g := newGridFromCols(n, c)

		// Continue if we don't reduce rows with this many columns.
		if g.Rows() >= rows {
			continue
		}

		// Get the maximum widths of the individual columns.
		span := make([]int, c)
		for i := range g.IterRows() {
			if i.Ok && list[i.Idx] > span[i.Col] {
				span[i.Col] = list[i.Idx]
			}
		}

		// Have we reached the limit yet?
		padspace := padding * (len(cols) - 1)
		// FIXME: the -padding should not be necessary!
		if sum(span)+padspace > hspace-padding {
			break trial
		}

		// Update our data.
		rows = g.Rows()
		cols = span
	}

	// Add the padding.
	for i := 0; i < len(cols)-1; i++ {
		cols[i] += padding
	}
	return cols
}

// runes returns the rune counts of each of the strings in the list.
func runes(list []string) []int {
	retval := make([]int, len(list))
	for i, s := range list {
		retval[i] = utf8.RuneCountInString(s)
	}
	return retval
}

// avg computes the average over the uints in the list.
func avg(list []int) int {
	return sum(list) / len(list)
}

// sum computes the total over the uints in the list.
func sum(list []int) int {
	var total int
	for _, i := range list {
		total += i
	}
	return total
}
