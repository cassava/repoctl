// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/goulash/util"
)

const errLimit = 15

func TestVerCmp(t *testing.T) {
	data, err := util.NewDecompressor("testdata/vercmp.dat.xz")
	if err != nil {
		t.Fatalf("cannot open 'testdata/testdata.dat.xz': %s", err)
	}
	defer data.Close()

	var ec int
	buf := bufio.NewReader(data)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("unexpected error: %s", err)
		}

		vvr := strings.Split(strings.TrimSpace(line), " ")
		if len(vvr) != 3 {
			t.Errorf("unexpected error: expected 3 parts from '%s', got %d", vvr, len(vvr))
			ec++
			continue
		}
		v1, v2 := vvr[0], vvr[1]
		r, err := strconv.Atoi(vvr[2])
		if err != nil {
			t.Errorf("unexpected error: unable to convert number %s", vvr[2])
			ec++
			continue
		}
		if c := VerCmp(v1, v2); c != r {
			t.Errorf("VerCmp: expected %s %c %s; got %s %c %s", v1, cmp2str(r), v2, v1, cmp2str(c), v2)
			ec++
		}

		if ec > errLimit {
			t.Fatalf("too many errors")
		}
	}
}

func cmp2str(c int) rune {
	if c < 0 {
		return '<'
	} else if c > 0 {
		return '>'
	}
	return '='
}
