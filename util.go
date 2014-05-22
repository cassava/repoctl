// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

// uniq returns a slice with all the duplicate elements in input removed;
// input must be sorted such that same elements are next to each other.
//
// TODO: This is a shoddy algorithm.
func uniq(input []string) []string {
	n := len(input)
	if n == 0 {
		return []string{}
	}

	output := make([]string, 1, n)
	output[0] = input[0]
	for _, v := range input[1:] {
		if output[len(output)-1] != v {
			output = append(output, v)
		}
	}
	return output
}
