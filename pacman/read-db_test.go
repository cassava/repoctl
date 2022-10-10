// Copyright (c) 2017, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/goulash/errs"
)

func TestReadLocalDatabase(z *testing.T) {
	if _, err := exec.LookPath("pacman"); err != nil {
		z.Skipf("pacman required for test, but not available: %s", err)
	}

	pkgs, err := ReadLocalDatabase(errs.Print(os.Stderr))
	if err != nil {
		z.Errorf("unexpected error: %s", err)
	}

	pkgmap := make(map[string]*Package)
	for _, p := range pkgs {
		pkgmap[p.Name] = p
	}

	bs, err := exec.Command("pacman", "-Q").Output()
	if err != nil {
		z.Fatalf("unexpected error: %s", err)
	}

	npkgs := 0
	scanner := bufio.NewScanner(bytes.NewReader(bs))
	for scanner.Scan() {
		npkgs++
		line := strings.Fields(scanner.Text())
		if len(line) != 2 {
			z.Fatalf("field size of %s is too large", line)
		}

		p, ok := pkgmap[line[0]]
		if !ok {
			z.Errorf("package %q not found in database", line[0])
			continue
		}
		if p.Version != line[1] {
			z.Errorf("package %q version mismatch: expected %s, got %s", line[0], line[1], p.Version)
		}
	}

	if npkgs != len(pkgs) {
		z.Errorf("database size mismatch: expected %d, got %d", npkgs, len(pkgs))
	}
}

func TestReadAllSyncDatabases(z *testing.T) {
	if _, err := exec.LookPath("pacman"); err != nil {
		z.Skipf("pacman required for test, but not available: %s", err)
	}

	pkgs, err := ReadAllSyncDatabases()
	if err != nil {
		z.Errorf("unexpected error: %s", err)
	}

	pkgmap := make(map[string]*Package)
	for _, p := range pkgs {
		pkgmap[p.Name] = p
	}

	bs, err := exec.Command("pacman", "-Ssq").Output()
	if err != nil {
		z.Fatalf("unexpected error: %s", err)
	}

	npkgs := 0
	scanner := bufio.NewScanner(bytes.NewReader(bs))
	for scanner.Scan() {
		npkgs++
		line := scanner.Text()
		_, ok := pkgmap[line]
		if !ok {
			z.Errorf("package %q not found in database", line[0])
			continue
		}
	}

	if npkgs != len(pkgs) {
		z.Errorf("database size mismatch: expected %d, got %d", npkgs, len(pkgs))
	}
}
