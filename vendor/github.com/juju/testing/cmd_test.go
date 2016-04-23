// Copyright 2012-2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	gc "gopkg.in/check.v1"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
)

type cmdSuite struct {
	testing.CleanupSuite
}

var _ = gc.Suite(&cmdSuite{})

func (s *cmdSuite) TestHookCommandOutput(c *gc.C) {
	var CommandOutput = (*exec.Cmd).CombinedOutput

	cmdChan, cleanup := testing.HookCommandOutput(&CommandOutput, []byte{1, 2, 3, 4}, nil)
	defer cleanup()

	testCmd := exec.Command("fake-command", "arg1", "arg2")
	out, err := CommandOutput(testCmd)
	c.Assert(err, gc.IsNil)
	cmd := <-cmdChan
	c.Assert(out, gc.DeepEquals, []byte{1, 2, 3, 4})
	c.Assert(cmd.Args, gc.DeepEquals, []string{"fake-command", "arg1", "arg2"})
}

func (s *cmdSuite) EnsureArgFileRemoved(name string) {
	s.AddCleanup(func(c *gc.C) {
		c.Assert(name+".out", jc.DoesNotExist)
	})
}

const testFunc = "test-output"

func (s *cmdSuite) TestPatchExecutableNoArgs(c *gc.C) {
	s.EnsureArgFileRemoved(testFunc)
	testing.PatchExecutableAsEchoArgs(c, s, testFunc)
	output := runCommand(c, testFunc)
	output = strings.TrimRight(output, "\r\n")
	c.Assert(output, gc.Equals, testFunc)
	testing.AssertEchoArgs(c, testFunc)
}

func (s *cmdSuite) TestPatchExecutableWithArgs(c *gc.C) {
	s.EnsureArgFileRemoved(testFunc)
	testing.PatchExecutableAsEchoArgs(c, s, testFunc)
	output := runCommand(c, testFunc, "foo", "bar baz")
	output = strings.TrimRight(output, "\r\n")

	c.Assert(output, gc.DeepEquals, testFunc+" 'foo' 'bar baz'")

	testing.AssertEchoArgs(c, testFunc, "foo", "bar baz")
}

func (s *cmdSuite) TestPatchExecutableThrowError(c *gc.C) {
	testing.PatchExecutableThrowError(c, s, testFunc, 1)
	cmd := exec.Command(testFunc)
	out, err := cmd.CombinedOutput()
	c.Assert(err, gc.ErrorMatches, "exit status 1")
	output := strings.TrimRight(string(out), "\r\n")
	c.Assert(output, gc.Equals, "failing")
}

func (s *cmdSuite) TestCaptureOutput(c *gc.C) {
	f := func() {
		_, err := fmt.Fprint(os.Stderr, "this is stderr")
		c.Assert(err, jc.ErrorIsNil)
		_, err = fmt.Fprint(os.Stdout, "this is stdout")
		c.Assert(err, jc.ErrorIsNil)
	}
	stdout, stderr := testing.CaptureOutput(c, f)
	c.Check(string(stdout), gc.Equals, "this is stdout")
	c.Check(string(stderr), gc.Equals, "this is stderr")
}

var _ = gc.Suite(&ExecHelperSuite{})

type ExecHelperSuite struct {
	testing.PatchExecHelper
}

func (s *ExecHelperSuite) TestExecHelperError(c *gc.C) {
	argChan := make(chan []string, 1)

	cfg := testing.PatchExecConfig{
		Stdout:   "Hellooooo stdout!",
		Stderr:   "Hellooooo stderr!",
		ExitCode: 55,
		Args:     argChan,
	}

	f := s.GetExecCommand(cfg)

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd := f("echo", "hello world!")
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err := cmd.Run()
	c.Assert(err, gc.NotNil)
	_, ok := err.(*exec.ExitError)
	if !ok {
		c.Errorf("Expected *exec.ExitError, but got %T", err)
	} else {
		c.Check(err.Error(), gc.Equals, "exit status 55")
	}
	c.Check(stderr.String(), gc.Equals, cfg.Stderr+"\n")
	c.Check(stdout.String(), gc.Equals, cfg.Stdout+"\n")

	select {
	case args := <-argChan:
		c.Assert(args, gc.DeepEquals, []string{"echo", "hello world!"})
	default:
		c.Fatalf("No arguments passed to output channel")
	}
}

func (s *ExecHelperSuite) TestExecHelper(c *gc.C) {
	argChan := make(chan []string, 1)

	cfg := testing.PatchExecConfig{
		Stdout: "Hellooooo stdout!",
		Stderr: "Hellooooo stderr!",
		Args:   argChan,
	}

	f := s.GetExecCommand(cfg)

	stderr := &bytes.Buffer{}
	stdout := &bytes.Buffer{}
	cmd := f("echo", "hello world!")
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err := cmd.Run()
	c.Assert(err, jc.ErrorIsNil)
	c.Check(stderr.String(), gc.Equals, cfg.Stderr+"\n")
	c.Check(stdout.String(), gc.Equals, cfg.Stdout+"\n")

	select {
	case args := <-argChan:
		c.Assert(args, gc.DeepEquals, []string{"echo", "hello world!"})
	default:
		c.Fatalf("No arguments passed to output channel")
	}
}

func runCommand(c *gc.C, command string, args ...string) string {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	c.Assert(err, jc.ErrorIsNil)
	return string(out)
}
