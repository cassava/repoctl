// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
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
