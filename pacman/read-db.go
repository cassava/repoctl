// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/goulash/archive"
	"github.com/goulash/errs"
	"github.com/goulash/osutil"
)

// PacmanConfPath contains the path to the pacman configuration.
// This is provided for those with special needs, such as me ;)
var PacmanConfPath = "/etc/pacman.conf"

// PacmanLocalDatabasePath contains the path to the local pacman library.
// This is provided for those with special needs to modify at their own risk.
var PacmanLocalDatabasePath = "/var/lib/pacman/local"

// PacmanSyncDatabaseFormat is the format that fmt.Sprinf needs to interpolate
// a repository name into a package database path.
// This is provided for those with special needs to modify at their own risk.
var PacmanSyncDatabaseFormat = "/var/lib/pacman/sync/%s.db"

// IsDatabaseLocked returns whether the database given at the path
// is currently locked for writing or not.
func IsDatabaseLocked(dbpath string) bool {
	lockpath := dbpath + ".lck"
	ex, _ := osutil.FileExists(lockpath)
	return ex
}

// ReadDatabase reads all the packages from a database file.
func ReadDatabase(dbpath string) (Packages, error) {
	debugf("Read database %s\n", dbpath)

	var dr io.ReadCloser
	var err error

	if ex, err := osutil.FileExists(dbpath); !ex {
		if err != nil {
			return nil, fmt.Errorf("read database %s: %w", dbpath, err)
		}
		return nil, fmt.Errorf("read database %s: no such file", dbpath)
	}

	dr, err = archive.NewDecompressor(dbpath)
	if err != nil {
		return nil, fmt.Errorf("read database %s: %w", dbpath, err)
	}
	defer dr.Close()

	tr := tar.NewReader(dr)
	var pkgs Packages

	hdr, err := tr.Next()
	for hdr != nil {
		fi := hdr.FileInfo()
		if !fi.IsDir() {
			return nil, fmt.Errorf("read database %s: unexpected file '%s'", dbpath, hdr.Name)
		}

		pr := archive.DirReader(tr, &hdr)
		pkg, err := readTarredDatabasePkgInfo(pr, dbpath)
		if err != nil {
			if err == archive.EOA {
				break
			}
			return nil, fmt.Errorf("read database %s: %w", dbpath, err)
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func readTarredDatabasePkgInfo(r io.Reader, dbpath string) (*Package, error) {
	pkg, err := readDatabasePkgInfo(r)
	if err != nil {
		return nil, err
	}
	pkg.Origin = DatabaseOrigin
	if pkg.Filename != "" {
		pkg.Filename = path.Join(path.Dir(dbpath), pkg.Filename)
	}
	return pkg, nil
}

// ReadSyncDatabase reads one of the package databases synced by pacman,
// such as "core", "extra", "community", and so on.
//
// It also reads "/etc/pacman.conf" to make sure that the repository is enabled.
func ReadSyncDatabase(name string) (Packages, error) {
	ok, err := IsRepositoryEnabled(name)
	if err != nil {
		return nil, fmt.Errorf("cannot determine if repository is enabled: %s", err)
	}
	if !ok {
		return nil, fmt.Errorf("repository %q is not enabled in %s", name, PacmanConfPath)
	}

	return ReadDatabase(fmt.Sprintf(PacmanSyncDatabaseFormat, name))
}

// ReadAllSyncDatabases reads all locally synced databases, using /etc/pacman.conf
// to determine which ones to read.
func ReadAllSyncDatabases() (Packages, error) {
	enabled, err := EnabledRepositories()
	if err != nil {
		return nil, err
	}

	// As of October 2017, the main repositories have in total less than 10,000 entries.
	// I expect this value to rise over time, so reserving space for 15,000 should be
	// enough for the next few years, hopefully.
	list := make(Packages, 0, 15000)
	for _, name := range enabled {
		pkgs, err := ReadDatabase(fmt.Sprintf(PacmanSyncDatabaseFormat, name))
		if err != nil {
			return nil, err
		}
		list = append(list, pkgs...)
	}
	return list, nil
}

// IsRepositoryEnabled returns whether the repository named is enabled.
func IsRepositoryEnabled(name string) (bool, error) {
	enabled, err := EnabledRepositories()
	if err != nil {
		return false, err
	}

	for _, repo := range enabled {
		if name == repo {
			return true, nil
		}
	}
	return false, nil
}

// EnabledRepositories returns a list of repository names that are enabled
// in the `/etc/pacman.conf` system configuration file.
func EnabledRepositories() ([]string, error) {
	f, err := os.Open(PacmanConfPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %s", PacmanConfPath, err)
	}

	repos := make([]string, 0, 5)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
			continue
		}

		name := line[1 : len(line)-1]
		if name == "options" {
			continue
		}

		repos = append(repos, name)
	}
	return repos, nil
}

// ReadLocalDatabase reads the database of locally installed packages.
//
// Note: Even if an error occurs, all successfully read packages will
// be returned.
//
// Note: Errors that occur are passed to the error handler eh, and it is
// highly recommended that eh always return nil; else reading the local
// database will be aborted during reading.
func ReadLocalDatabase(eh errs.Handler) (Packages, error) {
	var pkgs Packages
	err := filepath.Walk(PacmanLocalDatabasePath, func(p string, fi os.FileInfo, err error) error {
		if fi.Name() != "desc" || fi.IsDir() {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			return eh(fmt.Errorf("%s: %s", p, err))
		}
		defer f.Close()
		pkg, err := readDatabasePkgInfo(f)
		if err != nil {
			return eh(fmt.Errorf("%s: %s", p, err))
		}

		// Everything ok
		pkg.Origin = LocalOrigin
		pkg.Filename = path.Dir(p)
		pkgs = append(pkgs, pkg)
		return nil
	})
	return pkgs, err
}

func readDatabasePkgInfo(r io.Reader) (*Package, error) {
	var err error

	info := Package{}
	del := "%"
	state := ""
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, del) && strings.HasSuffix(line, del) {
			state = strings.ToLower(strings.Trim(line, del))
			continue
		}

		switch state {
		case "filename":
			info.Filename = line
		case "name":
			info.Name = line
		case "version":
			info.Version = line
		case "desc":
			info.Description = line
		case "base":
			info.Base = line
		case "url":
			info.URL = line
		case "builddate":
			n, err := strconv.ParseInt(line, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse build time '%s'\n", line)
			}
			info.BuildDate = time.Unix(n, 0)
		case "packager":
			info.Packager = line
		case "csize":
			info.Size, err = strconv.ParseUint(line, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse size value '%s'\n", line)
			}
		case "arch":
			info.Arch = line
		case "license":
			info.License = line
		case "depends":
			info.Depends = append(info.Depends, line)
		case "optdepends":
			info.OptionalDepends = append(info.OptionalDepends, line)
		case "makedepends":
			info.MakeDepends = append(info.MakeDepends, line)
		case "checkdepends":
			info.CheckDepends = append(info.CheckDepends, line)
		case "backup":
			info.Backups = append(info.Backups, line)
		case "replaces":
			info.Replaces = append(info.Replaces, line)
		case "provides":
			info.Provides = append(info.Provides, line)
		case "conflicts":
			info.Conflicts = append(info.Conflicts, line)
		case "groups":
			info.Groups = append(info.Groups, line)
		case "xdata":
			info.Xdata = append(info.Xdata, strings.Split(line, " ")...)
		case "isize", "md5sum", "pgpsig", "sha256sum":
			// We ignore these fields for now...
			continue
		case "installdate", "size", "validation", "reason":
			// These fields are in the local database desc file and we don't care about them
			continue
		default:
			return nil, fmt.Errorf("unknown field '%s' in database entry\n", state)
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return &info, nil
}
