// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package pacman

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/goulash/osutil"
)

// Read reads the package information from a pacman package
// and returns it in the Package datatype.
func Read(filename string) (*Package, error) {
	bs, err := osutil.ReadFileFromArchive(filename, ".PKGINFO")
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(bs)
	info, err := readFilePkgInfo(r)
	if err != nil {
		return nil, err
	}

	info.Filename = filename
	info.Origin = FileOrigin
	return info, nil
}

// readFilePkgInfo reads the package information from a pacman package.
//
// We don't do any specific controlling for you, so you should use
// HasPackageFormat on a path string before using this function on it.
// Even if you don't, nothing bad should happen, but just in case.
func readFilePkgInfo(r io.Reader) (*Package, error) {
	var (
		info  Package
		err   error
		epoch int
	)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix("#", line) {
			continue
		}

		kv := strings.Split(line, " = ")
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "pkgname":
			info.Name = kv[1]
		case "pkgver":
			info.Version = kv[1]
		case "pkgdesc":
			info.Description = kv[1]
		case "pkgbase":
			info.Base = kv[1]
		case "epoch":
			epoch, err = strconv.Atoi(kv[1])
			if err != nil {
				return nil, fmt.Errorf("cannot parse epoch value '%s'", kv[1])
			}
		case "url":
			info.URL = kv[1]
		case "builddate":
			n, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse build time '%s'", kv[1])
			}
			info.BuildDate = time.Unix(n, 0)
		case "packager":
			info.Packager = kv[1]
		case "size":
			info.Size, err = strconv.ParseUint(kv[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse size value '%s'", kv[1])
			}
		case "arch":
			info.Arch = kv[1]
		case "license":
			info.License = kv[1]
		case "depend":
			info.Depends = append(info.Depends, kv[1])
		case "optdepend":
			info.OptionalDepends = append(info.OptionalDepends, kv[1])
		case "makedepend":
			info.MakeDepends = append(info.MakeDepends, kv[1])
		case "checkdepend":
			info.CheckDepends = append(info.CheckDepends, kv[1])
		case "makepkgopt":
			info.MakeOptions = append(info.MakeOptions, kv[1])
		case "backup":
			info.Backups = append(info.Backups, kv[1])
		case "replaces":
			info.Replaces = append(info.Replaces, kv[1])
		case "provides":
			info.Provides = append(info.Provides, kv[1])
		case "conflict":
			info.Conflicts = append(info.Conflicts, kv[1])
		case "group":
			info.Groups = append(info.Groups, kv[1])
		default:
			return nil, fmt.Errorf("unknown field '%s' in .PKGINFO", kv[0])
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	// The Version field must include the epoch value. If it already has one,
	// we compare and take the maximum value.
	if epoch > 0 {
		if i := strings.IndexByte(info.Version, ':'); i != -1 {
			e, err := strconv.Atoi(info.Version[:i])
			if err != nil {
				return nil, fmt.Errorf("unable to read epoch from version '%s'", info.Version)
			}
			epoch = max(epoch, e)
			info.Version = info.Version[i+1:]
		}
		info.Version = fmt.Sprintf("%d:%s", epoch, info.Version)
	}

	return &info, nil
}
