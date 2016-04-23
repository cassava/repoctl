// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/juju/testing"
	gc "gopkg.in/check.v1"
)

var _ = gc.Suite(&tcpProxySuite{})

type tcpProxySuite struct{}

func (*tcpProxySuite) TestTCPProxy(c *gc.C) {
	var wg sync.WaitGroup

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	c.Assert(err, gc.IsNil)
	defer listener.Close()
	wg.Add(1)
	go tcpEcho(&wg, listener)

	p := testing.NewTCPProxy(c, listener.Addr().String())
	c.Assert(p.Addr(), gc.Not(gc.Equals), listener.Addr().String())

	// Dial the proxy and check that we see the text echoed correctly.
	conn, err := net.Dial("tcp", p.Addr())
	c.Assert(err, gc.IsNil)
	defer conn.Close()
	txt := "hello, world\n"
	fmt.Fprint(conn, txt)

	buf := make([]byte, len(txt))
	n, err := io.ReadFull(conn, buf)
	c.Assert(err, gc.IsNil)
	c.Assert(string(buf[0:n]), gc.Equals, txt)

	// Close the connection and check that we see
	// the connection closed for read.
	conn.(*net.TCPConn).CloseWrite()
	n, err = conn.Read(buf)
	c.Assert(err, gc.Equals, io.EOF)
	c.Assert(n, gc.Equals, 0)

	// Make another connection and close the proxy,
	// which should close down the proxy and cause us
	// to get an error.
	conn, err = net.Dial("tcp", p.Addr())
	c.Assert(err, gc.IsNil)
	defer conn.Close()

	p.Close()
	_, err = conn.Read(buf)
	c.Assert(err, gc.Equals, io.EOF)

	// Make sure that we cannot dial the proxy address either.
	conn, err = net.Dial("tcp", p.Addr())
	c.Assert(err, gc.ErrorMatches, ".*connection refused")

	listener.Close()
	// Make sure that all our connections have gone away too.
	wg.Wait()
}

// tcpEcho listens on the given listener for TCP connections,
// writes all traffic received back to the sender, and calls
// wg.Done when all its goroutines have completed.
func tcpEcho(wg *sync.WaitGroup, listener net.Listener) {
	defer wg.Done()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer conn.Close()
			// Echo anything that was written.
			io.Copy(conn, conn)
		}()
	}
}
