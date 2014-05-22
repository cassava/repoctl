// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "testing"

func TestUniq(t *testing.T) {
	tests := [][2][]string{
		{
			{"a", "a", "b", "b", "d", "e", "e", "f", "z", "z"},
			{"a", "b", "d", "e", "f", "z"},
		},
		{
			{"a", "a", "a", "a", "a", "a"},
			{"a"},
		},
		{
			{"a", "b", "b", "a"},
			{"a", "b", "a"},
		},
		{
			{"b", "c", "d"},
			{"b", "c", "d"},
		},
	}

	for _, tst := range tests {
		r := uniq(tst[0])
		if eq(r, tst[1]) {
			t.Errorf("uniq(%v) = %v, want %v", tst[0], r, tst[1])
		}
	}
}

// eq tests slices a and b for equality. Both the number of elements,
// the order, and the content must be identical.
func eq(a, b []string) bool {
	n := len(a)
	if n != len(b) {
		return false
	}

	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
