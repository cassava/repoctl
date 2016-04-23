// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"os"
	"runtime"

	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
)

type osEnvSuite struct {
	osEnvSuite testing.OsEnvSuite
}

var _ = gc.Suite(&osEnvSuite{})

func (s *osEnvSuite) SetUpSuite(c *gc.C) {
	s.osEnvSuite = testing.OsEnvSuite{}
}

func (s *osEnvSuite) TestOriginalEnvironment(c *gc.C) {
	// The original environment is properly cleaned and restored.
	err := os.Setenv("TESTING_OSENV_ORIGINAL", "original-value")
	c.Assert(err, gc.IsNil)
	s.osEnvSuite.SetUpSuite(c)
	c.Assert(os.Getenv("TESTING_OSENV_ORIGINAL"), gc.Equals, "")
	s.osEnvSuite.TearDownSuite(c)
	// The environment has been restored.
	c.Assert(os.Getenv("TESTING_OSENV_ORIGINAL"), gc.Equals, "original-value")
}

func (s *osEnvSuite) TestTestingEnvironment(c *gc.C) {
	// Environment variables set up by tests are properly removed.
	s.osEnvSuite.SetUpSuite(c)
	s.osEnvSuite.SetUpTest(c)
	err := os.Setenv("TESTING_OSENV_NEW", "new-value")
	c.Assert(err, gc.IsNil)
	s.osEnvSuite.TearDownTest(c)
	s.osEnvSuite.TearDownSuite(c)
	c.Assert(os.Getenv("TESTING_OSENV_NEW"), gc.Equals, "")
}

func (s *osEnvSuite) TestPreservesTestingVariables(c *gc.C) {
	err := os.Setenv("JUJU_MONGOD", "preserved-value")
	c.Assert(err, gc.IsNil)
	s.osEnvSuite.SetUpSuite(c)
	s.osEnvSuite.SetUpTest(c)
	c.Assert(os.Getenv("JUJU_MONGOD"), gc.Equals, "preserved-value")
	c.Assert(err, gc.IsNil)
	s.osEnvSuite.TearDownTest(c)
	s.osEnvSuite.TearDownSuite(c)
	c.Assert(os.Getenv("JUJU_MONGOD"), gc.Equals, "preserved-value")
}

func (s *osEnvSuite) TestRestoresTestingVariables(c *gc.C) {
	os.Clearenv()
	s.osEnvSuite.SetUpSuite(c)
	s.osEnvSuite.SetUpTest(c)
	err := os.Setenv("JUJU_MONGOD", "test-value")
	c.Assert(err, gc.IsNil)
	s.osEnvSuite.TearDownTest(c)
	s.osEnvSuite.TearDownSuite(c)
	c.Assert(os.Getenv("JUJU_MONGOD"), gc.Equals, "")
}

func (s *osEnvSuite) TestWindowsPreservesPath(c *gc.C) {
	if runtime.GOOS != "windows" {
		c.Skip("Windows-specific test case")
	}
	err := os.Setenv("PATH", "/new/path")
	c.Assert(err, gc.IsNil)
	s.osEnvSuite.SetUpSuite(c)
	s.osEnvSuite.SetUpTest(c)
	c.Assert(os.Getenv("PATH"), gc.Equals, "/new/path")
	s.osEnvSuite.TearDownTest(c)
	s.osEnvSuite.TearDownSuite(c)
	c.Assert(os.Getenv("PATH"), gc.Equals, "/new/path")
}

func (s *osEnvSuite) TestWindowsRestoresPath(c *gc.C) {
	if runtime.GOOS != "windows" {
		c.Skip("Windows-specific test case")
	}
	os.Clearenv()
	s.osEnvSuite.SetUpSuite(c)
	s.osEnvSuite.SetUpTest(c)
	err := os.Setenv("PATH", "/test/path")
	c.Assert(err, gc.IsNil)
	s.osEnvSuite.TearDownTest(c)
	s.osEnvSuite.TearDownSuite(c)
	c.Assert(os.Getenv("PATH"), gc.Equals, "")
}
