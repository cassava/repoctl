// Copyright 2012, 2013 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"fmt"

	gc "gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/juju/testing"
)

type mgoSuite struct {
	testing.IsolationSuite
	testing.MgoSuite
}

var _ = gc.Suite(&mgoSuite{})

func (s *mgoSuite) SetUpSuite(c *gc.C) {
	s.IsolationSuite.SetUpSuite(c)
	s.MgoSuite.SetUpSuite(c)
}

func (s *mgoSuite) TearDownSuite(c *gc.C) {
	s.MgoSuite.TearDownSuite(c)
	s.IsolationSuite.TearDownSuite(c)
}

func (s *mgoSuite) SetUpTest(c *gc.C) {
	s.IsolationSuite.SetUpTest(c)
	s.MgoSuite.SetUpTest(c)
}

func (s *mgoSuite) TearDownTest(c *gc.C) {
	s.MgoSuite.TearDownTest(c)
	s.IsolationSuite.TearDownTest(c)
}

func (s *mgoSuite) TestResetWhenUnauthorized(c *gc.C) {
	session, err := testing.MgoServer.Dial()
	c.Assert(err, gc.IsNil)
	defer session.Close()
	err = session.DB("admin").AddUser("admin", "foo", false)
	if err != nil && err.Error() != "need to login" {
		c.Assert(err, gc.IsNil)
	}
	// The test will fail if the reset does not succeed
}

func (s *mgoSuite) TestStartAndClean(c *gc.C) {
	c.Assert(testing.MgoServer.Addr(), gc.Not(gc.Equals), "")

	session, err := testing.MgoServer.Dial()
	c.Assert(err, gc.IsNil)
	defer session.Close()
	menu := session.DB("food").C("menu")
	err = menu.Insert(
		bson.D{{"spam", "lots"}},
		bson.D{{"eggs", "fried"}},
	)
	c.Assert(err, gc.IsNil)
	food := make([]map[string]string, 0)
	err = menu.Find(nil).All(&food)
	c.Assert(err, gc.IsNil)
	c.Assert(food, gc.HasLen, 2)
	c.Assert(food[0]["spam"], gc.Equals, "lots")
	c.Assert(food[1]["eggs"], gc.Equals, "fried")

	testing.MgoServer.Reset()
	morefood := make([]map[string]string, 0)
	err = menu.Find(nil).All(&morefood)
	c.Assert(err, gc.IsNil)
	c.Assert(morefood, gc.HasLen, 0)
}

func (s *mgoSuite) TestStartIPv6(c *gc.C) {
	info := testing.MgoServer.DialInfo()
	info.Addrs = []string{fmt.Sprintf("[::1]:%v", testing.MgoServer.Port())}
	session, err := mgo.DialWithInfo(info)
	c.Assert(err, gc.IsNil)
	session.Close()
}
