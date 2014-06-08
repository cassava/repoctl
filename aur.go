// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import "errors"

func ReadAUR(pkgname string) (*Package, error) {
	//https://aur.archlinux.org/rpc.php?type=info&arg=dropbox
	return nil, errors.New("not implemented")
}

func SearchAUR(match string) ([]*Package, error) {
	return nil, errors.New("not implemented")
}
