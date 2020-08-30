// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repo

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cassava/repoctl/internal/term"
	"github.com/cassava/repoctl/pacman/aur"
	"github.com/cassava/repoctl/pacman/graph"
	"github.com/goulash/archive"
	"github.com/goulash/errs"
	"github.com/goulash/osutil"
)

// DependencyGraph returns a dependency graph of the given package names.
func (r *Repo) DependencyGraph(h errs.Handler, pkgnames ...string) (*graph.Graph, error) {
	errs.Init(&h)

	aurpkgs, err := r.ReadAUR(h, pkgnames...)
	if err != nil {
		return nil, fmt.Errorf("cannot read AUR: %w", err)
	}

	// Get dependencies
	term.Debugf("Creating dependency graph ...\n")
	f, err := graph.NewFactory()
	if err != nil {
		return nil, fmt.Errorf("cannot create dependency graph: %w", err)
	}

	f.SetSkipInstalled(true)
	f.SetTruncate(true)
	return f.NewGraph(uniqueBases(aurpkgs))
}

// Download downloads and extracts the given package tarballs.
func (r *Repo) Download(h errs.Handler, destdir string, extract bool, clobber bool, pkgnames ...string) error {
	errs.Init(&h)
	if len(pkgnames) == 0 {
		return nil
	}

	// If a package cannot be found, we want to report it.
	aurpkgs, err := r.ReadAUR(h, pkgnames...)
	if err != nil {
		err = h(err)
		if err != nil {
			return err
		}
	}
	return r.DownloadPackages(h, uniqueBases(aurpkgs), destdir, extract, clobber)
}

// DownloadPackages downloads the given AUR packages, printing messages for each one.
func (r *Repo) DownloadPackages(h errs.Handler, pkgs aur.Packages, destdir string, extract bool, clobber bool) error {
	errs.Init(&h)
	for _, p := range pkgs {
		term.Printf("Downloading: %s\n", p.Name)
		download := DownloadTarballAUR
		if extract {
			download = DownloadExtractAUR
		}
		err := h(download(p, destdir, clobber))
		if err != nil {
			return err
		}
	}
	return nil
}

// DownloadExtractAUR is a helper for Download and DownloadUpgrades.
func DownloadExtractAUR(ap *aur.Package, destdir string, clobber bool) error {
	var err error
	if destdir == "" {
		destdir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// Make sure we don't clobber anything.
	if !clobber {
		ex, err := osutil.DirExists(ap.Name)
		if err != nil {
			return err
		}
		if ex {
			return ErrPkgDirExists
		}
	}

	response, err := http.Get(ap.DownloadURL())
	if err != nil {
		return err
	}
	defer response.Body.Close()

	gr, err := gzip.NewReader(response.Body)
	if err != nil {
		return err
	}

	return archive.ExtractTar(gr, destdir)
}

// DownloadTarballAUR downloads the given package from AUR.
func DownloadTarballAUR(ap *aur.Package, destdir string, clobber bool) error {
	var err error
	if destdir == "" {
		destdir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	url := ap.DownloadURL()
	of := ap.Name + ".tar.gz"

	// Make sure we don't clobber anything.
	if !clobber {
		ex, err := osutil.FileExists(of)
		if err != nil {
			return err
		}
		if ex {
			return ErrPkgFileExists
		}
	}

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	file, err := os.Create(of)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		// Should I delete?
		return err
	}
	return nil
}

// uniqueBases returns a subset of the given aurpkgs where the package bases
// are the same.
func uniqueBases(aurpkgs aur.Packages) aur.Packages {
	bases := make(aur.Packages, 0, len(aurpkgs))
	mp := make(map[string]bool)
	for _, p := range aurpkgs {
		if mp[p.PackageBase] {
			continue
		}
		mp[p.PackageBase] = true
		bases = append(bases, p)
	}
	return bases
}

func upgradesToPackages(us Upgrades) aur.Packages {
	pkgs := make(aur.Packages, len(us))
	for i, p := range us {
		pkgs[i] = p.New
	}
	return pkgs
}
