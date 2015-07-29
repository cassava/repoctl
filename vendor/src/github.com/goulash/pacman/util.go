// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

// SplitOld splits the input array into one containing the newest
// packages and another containing the outdated packages.
func SplitOld(pkgs []*Package) (updated []*Package, old []*Package) {
	var m = make(map[string]*Package)

	// Find out which packages are newest and put the others in the old array.
	for _, p := range pkgs {
		if cur, ok := m[p.Name]; ok {
			if cur.OlderThan(p) {
				old = append(old, cur)
			} else {
				old = append(old, p)
				continue
			}
		}
		m[p.Name] = p
	}

	// Add the newest packages to the updated array and return.
	updated = make([]*Package, 0, len(m))
	for _, v := range m {
		updated = append(updated, v)
	}

	return updated, old
}

// The following functions are only trying to be correct in the context that we
// are using them. They are used mostly (though not exclusively) in vercmp.go.

// isdigit returns true if c is a digit.
func isdigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// isalpha returns true if c is part of [a-z].
func isalpha(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// intcmp returns the comparison of two integers.
func intcmp(a, b int) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

// strcmp returns the comparison of two strings.
func strcmp(a, b string) int {
	m, n := len(a), len(b)
	z := min(m, n)
	for i := 0; i < z; i++ {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return intcmp(m, n)
}

// min returns the lesser of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the greater of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
