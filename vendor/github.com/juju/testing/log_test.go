// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing

import (
	gc "gopkg.in/check.v1"

	"github.com/juju/loggo"
)

type logSuite struct{}

var _ = gc.Suite(&logSuite{})

func (*logSuite) TestLog(c *gc.C) {
	logger := loggo.GetLogger("test")
	jujuLogger := loggo.GetLogger("juju")
	logConfig = "<root>=DEBUG;juju=TRACE"

	c.Assert(logger.EffectiveLogLevel(), gc.Equals, loggo.WARNING)
	var suite LoggingSuite
	suite.SetUpSuite(c)

	c.Assert(logger.EffectiveLogLevel(), gc.Equals, loggo.DEBUG)
	c.Assert(jujuLogger.EffectiveLogLevel(), gc.Equals, loggo.TRACE)

	logger.Debugf("message 1")
	logger.Tracef("message 2")
	jujuLogger.Tracef("message 3")

	c.Assert(c.GetTestLog(), gc.Matches,
		".*DEBUG test message 1\n"+
			".*TRACE juju message 3\n",
	)
	suite.TearDownSuite(c)
	logger.Debugf("message 1")
	logger.Tracef("message 2")
	jujuLogger.Tracef("message 3")

	c.Assert(c.GetTestLog(), gc.Matches,
		".*DEBUG test message 1\n"+
			".*TRACE juju message 3\n",
	)
	c.Assert(logger.EffectiveLogLevel(), gc.Equals, loggo.WARNING)
	c.Assert(jujuLogger.EffectiveLogLevel(), gc.Equals, loggo.WARNING)
}
