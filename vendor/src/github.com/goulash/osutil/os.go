// Copyright (c) 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package osutil

import (
	"fmt"
	"os"
)

// Exists returns exists = true if the given file exists, regardless whether
// it is a file or a directory. Normally you will probably want to use the more
// specific versions: FileExists and DirectoryExists.
func Exists(path string) (ex bool, err error) {
	ex, _, err = exists(path)
	return ex, err
}

// FileExists returns ex = true if the file exists and is not
// a directory, and returns err != nil if any other error occured (such as
// permission denied).
func FileExists(path string) (ex bool, err error) {
	var stat os.FileInfo

	ex, stat, err = exists(path)
	if err != nil && stat.IsDir() {
		err = fmt.Errorf("%s exists but is a directory not a file", path)
	}
	return ex, err
}

// DirExists returns ex = true if the file exists and is
// a directory, and returns err != nil if any other error occured (such as
// permission denied).
func DirExists(path string) (ex bool, err error) {
	var stat os.FileInfo

	ex, stat, err = exists(path)
	if err != nil && !stat.IsDir() {
		err = fmt.Errorf("%s exists but is not a directory", path)
	}
	return ex, err
}

// exists does the hard work for Exists, FileExists, and DirExists,
// returning ex = true if the file given by path exists.
func exists(path string) (ex bool, stat os.FileInfo, err error) {
	stat, err = os.Stat(path)

	ex = true
	if err != nil {
		// ex = true if file exists
		ex = !os.IsNotExist(err)
		if !ex {
			err = nil
		}
	}
	return ex, stat, err
}
