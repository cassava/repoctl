// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package osutil

import (
	"bytes"
	"crypto/md5"
	"io"
	"os"
)

// CopyFile tries to copy src to dst. If dst already exists, it will be
// overwritten. If it does not exist, it will be created.
func CopyFile(src, dst string) (err error) {
	// Make sure that both files are regular.
	if _, err = FileExists(src); err != nil {
		return
	}
	if _, err = FileExists(dst); err != nil {
		return
	}

	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// MoveFile tries to move src to dst. If dst already exists, it will be
// overwritten.
func MoveFile(src, dst string) error {
	// Make sure that both files are regular.
	if _, err := FileExists(src); err != nil {
		return err
	}
	if _, err := FileExists(dst); err != nil {
		return err
	}

	err := os.Rename(src, dst)
	if err != nil {
		err := CopyFile(src, dst)
		if err != nil {
			return err
		}
		return os.Remove(src)
	}

	return nil
}

// MoveFileLazy is the same as MoveFile, except that it avoids copying
// the file when the destination already has the same contents.
func MoveFileLazy(src, dst string) error {
	same, err := SameContents(src, dst)
	if err != nil {
		return err
	}
	if same {
		return os.Remove(src)
	}
	return MoveFile(src, dst)
}

// CopyFileLazy is the same as CopyFile, except that it avoids copying
// the file when the destination already has the same contents.
func CopyFileLazy(src, dst string) error {
	same, err := SameContents(src, dst)
	if err != nil {
		return err
	}
	if same {
		return nil
	}
	return CopyFile(src, dst)
}

// SameContents returns same = true if src and dst both exist and have the
// same file contents. Whether the file data is at the same place on
// disk is a different question, which is not answered.
//
// If either file is a directory, FileTypeError is returned.
func SameContents(src, dst string) (same bool, err error) {
	ex, err := FileExists(dst)
	if err != nil || !ex {
		return false, err
	}

	same, err = SameFile(src, dst)
	if err != nil || same {
		return same, err
	}

	// TODO: I could make this more efficient, based on the file size.
	ssum, err := sumFile(src)
	if err != nil {
		return false, err
	}
	dsum, err := sumFile(dst)
	if err != nil {
		return false, err
	}
	return bytes.Compare(ssum, dsum) == 0, nil
}

// sumFile creates an md5sum of a file
func sumFile(path string) (sum []byte, err error) {
	h := md5.New()
	f, err := os.Open(path)
	if err != nil {
		return sum, err
	}
	io.Copy(h, f)
	return h.Sum(nil), nil
}

func SameFile(src, dst string) (same bool, err error) {
	fs, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	if fs.IsDir() {
		return false, FileTypeError{src}
	}

	fd, err := os.Stat(dst)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if fd.IsDir() {
		return false, FileTypeError{dst}
	}

	return os.SameFile(fs, fd), nil
}

// Exists returns ex = true if the given file exists, regardless whether
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
	if stat != nil && stat.IsDir() {
		return false, FileTypeError{path}
	}
	return ex, err
}

// DirExists returns ex = true if the file exists and is
// a directory, and returns err != nil if any other error occured (such as
// permission denied).
func DirExists(path string) (ex bool, err error) {
	var stat os.FileInfo

	ex, stat, err = exists(path)
	if stat != nil && !stat.IsDir() {
		return false, FileTypeError{path}
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
