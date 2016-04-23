// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"os"

	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
)

type cleanupSuite struct {
	testing.CleanupSuite
}

var _ = gc.Suite(&cleanupSuite{})

func (s *cleanupSuite) TestTearDownSuiteEmpty(c *gc.C) {
	// The suite stack is empty initially, check we can tear that down.
	s.TearDownSuite(c)
	s.SetUpSuite(c)
}

func (s *cleanupSuite) TestTearDownTestEmpty(c *gc.C) {
	// The test stack is empty initially, check we can tear that down.
	s.TearDownTest(c)
	s.SetUpTest(c)
}

func (s *cleanupSuite) TestAddCleanup(c *gc.C) {
	order := []string{}
	s.AddCleanup(func(*gc.C) {
		order = append(order, "first")
	})
	s.AddCleanup(func(*gc.C) {
		order = append(order, "second")
	})

	s.TearDownTest(c)
	c.Assert(order, gc.DeepEquals, []string{"second", "first"})

	// SetUpTest resets the cleanup stack, this stops the cleanup functions
	// being called again.
	s.SetUpTest(c)
}

func (s *cleanupSuite) TestPatchEnvironment(c *gc.C) {
	const envName = "TESTING_PATCH_ENVIRONMENT"
	// remember the old value, and set it to something we can check
	oldValue := os.Getenv(envName)
	os.Setenv(envName, "initial")

	s.PatchEnvironment(envName, "new value")
	// Using check to make sure the environment gets set back properly in the test.
	c.Check(os.Getenv(envName), gc.Equals, "new value")

	s.TearDownTest(c)
	c.Check(os.Getenv(envName), gc.Equals, "initial")

	// SetUpTest resets the cleanup stack, this stops the cleanup functions
	// being called again.
	s.SetUpTest(c)
	// explicitly return the envName to the old value
	os.Setenv(envName, oldValue)
}

func (s *cleanupSuite) TestPatchValueInt(c *gc.C) {
	i := 42
	s.PatchValue(&i, 0)
	c.Assert(i, gc.Equals, 0)

	s.TearDownTest(c)
	c.Assert(i, gc.Equals, 42)

	// SetUpTest resets the cleanup stack, this stops the cleanup functions
	// being called again.
	s.SetUpTest(c)
}

func (s *cleanupSuite) TestPatchValueFunction(c *gc.C) {
	function := func() string {
		return "original"
	}

	s.PatchValue(&function, func() string {
		return "patched"
	})
	c.Assert(function(), gc.Equals, "patched")

	s.TearDownTest(c)
	c.Assert(function(), gc.Equals, "original")

	// SetUpTest resets the cleanup stack, this stops the cleanup functions
	// being called again.
	s.SetUpTest(c)
}

// noopCleanup is a simple function that does nothing that can be passed to
// AddCleanup
func noopCleanup(*gc.C) {
}

func (s cleanupSuite) TestAddCleanupPanicIfUnsafe(c *gc.C) {
	// It is unsafe to call AddCleanup when the test itself is not a
	// pointer receiver, because AddCleanup modifies the s.testStack
	// attribute, but in a non-pointer receiver, that object is lost when
	// the Test function returns.
	// This Test must, itself, be a non pointer receiver to trigger this
	c.Assert(func() { s.AddCleanup(noopCleanup) },
		gc.PanicMatches,
		"unsafe to call AddCleanup from non pointer receiver test")
}

type cleanupSuiteAndTestLifetimes struct {
}

var _ = gc.Suite(&cleanupSuiteAndTestLifetimes{})

func (s *cleanupSuiteAndTestLifetimes) TestAddCleanupBeforeSetUpSuite(c *gc.C) {
	suite := &testing.CleanupSuite{}
	c.Assert(func() { suite.AddCleanup(noopCleanup) },
		gc.PanicMatches,
		"unsafe to call AddCleanup before SetUpSuite")
	suite.SetUpSuite(c)
	suite.SetUpTest(c)
	suite.TearDownTest(c)
	suite.TearDownSuite(c)
}

func (s *cleanupSuiteAndTestLifetimes) TestAddCleanupAfterTearDownSuite(c *gc.C) {
	suite := &testing.CleanupSuite{}
	suite.SetUpSuite(c)
	suite.SetUpTest(c)
	suite.TearDownTest(c)
	suite.TearDownSuite(c)
	c.Assert(func() { suite.AddCleanup(noopCleanup) },
		gc.PanicMatches,
		"unsafe to call AddCleanup after TearDownSuite")
}

func (s *cleanupSuiteAndTestLifetimes) TestAddCleanupMixedSuiteAndTest(c *gc.C) {
	calls := []string{}
	suite := &testing.CleanupSuite{}
	suite.SetUpSuite(c)
	suite.AddCleanup(func(*gc.C) { calls = append(calls, "before SetUpTest") })
	suite.SetUpTest(c)
	suite.AddCleanup(func(*gc.C) { calls = append(calls, "during Test1") })
	suite.TearDownTest(c)
	c.Check(calls, gc.DeepEquals, []string{
		"during Test1",
	})
	c.Assert(func() { suite.AddCleanup(noopCleanup) },
		gc.PanicMatches,
		"unsafe to call AddCleanup after a test has been torn down"+
			" before a new test has been set up"+
			" \\(Suite level changes only make sense before first test is run\\)")
	suite.SetUpTest(c)
	suite.AddCleanup(func(*gc.C) { calls = append(calls, "during Test2") })
	suite.TearDownTest(c)
	c.Check(calls, gc.DeepEquals, []string{
		"during Test1",
		"during Test2",
	})
	suite.TearDownSuite(c)
	c.Check(calls, gc.DeepEquals, []string{
		"during Test1",
		"during Test2",
		"before SetUpTest",
	})
}
