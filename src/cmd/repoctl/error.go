// Copyright (c) 2015, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import "text/template"

// Error is the error type that is returned by functions in repoctl.
// Errors are predefined and consist of two components:
//
//  1. msg, the short one-line error message, and
//  2. desc, the longer error explanation.
//
// The longer explanation can be used to help the user fix the problem.
type Error struct {
	Trigger error  // error that triggered this error, can be nil
	Msg     string // short one-line error message
	Desc    string // in depth explanation to the user as to what happened
}

var errorTmpl = template.Must(template.New("error").Parse(`Error: {{.Msg}}`))

var errorDescribeTmpl = template.Must(template.New("error").Parse(`Error: {{.Msg}}`))

func (e *Error) Error() string {
	panic("implement me")
}

func (e *Error) Describe() string {
	panic("implement me")
}
