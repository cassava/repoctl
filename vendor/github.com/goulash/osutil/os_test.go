// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package osutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testfile  = "testdata/random_a.dat"
	testother = "testdata/random_b.dat"
	testdest  = "testdata/remove_me.dat"
)

func TestFileExists(z *testing.T) {
	assert := assert.New(z)

	ex, err := FileExists(testfile)
	assert.Nil(err)
	assert.True(ex, "expect file to exist", testfile)

	ex, err = FileExists(testdest)
	assert.Nil(err)
	assert.False(ex, "expect file not to exist", testdest)
}

func TestSameFile(z *testing.T) {
	assert := assert.New(z)

	same, err := SameFile(testfile, testfile)
	assert.Nil(err)
	assert.True(same, "two same files should be the same", testfile)

	same, err = SameFile(testfile, testother)
	assert.Nil(err)
	assert.False(same, "different files are not the same", testfile, testother)

	err = CopyFile(testfile, testdest)
	assert.Nil(err)
	defer os.Remove(testdest)

	same, err = SameFile(testfile, testdest)
	assert.Nil(err)
	assert.False(same, "different files with same contents not the same", testfile, testdest)
}

func TestSameContents(z *testing.T) {
	assert := assert.New(z)

	same, err := SameContents(testfile, testother)
	assert.Nil(err)
	assert.False(same, "different files not same contents", testfile, testother)

	same, err = SameContents(testfile, testfile)
	assert.Nil(err)
	assert.True(same, "same file same contents", testfile)

	err = CopyFile(testfile, testdest)
	assert.Nil(err)
	defer os.Remove(testdest)

	same, err = SameContents(testfile, testdest)
	assert.Nil(err)
	assert.True(same, "copied file same contents", testfile, testdest)
}

func TestFileCopy(z *testing.T) {
	assert := assert.New(z)

	err := CopyFile(testfile, testdest)
	assert.Nil(err)
	defer os.Remove(testdest)

	ex, err := FileExists(testdest)
	assert.Nil(err)
	assert.True(ex, "file should have been copied", testdest)

	same, err := SameContents(testfile, testdest)
	assert.Nil(err)
	assert.True(same, "copied file should be same", testdest)

	err = CopyFile(testother, testdest)
	assert.Nil(err)

	same, err = SameContents(testfile, testdest)
	assert.Nil(err)
	assert.False(same, "second copy should overwrite first")
}

func TestFileCopyLazy(z *testing.T) {
	assert := assert.New(z)

	err := CopyFileLazy(testfile, testdest)
	assert.Nil(err)
	defer os.Remove(testdest)
}
