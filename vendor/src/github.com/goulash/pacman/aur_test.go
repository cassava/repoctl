// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import "testing"

var (
	aurExists    = []string{"repoctl", "fairsplit", "moped"}
	aurNotExists = []string{"repoctl-34534", "arstaorsf", "911222234"}
	aurInvalid   = []string{"-", "_", "*", "-1q"}
)

func TestReadAUR1(z *testing.T) {
	for _, n := range aurExists {
		i, err := ReadAUR(n)
		if err != nil {
			z.Errorf("unexpected error: %s", err)
		}
		if i == nil {
			z.Errorf("expected i = non-nil, got nil")
		} else if i.Name != n {
			z.Errorf("wrong name returned")
		}
	}
}

func TestReadAUR2(z *testing.T) {
	for _, n := range aurNotExists {
		i, err := ReadAUR(n)
		if i != nil {
			z.Errorf("expecting i to be nil")
		}
		if err == nil {
			z.Errorf("expecting error, got nil")
		} else if nf, ok := err.(*NotFoundError); ok {
			if len(nf.Names) != 1 {
				z.Errorf("wrong number of names returned")
			} else if nf.Names[0] != n {
				z.Errorf("wrong name returned")
			}
		} else {
			z.Errorf("unexpected error: %s", err)
		}
	}
}

func TestReadAllAUR1(z *testing.T) {
	is, err := ReadAllAUR(aurExists)
	if err != nil {
		z.Errorf("unexpected error: %s", err)
	}
	if is == nil {
		z.Errorf("expected i = non-nil, got nil")
	}
}

func TestReadAllAUR2(z *testing.T) {
	is, err := ReadAllAUR(aurNotExists)
	if len(is) != 0 {
		z.Errorf("expecting is to have zero elements")
	}
	if err == nil {
		z.Errorf("expecting error, got nil")
	} else if nf, ok := err.(*NotFoundError); ok {
		if len(nf.Names) != len(aurNotExists) {
			z.Errorf("wrong number of names returned")
		} else {
			for i, n := range aurNotExists {
				if nf.Names[i] != n {
					z.Errorf("wrong name returned")
				}
			}
		}
	} else {
		z.Errorf("unexpected error: %s", err)
	}
}

func TestDownloadURL(z *testing.T) {
	i, err := ReadAUR("repoctl")
	if err != nil {
		z.Errorf("unexpected error: %s", err)
		z.FailNow()
	}
	if i.DownloadURL() != "https://aur4.archlinux.org/cgit/aur.git/snapshot/repoctl.tar.gz" {
		z.Errorf("download url incorrect: %s", i.DownloadURL())
	}
}
