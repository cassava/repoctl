// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import "github.com/goulash/pacman"

func ReadPackages(h ErrHandler, pkgfiles ...string) (pacman.Packages, error) {
	AssertHandler(&h)
	if len(pkgfiles) == 0 {
		return nil, nil
	}

	pkgs := make(pacman.Packages, 0, len(pkgfiles))
	for _, pf := range pkgfiles {
		p, err := pacman.ReadPackage(pf)
		if err != nil {
			err = h(err)
			if err != nil {
				return pkgs, err
			}
		} else {
			pkgs = append(pkgs, p)
		}
	}
	return pkgs, nil
}

func FilterNewest(pkgs pacman.Packages) pacman.Packages {
	m := make(map[string]*pacman.Package)
	for _, p := range pkgs {
		if p.NewerThan(m[p.Name]) {
			m[p.Name] = p
		}
	}

	out := make(pacman.Packages, 0, len(m))
	for _, p := range m {
		out = append(out, p)
	}
	return out
}
