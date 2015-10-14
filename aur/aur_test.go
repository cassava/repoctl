// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package aur

import "testing"

var (
	exists    = []string{"repoctl", "fairsplit", "moped"}
	notExists = []string{"repoctl-34534", "arstaorsf", "911222234"}
	invalid   = []string{"-", "_", "*", "-1q"}
)

func TestRead1(z *testing.T) {
	for _, n := range exists {
		i, err := Read(n)
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

func TestRead2(z *testing.T) {
	for _, n := range notExists {
		i, err := Read(n)
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

func TestReadAll1(z *testing.T) {
	is, err := ReadAll(exists)
	if err != nil {
		z.Errorf("unexpected error: %s", err)
	}
	if is == nil {
		z.Errorf("expected i = non-nil, got nil")
	}
}

func TestReadAll2(z *testing.T) {
	is, err := ReadAll(notExists)
	if len(is) != 0 {
		z.Errorf("expecting is to have zero elements")
	}
	if err == nil {
		z.Errorf("expecting error, got nil")
	} else if nf, ok := err.(*NotFoundError); ok {
		if len(nf.Names) != len(notExists) {
			z.Errorf("wrong number of names returned")
		} else {
			for i, n := range notExists {
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
	i, err := Read("repoctl")
	if err != nil {
		z.Errorf("unexpected error: %s", err)
		z.FailNow()
	}
	if i.DownloadURL() != "https://aur.archlinux.org/cgit/aur.git/snapshot/repoctl.tar.gz" {
		z.Errorf("download url incorrect: %s", i.DownloadURL())
	}
}
