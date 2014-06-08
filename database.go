// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"archive/tar"
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/goulash/util"
)

// HasDatabaseFormat returns true if the filename matches a pacman package
// format that we can do anything with.
//
// Currently, only the following formats are supported:
//	.db.tar.gz
//
func HasDatabaseFormat(filename string) bool {
	return strings.HasSuffix(filename, ".db.tar.gz")
}

// ReadDatabase reads all the packages from a database file.
func ReadDatabase(dbpath string) ([]*Package, error) {
	dr, err := util.NewDecompressor(dbpath)
	if err != nil {
		return nil, err
	}
	defer dr.Close()

	tr := tar.NewReader(dr)
	var pkgs []*Package

	hdr, err := tr.Next()
	for hdr != nil {
		fi := hdr.FileInfo()
		if !fi.IsDir() {
			return nil, errors.New(fmt.Sprintf("unexpected file '%s'", hdr.Name))
		}

		pr := util.DirReader(tr, &hdr)
		pkg, err := readDatabasePkgInfo(pr, dbpath)
		if err != nil {
			if err == util.EOA {
				break
			}
			return nil, err
		}

		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func readDatabasePkgInfo(r io.Reader, dbpath string) (*Package, error) {
	var info Package
	var err error

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
			info.Filename = path.Join(path.Dir(dbpath), line)
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
				log.Printf("Warning: cannot parse build time '%s'\n", line)
			}
			info.BuildDate = time.Unix(n, 0)
		case "packager":
			info.Packager = line
		case "csize":
			info.Size, err = strconv.ParseUint(line, 10, 64)
			if err != nil {
				log.Printf("Warning: cannot parse size value '%s'\n", line)
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
		case "isize", "md5sum", "pgpsig", "sha256sum":
			// We ignore these fields for now...
			continue
		default:
			log.Printf("Warning: unknown field '%s' in database entry\n", state)
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return &info, nil
}
