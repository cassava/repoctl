// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package util_test

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/goulash/util"
)

func ExampleDirReader() {
	file, err := os.Open("testdata/dir_reader_data.tar")
	if err != nil {
		die(err)
	}
	defer file.Close()

	tr := tar.NewReader(file)
	hdr, err := tr.Next()
	for hdr != nil {
		fi := hdr.FileInfo()
		if !fi.IsDir() {
			fmt.Fprintf(os.Stderr, "error: unexpected file '%s'\n", hdr.Name)

			hdr, err = tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				die(err)
			}

			continue
		}

		fmt.Println(hdr.Name)
		r := util.DirReader(tr, &hdr)
		err = printPrefixed(r, "\t")
		if err != nil {
			if err == util.EOA {
				break
			}
			die(err)
		}
	}

	// Output:
	// dir1/
	// 	dir1/file2 content
	// 	dir1/file1 content
	// dir2/
	// 	dir2/file3
	// 	this sentence should span three files.
	// 	dir2/file2
	// 	and which will appear in betwwen the three files,
	// 	dir2/file1
	// 	Apart from the header which is written in each file,
}

func printPrefixed(r io.Reader, prefix string) error {
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		fmt.Printf("%s%s", prefix, line)
	}
	return nil
}

func die(err error) {
	log.Fatalf("error: %s\n", err)
}
