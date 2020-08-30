// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package archive

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
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

	return nil, fmt.Errorf("cannot find file %q", file)
}

// ExtractArchive extracts an archive on disk to the provided destination
// directory.
func ExtractArchive(archive, destdir string) error {
	d, err := NewDecompressor(archive)
	if err != nil {
		return err
	}
	defer d.Close()

	return ExtractTar(d, destdir)
}

// ExtractTar extracts all files from the reader into the provided destination
// directory.
func ExtractTar(r io.Reader, destdir string) error {
	tr := tar.NewReader(r)

	mkParentDirs := func(fpath string) error {
		// If the directory component of fpath is already a directory, MkdirAll
		// does nothing and returns nil.
		err := os.MkdirAll(filepath.Dir(fpath), os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("cannot create parent directories for %q: %s", fpath, err)
		}
		return nil
	}

	mkFile := func(fpath string, mode os.FileMode, r io.Reader) error {
		file, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, r)
		return err
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		fpath := filepath.Join(destdir, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(fpath, os.FileMode(hdr.Mode))
			if err != nil {
				return fmt.Errorf("cannot extract directory %q: %s", fpath, err)
			}

		case tar.TypeReg, tar.TypeRegA:
			err = mkParentDirs(fpath)
			if err != nil {
				return err
			}
			err = mkFile(fpath, os.FileMode(hdr.Mode), tr)
			if err != nil {
				return fmt.Errorf("cannot extract file %q: %s", fpath, err)
			}

		case tar.TypeSymlink:
			err = mkParentDirs(fpath)
			if err != nil {
				return err
			}
			err = os.Symlink(hdr.Linkname, fpath)
			if err != nil {
				return fmt.Errorf("cannot extract symlink %q to %q: %s", hdr.Linkname, fpath, err)
			}

		case tar.TypeLink, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			// These types could be potentially handled in the future, for now we'll
			// just ignore them and print a message.
			println("not extracting %q: link, char, block, and fifo types not handled", fpath)

		default:
			// We can pretty much ignore the remaining types, as they aren't
			// something we'd put on the filesystem:
			//
			//	 TypeCont, TypeXHeader, TypeXGlobalHeader, TypeGNUSparse,
			//   TypeGNULongName, TypeGNULongLink
		}
	}

	return nil
}

var ErrUnknownCodec = errors.New("unknown compression codec")

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
// of the given file. If no known file extension is encountered, each of
// the supported compression codecs is tried until a positive match is
// found.
//
// The returned Decompressor can be Read and Closed, even if the underlying
// compressor doesn't support the Close interface.
func NewDecompressor(fpath string) (*Decompressor, error) {
	var d Decompressor
	var err error

	d.file, err = os.Open(fpath)
	if err != nil {
		return nil, err
	}

	tryXz := func() error {
		_, err := d.file.Seek(0, 0)
		if err != nil {
			return err
		}
		xzr, err := xz.NewReader(d.file)
		if err != nil {
			return err
		}
		// Success:
		d.reader = xzr
		return nil
	}

	tryGzip := func() error {
		_, err := d.file.Seek(0, 0)
		if err != nil {
			return err
		}
		gz, err := gzip.NewReader(d.file)
		if err != nil {
			return err
		}
		// Success:
		d.reader = gz
		d.closer = gz
		return nil
	}

	tryZst := func() error {
		_, err := d.file.Seek(0, 0)
		if err != nil {
			return err
		}
		zd, err := zstd.NewReader(d.file)
		if err != nil {
			return err
		}
		// zst.Reader doesn't read the header until we start reading,
		// so we don't know if this is okay yet.
		buf := make([]byte, 1)
		_, err = zd.Read(buf)
		if err != nil {
			return err
		}
		// Success:
		zd.Reset(d.file)
		d.file.Seek(0, 0)
		d.reader = zd
		d.closer = zd.IOReadCloser()
		return nil
	}

	tryBzip2 := func() error {
		_, err := d.file.Seek(0, 0)
		if err != nil {
			return err
		}
		bzd := bzip2.NewReader(d.file)
		// We have to call Read once for the header to be read.
		buf := make([]byte, 0)
		_, err = bzd.Read(buf)
		if err != nil {
			return err
		}
		// Success:
		d.reader = bzd
		return nil
	}

	tryAll := func() bool {
		if tryXz() == nil {
			return true
		}
		if tryGzip() == nil {
			return true
		}
		if tryZst() == nil {
			return true
		}
		if tryBzip2() == nil {
			return true
		}
		return false
	}

	switch filepath.Ext(fpath) {
	case ".xz":
		err = tryXz()
	case ".gz":
		err = tryGzip()
	case ".zst":
		err = tryZst()
	case ".bz2":
		err = tryBzip2()
	case ".tar":
		// special use-case
		d.reader = d.file
	default:
		// Try each of the decompressors:
		ok := tryAll()
		if !ok {
			return nil, ErrUnknownCodec
		}
	}

	if err != nil {
		return nil, err
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
