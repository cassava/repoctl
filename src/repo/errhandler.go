// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"io"
)

type ErrHandler func(error) error

func PrinterEH(w io.Writer) ErrHandler {
	if w == nil {
		return func(_ error) error {
			return nil
		}
	}

	return func(err error) error {
		if err != nil {
			fmt.Fprintf(w, "Error: %s.\n", err)
		}
		return nil
	}
}

func QuiterEH() ErrHandler {
	return quiterEH
}

func quiterEH(err error) error {
	return err
}

type ErrorList []error

func BundlerEH(el *ErrorList) ErrHandler {
	return func(err error) error {
		if err != nil {
			if *el == nil {
				*el = make(ErrorList, 0, 1)
			}
			*el = append(*el, err)
		}
		return nil
	}
}

func ChannelerEH(ch chan error) ErrHandler {
	return func(err error) error {
		if err != nil {
			ch <- err
		}
		return nil
	}
}
