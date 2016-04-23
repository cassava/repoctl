// Copyright (c) 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package pr

import (
	"math"
)

// grid represents a list of values as a grid.
//
// This is an experimental way of dealing with the need to map a list of
// items to a grid.
//
// Given a certain number of columns and rows, as well as elements in total,
// grid maps them to a 2D grid. Member functions such as IterRows can then be
// used to traverse the elements. In particular, grid assumes that the n items
// are layed out in the following way:
//
//	0	3	6	9
//  1	4	7	10
//  2	5	8	11
//
// That was for example, a grid made of 12 elements, with 4 columns and 3 rows.
// Only two of these three details are needed for initialization, the third is
// inferred.
//
// Note: the layout assumption that grid makes will probably be removed in the
// future, as the only functions that make use of these assumptions are the
// iterators.
type grid struct {
	n    int
	cols int
	rows int
}

// gridIndex represents a single point in the grid.  When iterating a grid,
// a gridIndex is returned.
type gridIndex struct {
	// The index of the original sequential list.
	Idx int
	// Col goes from 0 to the number of columns that the grid has.
	Col int
	// Row goes from 0 to the number of rows that the grid has.
	Row int
	// Ok describes whether the Idx value is within bounds.
	Ok bool
}

func newGrid(cols, rows int) grid {
	return grid{cols * rows, cols, rows}
}

func newGridFromCols(n, cols int) grid {
	rows := int(math.Ceil(float64(n) / float64(cols)))
	return grid{n, cols, rows}
}

func newGridFromRows(n, rows int) grid {
	cols := int(math.Ceil(float64(n) / float64(rows)))
	return grid{n, cols, rows}
}

func (g grid) N() int {
	return g.n
}

func (g grid) Cols() int {
	return g.cols
}

func (g grid) Rows() int {
	return g.rows
}

// IterRows gets the points from the grid row by row. Given the example grid at
// the definition of grid, this would return the elements thus in the following
// order:
//
//	0, 3, 6, 9, 1, 4, 7, 10, 2, 5, 8, 11
//
// It therefore assumes that the elements flow top-to-bottom and then
// left-to-right.
func (g grid) IterRows() <-chan gridIndex {
	ch := make(chan gridIndex)
	go func() {
		pts := g.cols * g.rows
		for i := 0; i < pts; i++ {
			col := i % g.cols
			row := i / g.cols
			j := col*g.rows + row
			ok := j < g.n

			ch <- gridIndex{j, col, row, ok}
		}
		close(ch)
	}()
	return ch
}
