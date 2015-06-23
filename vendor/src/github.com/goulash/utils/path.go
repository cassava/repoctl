// Copyright (c) 2013, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util

import (
	"path"
	"strings"
)

// FileExt returns the file type as identified by the (lowercase) extension.
// As such it is very limited, but for limited purposes, simple enough.
func FileExt(filepath string) string {
	return strings.ToLower(path.Ext(filepath))[1:]
}
