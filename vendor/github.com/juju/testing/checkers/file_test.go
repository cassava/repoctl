// Copyright 2013 Canonical Ltd.
// Copyright 2014 Cloudbase Solutions SRL
// Licensed under the LGPLv3, see LICENCE file for details.

package checkers_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	gc "gopkg.in/check.v1"

	jc "github.com/juju/testing/checkers"
)

type FileSuite struct{}

var _ = gc.Suite(&FileSuite{})

func (s *FileSuite) TestIsNonEmptyFile(c *gc.C) {
	file, err := ioutil.TempFile(c.MkDir(), "")
	c.Assert(err, gc.IsNil)
	fmt.Fprintf(file, "something")
	file.Close()

	c.Assert(file.Name(), jc.IsNonEmptyFile)
}

func (s *FileSuite) TestIsNonEmptyFileWithEmptyFile(c *gc.C) {
	file, err := ioutil.TempFile(c.MkDir(), "")
	c.Assert(err, gc.IsNil)
	file.Close()

	result, message := jc.IsNonEmptyFile.Check([]interface{}{file.Name()}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, file.Name()+" is empty")
}

func (s *FileSuite) TestIsNonEmptyFileWithMissingFile(c *gc.C) {
	name := filepath.Join(c.MkDir(), "missing")

	result, message := jc.IsNonEmptyFile.Check([]interface{}{name}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, name+" does not exist")
}

func (s *FileSuite) TestIsNonEmptyFileWithNumber(c *gc.C) {
	result, message := jc.IsNonEmptyFile.Check([]interface{}{42}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, "obtained value is not a string and has no .String(), int:42")
}

func (s *FileSuite) TestIsDirectory(c *gc.C) {
	dir := c.MkDir()
	c.Assert(dir, jc.IsDirectory)
}

func (s *FileSuite) TestIsDirectoryMissing(c *gc.C) {
	absentDir := filepath.Join(c.MkDir(), "foo")

	result, message := jc.IsDirectory.Check([]interface{}{absentDir}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, absentDir+" does not exist")
}

func (s *FileSuite) TestIsDirectoryWithFile(c *gc.C) {
	file, err := ioutil.TempFile(c.MkDir(), "")
	c.Assert(err, gc.IsNil)
	file.Close()

	result, message := jc.IsDirectory.Check([]interface{}{file.Name()}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, file.Name()+" is not a directory")
}

func (s *FileSuite) TestIsDirectoryWithNumber(c *gc.C) {
	result, message := jc.IsDirectory.Check([]interface{}{42}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, "obtained value is not a string and has no .String(), int:42")
}

func (s *FileSuite) TestDoesNotExist(c *gc.C) {
	absentDir := filepath.Join(c.MkDir(), "foo")
	c.Assert(absentDir, jc.DoesNotExist)
}

func (s *FileSuite) TestDoesNotExistWithPath(c *gc.C) {
	dir := c.MkDir()
	result, message := jc.DoesNotExist.Check([]interface{}{dir}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, dir+" exists")
}

func (s *FileSuite) TestDoesNotExistWithSymlink(c *gc.C) {
	dir := c.MkDir()
	deadPath := filepath.Join(dir, "dead")
	symlinkPath := filepath.Join(dir, "a-symlink")
	err := os.Symlink(deadPath, symlinkPath)
	c.Assert(err, gc.IsNil)
	// A valid symlink pointing to something that doesn't exist passes.
	// Use SymlinkDoesNotExist to check for the non-existence of the link itself.
	c.Assert(symlinkPath, jc.DoesNotExist)
}

func (s *FileSuite) TestDoesNotExistWithNumber(c *gc.C) {
	result, message := jc.DoesNotExist.Check([]interface{}{42}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, "obtained value is not a string and has no .String(), int:42")
}

func (s *FileSuite) TestSymlinkDoesNotExist(c *gc.C) {
	absentDir := filepath.Join(c.MkDir(), "foo")
	c.Assert(absentDir, jc.SymlinkDoesNotExist)
}

func (s *FileSuite) TestSymlinkDoesNotExistWithPath(c *gc.C) {
	dir := c.MkDir()
	result, message := jc.SymlinkDoesNotExist.Check([]interface{}{dir}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, dir+" exists")
}

func (s *FileSuite) TestSymlinkDoesNotExistWithSymlink(c *gc.C) {
	dir := c.MkDir()
	deadPath := filepath.Join(dir, "dead")
	symlinkPath := filepath.Join(dir, "a-symlink")
	err := os.Symlink(deadPath, symlinkPath)
	c.Assert(err, gc.IsNil)

	result, message := jc.SymlinkDoesNotExist.Check([]interface{}{symlinkPath}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, symlinkPath+" exists")
}

func (s *FileSuite) TestSymlinkDoesNotExistWithNumber(c *gc.C) {
	result, message := jc.SymlinkDoesNotExist.Check([]interface{}{42}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, "obtained value is not a string and has no .String(), int:42")
}

func (s *FileSuite) TestIsSymlink(c *gc.C) {
	file, err := ioutil.TempFile(c.MkDir(), "")
	c.Assert(err, gc.IsNil)
	c.Log(file.Name())
	c.Log(filepath.Dir(file.Name()))
	symlinkPath := filepath.Join(filepath.Dir(file.Name()), "a-symlink")
	err = os.Symlink(file.Name(), symlinkPath)
	c.Assert(err, gc.IsNil)

	c.Assert(symlinkPath, jc.IsSymlink)
}

