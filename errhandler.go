// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package repoctl

import (
	"fmt"
	"io"
	"os"
)

// ErrHandler is used by many functions to deal with errors, most of
// which will be nil errors.
//
// There are several ErrHandlers already available for use.
// Most functions expect that you return nil. Program functionality
// may be impaired otherwise.
type ErrHandler func(error) error

// DefaultEH is the default ErrHandler that should be used when nil
// is supplied, see AssertHandler.
var DefaultEH ErrHandler = PrinterEH(os.Stderr)

// AssertHandler is called by functions that take an ErrHandler to
// make sure that an ErrHandler is in fact available. It is usually
// valid to provide functions that expect an ErrHandler a nil, because
// this function replaces that value with the default ErrHandler,
// which you can set yourself.
func AssertHandler(h *ErrHandler) {
	if *h == nil {
		*h = DefaultEH
	}
}

// QuiterEH returns whatever error it receives, which causes functions
// to quit whenever a non-nil error is received. If in doubt, do not
// use this.
func QuiterEH(err error) error {
	return err
}

// IgnoreEH simply ignores all errors. Not recommended.
func IgnorerEH(_ error) error {
	return nil
}

// PrinterEH prints any errors to the writer supplied here. The
// format it uses is:
//
//  error: %s.\n
func PrinterEH(w io.Writer) ErrHandler {
	if w == nil {
		return IgnorerEH
	}

	return func(err error) error {
		if err != nil {
			fmt.Fprintf(w, "error: %s.\n", err)
		}
		return nil
	}
}

// ErrorList is a list of errors, and is used by BundlerEH.
type ErrorList []error

// BundlerEH bundles all received errors into an ErrorList, which you
// have to supply (as el).
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

// ChannelerEH sends all received errors that are not nil to the
// supplied channel.
func ChannelerEH(ch chan<- error) ErrHandler {
	return func(err error) error {
		if err != nil {
			ch <- err
		}
		return nil
	}
}
