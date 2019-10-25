// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package archive

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/klauspost/compress/zstd"
	lzma "github.com/remyoudompheng/go-liblzma"
)

// ReadFileFromArchive tries to read the file specified from the (compressed) archive.
// Archive formats supported are:
//	.tar
//	.tar.gz
//	.tar.bz2
//	.tar.xz
//	.tar.zst
func ReadFileFromArchive(archive, file string) ([]byte, error) {
	d, err := NewDecompressor(archive)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	return ReadFileFromTar(d, file)
}

// ReadFileFromTar tries to read the file specified from an opened tar file.
// This function is used together with ReadFileFromArchive, hence the io.Reader.
func ReadFileFromTar(r io.Reader, file string) ([]byte, error) {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if hdr.Name == file {
			bytes, err := ioutil.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return bytes, nil
		}
	}

	return nil, fmt.Errorf("file '%s' not found", file)
}

// Decompressor is a universal decompressor that, given a filepath,
// chooses the appropriate decompression algorithm.
//
// At the moment, only the gzip, bzip2, and lzma (as in ".xz") are
// supported. The decompressor needs to be closed after usage.
type Decompressor struct {
	file   *os.File
	reader io.Reader
	closer io.Closer
}

// NewDecompressor creates a new decompressor based on the file extension
// of the given file. The returned Decompressor can be Read and Closed.
func NewDecompressor(filepath string) (*Decompressor, error) {
	var d Decompressor
	var err error

	d.file, err = os.Open(filepath)
	if err != nil {
		return nil, err
	}

	switch path.Ext(filepath) {
	case ".xz":
		xz, err := lzma.NewReader(d.file)
		if err != nil {
			return nil, err
		}
		d.reader = xz
		d.closer = xz
	case ".gz":
		gz, err := gzip.NewReader(d.file)
		if err != nil {
			return nil, err
		}
		d.reader = gz
		d.closer = gz
	case ".bz2":
		d.reader = bzip2.NewReader(d.file)
	case ".zst":
		zd, err := newZstDecompressor(d.file)
		if err != nil {
			return nil, err
		}
		d.reader = zd
		d.closer = zd
	case ".tar":
		d.reader = d.file
	default:
		return nil, fmt.Errorf("unknown file format")
	}

	return &d, nil
}

func (d *Decompressor) Read(p []byte) (n int, err error) {
	return d.reader.Read(p)
}

func (d *Decompressor) Close() error {
	if d.closer != nil {
		err := d.closer.Close()
		if err != nil {
			return err
		}
	}
	return d.file.Close()
}

// zstDecompressor wraps the zstd.Decoder type to implement io.Closer,
// which it unfortunately doesn't quite implement.
type zstDecompressor struct {
	decoder *zstd.Decoder
}

func newZstDecompressor(r io.Reader) (*zstDecompressor, error) {
	var d zstDecompressor
	z, err := zstd.NewReader(r)
	if err != nil {
		return nil, err
	}
	d.decoder = z
	return &d, nil
}

func (d *zstDecompressor) Read(p []byte) (int, error) {
	return d.decoder.Read(p)
}

func (d *zstDecompressor) Close() error {
	d.decoder.Close()
	return nil
}
