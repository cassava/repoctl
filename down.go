// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"compress/gzip"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
	"github.com/juju/utils/tar"
)

// Download downloads and extracts the given package tarballs.
func (r *Repo) Download(h ErrHandler, destdir string, extract bool, clobber bool, pkgnames ...string) error {
	AssertHandler(&h)
	if len(pkgnames) == 0 {
		r.debugf("repoctl.(Repo).Download: pkgnames empty.\n")
		return nil
	}

	// If a package cannot be found, we want to report it.
	aurpkgs, err := r.ReadAUR(h, pkgnames...)
	if err != nil {
		return err
	}
	aurpkgs = uniqueBases(aurpkgs)
	for _, ap := range aurpkgs {
		r.printf("downloading: %s\n", ap.Name)
		download := DownloadTarballAUR
		if extract {
			download = DownloadExtractAUR
		}
		err = h(download(ap, destdir, clobber))
		if err != nil {
			return err
		}
	}
	return nil
}

// DownloadUpgrades downloads all available upgrades for the given
// package names.
//
// If pkgnames is empty, all available upgrades are downloaded.
func (r *Repo) DownloadUpgrades(h ErrHandler, destdir string, extract bool, clobber bool, pkgnames ...string) error {
	AssertHandler(&h)

	upgrades, err := r.FindUpgrades(h, pkgnames...)
	if err != nil {
		return err
	}

	upgrades = uniqueUpgrades(upgrades)
	for _, u := range upgrades {
		r.printf("downloading: %s\n", u.Name())
		download := DownloadTarballAUR
		if extract {
			download = DownloadExtractAUR
		}
		err = h(download(u.New, destdir, clobber))
		if err != nil {
			return err
		}
	}
	return nil
}

// DownloadExtractAUR is a helper for Download and DownloadUpgrades.
func DownloadExtractAUR(ap *pacman.AURPackage, destdir string, clobber bool) error {
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

	// FIXME: This is a workaround for a bug somewhere in the chain of
	// command starting from UntarFiles which extracts the header in
	// the tar file.
	paxf := path.Join(destdir, "pax_global_header")
	if ex, _ := osutil.FileExists(paxf); !ex {
		defer os.Remove(paxf)
	}

	return tar.UntarFiles(gr, destdir)
}

func DownloadTarballAUR(ap *pacman.AURPackage, destdir string, clobber bool) error {
	var err error
	if destdir == "" {
		destdir, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	url := ap.DownloadURL()
	tokens := strings.Split(url, "/")
	of := tokens[len(tokens)-1]

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
func uniqueBases(aurpkgs pacman.AURPackages) pacman.AURPackages {
	bases := make(pacman.AURPackages, 0, len(aurpkgs))
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

// uniqueBases returns a subset of the given aurpkgs where the package bases
// are the same.
func uniqueUpgrades(us Upgrades) Upgrades {
	bases := make(Upgrades, 0, len(us))
	mp := make(map[string]bool)
	for _, u := range us {
		if mp[u.New.PackageBase] {
			continue
		}
		mp[u.New.PackageBase] = true
		bases = append(bases, u)
	}
	return bases
}
