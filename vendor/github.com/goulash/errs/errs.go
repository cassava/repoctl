// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Package errs provides error handlers for functions.
//
// The purpose of the errs (short for error handler) package is to
// provide an error handler base for functions that might encounter
// errors which they do not know what to do with.
//
// For example, there are many valid ways to respond to an error
// when processing a directory of files. You might want to ignore
// them, print them, handle them specially depending on the error.
//
// When writing a function that takes an error handler, allow the
// user to set a default handler and then pass in nil to specify
// that the default is desired:
//
//  func ACME(h errs.Handler, args ...interface{}) error {
//      errs.Init(&h)
//      ...
//      if err = h(err); err != nil {
//          return err
//      }
//      ...
//      return nil
//  }
//
package errs

import (
	"fmt"
	"io"
	"os"
)

// Handler is used by many functions to deal with errors, most of
// which will be nil errors.
//
// There are several Handlers already available for use.
// Most functions expect that you return nil. Program functionality
// may be impaired otherwise.
type Handler func(error) error

// Default is the default Handler that should be used when nil
// is supplied, see AssertHandler.
var Default Handler = Print(os.Stderr)

// Init is called by functions that take an Handler to
// make sure that an Handler is in fact available. It is usually
// valid to provide functions that expect an Handler a nil, because
// this function replaces that value with the default Handler,
// which you can set yourself.
func Init(h *Handler) {
	if *h == nil {
		*h = Default
	}
}

// Quit returns whatever error it receives, which causes functions
// to quit whenever a non-nil error is received. If in doubt, do not
// use this.
func Quit(err error) error {
	return err
}

// Ignore simply ignores all errors. Not recommended.
func Ignore(_ error) error {
	return nil
}

// Print prints any errors to the writer supplied here. The
// format it uses is:
//
//  error: %s.\n
func Print(w io.Writer) Handler {
	if w == nil {
		return Ignore
	}

	return func(err error) error {
		if err != nil {
			fmt.Fprintf(w, "error: %s.\n", err)
		}
		return nil
	}
}

// List is a list of errors, and is used by Bundle.
type List []error

// Bundle bundles all received errors into an ErrorList, which you
// have to supply (as el).
func Bundle(el *List) Handler {
	return func(err error) error {
		if err != nil {
			*el = append(*el, err)
		}
		return nil
	}
}

// Channel sends all received errors that are not nil to the
// supplied channel.
func Channel(ch chan<- error) Handler {
	return func(err error) error {
		if err != nil {
			ch <- err
		}
		return nil
	}
}
