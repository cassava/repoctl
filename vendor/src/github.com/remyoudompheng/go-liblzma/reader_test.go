// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xz

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestDecompress(T *testing.T) {
	f, er := os.Open("testdata/go_spec.html.xz")
	if er != nil {
		T.Fatalf("could not open test file: %s", er)
	}
	defer f.Close()
	dec, _ := NewReader(f)
	total := 0
	for {
		var buf [2048]byte
		n, er := dec.Read(buf[:])
		total += n
		if n == 0 || er != nil {
			T.Log(er)
			break
		}
	}
	T.Logf("Total %d bytes written", total)
}

func TestDecompressSmall(t *testing.T) {
	f, _ := os.Open("testdata/go_spec.html.xz")
	dec, _ := NewReader(f)
	buf := new(bytes.Buffer)
	io.Copy(buf, dec)
	contents := buf.Bytes()
	f.Close()

	f, _ = os.Open("testdata/go_spec.html.xz")
	dec, _ = NewReader(f)
	var contents2 []byte
	for {
		var buf [14]byte
		n, er := dec.Read(buf[:])
		contents2 = append(contents2, buf[:n]...)
		if n == 0 || er != nil {
			t.Log(er)
			break
		}
	}

	if !bytes.Equal(contents, contents2) {
		t.Fatalf("contents (%d bytes) and contents2 (%d bytes) differ!", len(contents), len(contents2))
	}
}
