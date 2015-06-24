// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xz

/*
#cgo LDFLAGS: -llzma
#include <lzma.h>
*/
import "C"

import (
	"io"
	"math"
	"unsafe"
)

type Decompressor struct {
	handle C.lzma_stream
	rd     io.Reader
	buffer []byte
	offset int
}

var _ io.ReadCloser = &Decompressor{}

func NewReader(r io.Reader) (*Decompressor, error) {
	dec := new(Decompressor)
	// The zero lzma_stream is the same thing as LZMA_STREAM_INIT.
	dec.rd = r
	dec.buffer = make([]byte, DefaultBufsize)
	dec.offset = DefaultBufsize

	// Initialize decoder
	ret := C.lzma_auto_decoder(&dec.handle, math.MaxUint64, 0)
	if Errno(ret) != Ok {
		return nil, Errno(ret)
	}

	return dec, nil
}

func (r *Decompressor) Read(out []byte) (out_count int, er error) {
	if r.offset == len(r.buffer) {
		var n int
		n, er = r.rd.Read(r.buffer)
		if n == 0 {
			return 0, er
		}
		r.offset = 0
		r.handle.next_in = (*C.uint8_t)(unsafe.Pointer(&r.buffer[0]))
		r.handle.avail_in = C.size_t(n)
	}

	r.handle.next_out = (*C.uint8_t)(unsafe.Pointer(&out[0]))
	r.handle.avail_out = C.size_t(len(out))

	ret := C.lzma_code(&r.handle, C.lzma_action(Run))
	switch Errno(ret) {
	case Ok:
		break
	case StreamEnd:
		er = io.EOF
	default:
		er = Errno(ret)
	}

	r.offset = len(r.buffer) - int(r.handle.avail_in)

	return len(out) - int(r.handle.avail_out), er
}

// Frees any resources allocated by liblzma. It does not close the
// underlying reader.
func (r *Decompressor) Close() error {
	if r != nil {
		C.lzma_end(&r.handle)
	}
	return nil
}
