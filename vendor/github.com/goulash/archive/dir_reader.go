// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package archive

import (
	"archive/tar"
	"errors"
	"io"
	"path"
)

// EOA signifies that the archive has been exhausted.
var EOA error = errors.New("end of archive")

type dirReader struct {
	tr   *tar.Reader
	hdr  **tar.Header
	base string
	err  error
}

// DirReader returns a specialized reader that reads all the files in
// a directory in a tar archive as if they were one file.
//
// In this sense, it is very similar to the reader returned by io.MultiReader.
//
// The reader advances the header of tar forward as long as it reads from files
// in a given directory; this is modified on the callers side, hence the need
// to pass a pointer to a pointer of tar.Header. When Read then returns io.EOF,
// the current header is on the next valid file. tr.Next() must not be called
// after reading from DirReader.
//
// DirReader differentiates between the end of directory and end of
// archive by returning io.EOF in the first and EOA in the latter case.
//
// BUG: at the moment it chokes if there is another directory in the directory
// given.
func DirReader(tr *tar.Reader, dirHeader **tar.Header) io.Reader {
	dr := dirReader{
		tr:   tr,
		hdr:  dirHeader,
		base: path.Clean((*dirHeader).Name),
	}

	// We are just ignoring any error that happens here;
	// the Read() will return the error if there is one.
	*dr.hdr, dr.err = dr.tr.Next()
	return &dr
}

func (dr *dirReader) Read(b []byte) (n int, err error) {
	// Are there previous errors to return?
	if dr.err != nil {
		defer func() { dr.err = nil }()
		if dr.err == io.EOF {
			return 0, EOA
		}
		return 0, dr.err
	}

	// Try to read, else advance to next entry.
	for {
		// Make sure that the entry belongs to the dir we are trying to read.
		if path.Dir((*dr.hdr).Name) != dr.base {
			return 0, io.EOF
		}

		// Try to read from the current file, and if we do, then return that.
		n, err = dr.tr.Read(b)
		if n > 0 {
			if err == io.EOF {
				err = nil
			}
			break // return n, err
		}

		// We finished the current file, so advance to the next.
		if err != io.EOF {
			break // return 0, err
		}

		*dr.hdr, err = dr.tr.Next()
		if err != nil {
			break // return 0, err
		}
	}

	return n, err
}
