// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package osutil

import (
	"fmt"
)

type FileTypeError struct {
	Filepath string
}

func (e FileTypeError) Error() string {
	return fmt.Sprintf("unexpected file type at %q")
}
