// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"compress/gzip"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/goulash/osutil"
	"github.com/goulash/pacman"
	"github.com/juju/utils/tar"
)

func (r *Repo) Copy(h ErrHandler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, osutil.CopyFileLazy, "copying")
}

func (r *Repo) Move(h ErrHandler, pkgfiles ...string) error {
	return r.add(h, pkgfiles, osutil.MoveFileLazy, "moving")
}

func (r *Repo) add(h ErrHandler, pkgfiles []string, ar func(string, string) error, lbl string) error {
	AssertHandler(&h)
	if len(pkgfiles) == 0 {
		r.debugf("repoctl.(Repo).add: pkgfiles empty.\n")
		return nil
	}

	added := make([]string, 0, len(pkgfiles))
	for _, src := range pkgfiles {
		dst := path.Join(r.Directory, path.Base(src))
		r.printf("%s and adding to repository: %s\n", lbl, dst)
		err := ar(src, dst)
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			continue
		}
		added = append(added, dst)
	}

	err := r.DatabaseAdd(added...)
	if err != nil {
		return err
	}

	pkgs, err := r.FindSimilar(h, added...)
	return r.Dispatch(h, pkgs.Map(pacman.PkgFilename)...)
}

func (r *Repo) Remove(h ErrHandler, pkgnames ...string) error {
	AssertHandler(&h)
	if len(pkgnames) == 0 {
		r.debugf("repoctl.(Repo).Remove: pkgnames empty.\n")
		return nil
	}

	pkgs, err := pacman.ReadMatchingNames(r.Directory, pkgnames, h)
	if err != nil {
		return err
	}
	err = h(r.DatabaseRemove(pkgs.Map(pacman.PkgName)...))
	if err != nil {
		return err
	}
	return r.Dispatch(h, pkgs.Map(pacman.PkgFilename)...)
}

func (r *Repo) Dispatch(h ErrHandler, pkgfiles ...string) error {
	AssertHandler(&h)
	if len(pkgfiles) == 0 {
		r.debugf("repoctl.(Repo).Dispatch: pkgfiles empty.\n")
		return nil
	}

	if r.Backup {
		return r.backup(h, pkgfiles)
	}
	return r.unlink(h, pkgfiles)
}

func (r *Repo) backup(h ErrHandler, pkgfiles []string) error {
	for _, f := range pkgfiles {
		src := path.Base(f)
		r.printf("backing up: %s\n", src)
		dst := path.Join(r.Directory, r.BackupDir, src)
		err := osutil.MoveFileLazy(src, dst)
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func (r *Repo) unlink(h ErrHandler, pkgfiles []string) error {
	for _, f := range pkgfiles {
		src := path.Base(f)
		r.printf("deleting: %s\n", src)
		err := os.Remove(src)
		if err != nil {
			err = h(err)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

// Update adds the newest package found for the given name to the
// database and dispatches the obsolete packages.
//
// If pkgnames is empty, the entire repository is scanned.
func (r *Repo) Update(h ErrHandler, pkgnames ...string) error {
	AssertHandler(&h)

	var pkgs pacman.Packages
	var err error
	if len(pkgnames) == 0 {
		pkgs, err = r.FindNewest(h)
		if err != nil {
			return err
		}
	} else {
		pkgs, err = r.FindNames(h, pkgnames...)
		if err != nil {
			return err
		}
		pkgs = FilterNewest(pkgs)
	}

	files := pkgs.Map(pacman.PkgFilename)
	err = r.DatabaseAdd(files...)
	if err != nil {
		return err
	}
	pkgs, err = r.FindSimilar(h, files...)
	if err != nil {
		return err
	}
	return r.Dispatch(h, pkgs.Map(pacman.PkgFilename)...)
}

// Download downloads and extracts the given package tarballs.
func (r *Repo) Download(h ErrHandler, pkgnames ...string) error {
	AssertHandler(&h)
	if len(pkgnames) == 0 {
		r.debugf("repoctl.(Repo).Download: pkgnames empty.\n")
		return nil
	}

	aurpkgs, err := r.ReadAUR(h, pkgnames...)
	if err != nil {
		return err
	}
	for _, ap := range aurpkgs {
		err := h(r.DownloadAUR(ap, ""))
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
func (r *Repo) DownloadUpgrades(h ErrHandler, pkgnames ...string) error {
	AssertHandler(&h)

	upgrades, err := r.FindUpgrades(h, pkgnames...)
	if err != nil {
		return err
	}

	for _, u := range upgrades {
		err := h(r.DownloadAUR(u.New, ""))
		if err != nil {
			return err
		}
	}
	return nil
}

// DownloadAUR is actually a helper for Download and DownloadUpgrades.
// It is provided for your convenience.
func (r *Repo) DownloadAUR(ap *pacman.AURPackage, destdir string) error {
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
	ex, err := osutil.DirExists(ap.Name)
	if err != nil {
		return err
	}
	if ex {
		return ErrPkgDirExists
	}

	r.printf("downloading: %s\n", of)
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	gr, err := gzip.NewReader(response.Body)
	if err != nil {
		return err
	}
	return tar.UntarFiles(gr, destdir)
}
