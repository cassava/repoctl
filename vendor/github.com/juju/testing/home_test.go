// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/juju/utils"
	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
)

type fakeHomeSuite struct {
	testing.IsolationSuite
	fakeHomeSuite testing.FakeHomeSuite
}

var _ = gc.Suite(&fakeHomeSuite{})

func (s *fakeHomeSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.fakeHomeSuite = testing.FakeHomeSuite{}
	s.fakeHomeSuite.SetUpSuite(c)
}

func (s *fakeHomeSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	utils.SetHome("/tmp/tests")
}

func (s *fakeHomeSuite) TearDownSuite(c *gc.C) {
	s.fakeHomeSuite.TearDownSuite(c)
	s.IsolationSuite.TearDownSuite(c)
}

func (s *fakeHomeSuite) TestHomeCreated(c *gc.C) {
	// A fake home is created and set.
	s.fakeHomeSuite.SetUpTest(c)
	home := utils.Home()
	c.Assert(home, gc.Not(gc.Equals), "/tmp/tests")
	c.Assert(home, jc.IsDirectory)
	s.fakeHomeSuite.TearDownTest(c)
	// The original home has been restored.
	switch runtime.GOOS {
	case "windows":
		c.Assert(utils.Home(), jc.SamePath, "C:/tmp/tests")
	default:
		c.Assert(utils.Home(), jc.SamePath, "/tmp/tests")
	}
}

func (s *fakeHomeSuite) TestSshDirSetUp(c *gc.C) {
	// The SSH directory is properly created and set up.
	s.fakeHomeSuite.SetUpTest(c)
	sshDir := testing.HomePath(".ssh")
	c.Assert(sshDir, jc.IsDirectory)
	PrivKeyFile := filepath.Join(sshDir, "id_rsa")
	c.Assert(PrivKeyFile, jc.IsNonEmptyFile)
	PubKeyFile := filepath.Join(sshDir, "id_rsa.pub")
	c.Assert(PubKeyFile, jc.IsNonEmptyFile)
	s.fakeHomeSuite.TearDownTest(c)
}

type makeFakeHomeSuite struct {
	testing.IsolationSuite
	home *testing.FakeHome
}

var _ = gc.Suite(&makeFakeHomeSuite{})

func (s *makeFakeHomeSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.home = testing.MakeFakeHome(c)
	testFile := testing.TestFile{
		Name: "testfile-name",
		Data: "testfile-data",
	}
	s.home.AddFiles(c, testFile)
}

func (s *makeFakeHomeSuite) TestAddFiles(c *gc.C) {
	// Files are correctly added to the fake home.
	expectedPath := filepath.Join(utils.Home(), "testfile-name")
	contents, err := ioutil.ReadFile(expectedPath)
	c.Assert(err, gc.IsNil)
	c.Assert(string(contents), gc.Equals, "testfile-data")
}

func (s *makeFakeHomeSuite) TestFileContents(c *gc.C) {
	// Files contents are returned as strings.
	contents := s.home.FileContents(c, "testfile-name")
	c.Assert(contents, gc.Equals, "testfile-data")
}

func (s *makeFakeHomeSuite) TestFileExists(c *gc.C) {
	// It is possible to check whether a file exists in the fake home.
	c.Assert(s.home.FileExists("testfile-name"), jc.IsTrue)
	c.Assert(s.home.FileExists("no-such-file"), jc.IsFalse)
}