func (s *FileSuite) TestIsSymlinkWithFile(c *gc.C) {
	file, err := ioutil.TempFile(c.MkDir(), "")
	c.Assert(err, gc.IsNil)
	result, message := jc.IsSymlink.Check([]interface{}{file.Name()}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, jc.Contains, " is not a symlink")
}

func (s *FileSuite) TestIsSymlinkWithDir(c *gc.C) {
	result, message := jc.IsSymlink.Check([]interface{}{c.MkDir()}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, jc.Contains, " is not a symlink")
}

func (s *FileSuite) TestSamePathWithNumber(c *gc.C) {
	result, message := jc.SamePath.Check([]interface{}{42, 52}, nil)
	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, "obtained value is not a string and has no .String(), int:42")
}

func (s *FileSuite) TestSamePathBasic(c *gc.C) {
	dir := c.MkDir()

	result, message := jc.SamePath.Check([]interface{}{dir, dir}, nil)

	c.Assert(result, jc.IsTrue)
	c.Assert(message, gc.Equals, "")
}

type SamePathLinuxSuite struct{}

var _ = gc.Suite(&SamePathLinuxSuite{})

func (s *SamePathLinuxSuite) SetUpSuite(c *gc.C) {
	if runtime.GOOS == "windows" {
		c.Skip("Skipped Linux-intented SamePath tests on Windows.")
	}
}

func (s *SamePathLinuxSuite) TestNotSamePathLinuxBasic(c *gc.C) {
	dir := c.MkDir()
	path1 := filepath.Join(dir, "Test")
	path2 := filepath.Join(dir, "test")

	result, message := jc.SamePath.Check([]interface{}{path1, path2}, nil)

	c.Assert(result, jc.IsFalse)
	c.Assert(message, gc.Equals, "stat "+path1+": no such file or directory")
}

func (s *SamePathLinuxSuite) TestSamePathLinuxSymlinks(c *gc.C) {
	file, err := ioutil.TempFile(c.MkDir(), "")
	c.Assert(err, gc.IsNil)
	symlinkPath := filepath.Join(filepath.Dir(file.Name()), "a-symlink")
	err = os.Symlink(file.Name(), symlinkPath)

	result, message := jc.SamePath.Check([]interface{}{file.Name(), symlinkPath}, nil)

	c.Assert(result, jc.IsTrue)
	c.Assert(message, gc.Equals, "")
}

type SamePathWindowsSuite struct{}

var _ = gc.Suite(&SamePathWindowsSuite{})

func (s *SamePathWindowsSuite) SetUpSuite(c *gc.C) {
	if runtime.GOOS != "windows" {
		c.Skip("Skipped Windows-intented SamePath tests.")
	}
}

func (s *SamePathWindowsSuite) TestNotSamePathBasic(c *gc.C) {
	dir := c.MkDir()
	path1 := filepath.Join(dir, "notTest")
	path2 := filepath.Join(dir, "test")

	result, message := jc.SamePath.Check([]interface{}{path1, path2}, nil)

	c.Assert(result, jc.IsFalse)
	path1 = strings.ToUpper(path1)
	c.Assert(message, gc.Equals, "GetFileAttributesEx "+path1+": The system cannot find the file specified.")
}

func (s *SamePathWindowsSuite) TestSamePathWindowsCaseInsensitive(c *gc.C) {
	dir := c.MkDir()
	path1 := filepath.Join(dir, "Test")
	path2 := filepath.Join(dir, "test")

	result, message := jc.SamePath.Check([]interface{}{path1, path2}, nil)

	c.Assert(result, jc.IsTrue)
	c.Assert(message, gc.Equals, "")
}

func (s *SamePathWindowsSuite) TestSamePathWindowsFixSlashes(c *gc.C) {
	result, message := jc.SamePath.Check([]interface{}{"C:/Users", "C:\\Users"}, nil)

	c.Assert(result, jc.IsTrue)
	c.Assert(message, gc.Equals, "")
}

func (s *SamePathWindowsSuite) TestSamePathShortenedPaths(c *gc.C) {
	dir := c.MkDir()
	dir1, err := ioutil.TempDir(dir, "Programming")
	defer os.Remove(dir1)
	c.Assert(err, gc.IsNil)
	result, message := jc.SamePath.Check([]interface{}{dir + "\\PROGRA~1", dir1}, nil)

	c.Assert(result, jc.IsTrue)
	c.Assert(message, gc.Equals, "")
}

func (s *SamePathWindowsSuite) TestSamePathShortenedPathsConsistent(c *gc.C) {
	dir := c.MkDir()
	dir1, err := ioutil.TempDir(dir, "Programming")
	defer os.Remove(dir1)
	c.Assert(err, gc.IsNil)
	dir2, err := ioutil.TempDir(dir, "Program Files")
	defer os.Remove(dir2)
	c.Assert(err, gc.IsNil)

	result, message := jc.SamePath.Check([]interface{}{dir + "\\PROGRA~1", dir2}, nil)

	c.Assert(result, gc.Not(jc.IsTrue))
	c.Assert(message, gc.Equals, "Not the same file")

	result, message = jc.SamePath.Check([]interface{}{"C:/PROGRA~2", "C:/Program Files (x86)"}, nil)

	c.Assert(result, jc.IsTrue)
	c.Assert(message, gc.Equals, "")
}
