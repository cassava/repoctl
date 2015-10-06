// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"errors"
	"fmt"
)

var (
	ErrRepoDirRelative = errors.New("repository directory path must be absolute")
	ErrRepoDirMissing  = errors.New("repository directory path does not exist")
	ErrRepoDirInvalid  = errors.New("repository directory path is invalid")
)

type NotExistsError struct {
	Filepath string
}

func (e NotExistsError) Error() string {
	return fmt.Sprintf("file %q does not exist", e.Filepath)
}

type InvalidFileError struct {
	Filepath string
	WantDir  bool
}

func (e InvalidFileError) Error() string {
	if e.WantDir {
		return fmt.Sprintf("expected directory at %q", e.Filepath)
	}
	return fmt.Sprintf("expected file at %q", e.Filepath)
}
