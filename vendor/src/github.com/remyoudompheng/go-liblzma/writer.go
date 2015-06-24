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
	"bytes"
	"io"
	"unsafe"
)

type Compressor struct {
	handle C.lzma_stream
	writer io.Writer
	buffer []byte
}

var _ io.WriteCloser = &Compressor{}

func NewWriter(w io.Writer, preset Preset) (*Compressor, error) {
	enc := new(Compressor)
	// The zero lzma_stream is the same thing as LZMA_STREAM_INIT.
	enc.writer = w
	enc.buffer = make([]byte, DefaultBufsize)

	// Initialize encoder
	ret := C.lzma_easy_encoder(&enc.handle, C.uint32_t(preset), C.lzma_check(CheckCRC64))
	if Errno(ret) != Ok {
		return nil, Errno(ret)
	}

	return enc, nil
}

// Initializes a XZ encoder with additional settings.
func NewWriterCustom(w io.Writer, preset Preset, check Checksum, bufsize int) (*Compressor, error) {
	enc := new(Compressor)
	// The zero lzma_stream is the same thing as LZMA_STREAM_INIT.
	enc.writer = w
	enc.buffer = make([]byte, bufsize)

	// Initialize encoder
	ret := C.lzma_easy_encoder(&enc.handle, C.uint32_t(preset), C.lzma_check(check))
	if Errno(ret) != Ok {
		return nil, Errno(ret)
	}

	return enc, nil
}

func (enc *Compressor) Write(in []byte) (n int, er error) {
	for n < len(in) {
		enc.handle.next_in = (*C.uint8_t)(unsafe.Pointer(&in[n]))
		enc.handle.avail_in = C.size_t(len(in) - n)
		enc.handle.next_out = (*C.uint8_t)(unsafe.Pointer(&enc.buffer[0]))
		enc.handle.avail_out = C.size_t(len(enc.buffer))

		ret := C.lzma_code(&enc.handle, C.lzma_action(Run))
		switch Errno(ret) {
		case Ok:
			break
		default:
			er = Errno(ret)
		}

		n = len(in) - int(enc.handle.avail_in)
		// Write back result.
		produced := len(enc.buffer) - int(enc.handle.avail_out)
		_, er = enc.writer.Write(enc.buffer[:produced])
		if er != nil {
			// Short write.
			return
		}
	}
	return
}

func (enc *Compressor) Flush() error {
	enc.handle.avail_in = 0

	for {
		enc.handle.next_out = (*C.uint8_t)(unsafe.Pointer(&enc.buffer[0]))
		enc.handle.avail_out = C.size_t(len(enc.buffer))
		ret := C.lzma_code(&enc.handle, C.lzma_action(Finish))

		// Write back result.
		produced := len(enc.buffer) - int(enc.handle.avail_out)
		to_write := bytes.NewBuffer(enc.buffer[:produced])
		_, er := io.Copy(enc.writer, to_write)
		if er != nil {
			// Short write.
			return er
		}

		if Errno(ret) == StreamEnd {
			return nil
		}
	}
	panic("unreachable")
}

// Frees any resources allocated by liblzma. It does not close the
// underlying reader.
func (enc *Compressor) Close() error {
	if enc != nil {
		er := enc.Flush()
		C.lzma_end(&enc.handle)
		if er != nil {
			return er
		}
	}
	return nil
}
