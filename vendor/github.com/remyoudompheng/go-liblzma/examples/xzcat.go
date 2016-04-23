// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"os"
	xz "github.com/remyoudompheng/go-liblzma"
)

func main() {
	dec, er := xz.NewReader(os.Stdin)
	if er != nil {
		fmt.Println(er)
		os.Exit(1)
	}

	io.Copy(os.Stdout, dec)
}
