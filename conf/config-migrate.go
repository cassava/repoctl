// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Prior to version 0.14, repoctl used a simple configuration file
// at $HOME/.repoctl.conf.
//
// As of 0.14, this configuration file is no longer sourced, and indeed,
// the configuration file syntax has also changed. repoctl can create
// an example configuration file with the new command, but an upgrading
// user might not know this.
//
// Here we check whether there is only this old configuration file,
// and we tell the user how to deal with it.

package conf

import (
	"os"
	"path"

	"github.com/goulash/osutil"
)

// oldConfigPath returns the original path to the repoctl configuration file.
func oldConfigPath() string {
	home := os.Getenv("HOME")

	// If home is not set, there is nothing we can do.
	if home == "" {
		return ""
	}

	return path.Join(home, ".repoctl.conf")
}

func oldConfigExists() bool {
	// If we can't access the file for some reason, it might as well not exist.
	ex, _ := osutil.FileExists(oldConfigPath())
	if ex {
		return true
	}
	return false
}
