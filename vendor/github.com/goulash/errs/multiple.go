// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package errs

import (
	"fmt"
	"strings"
)

// Collector collects multiple errors and returns a MultipleError
// if any of the errors are non-nil.
type Collector struct {
	Message string
	Errors  []error
}

func NewCollector(msg string) *Collector {
	return &Collector{
		Message: msg,
		Errors:  make([]error, 0),
	}
}

// Add adds err to the list of errors, without checking
// whether it is nil or not.
func (c *Collector) Add(err error) {
	c.Errors = append(c.Errors, err)
}

// Collect adds err if it is non-nil.
func (c *Collector) Collect(err error) {
	if err != nil {
		c.Add(err)
	}
}

// Error returns a MultipleError if it contains any errors,
// otherwise it returns nil.
func (c *Collector) Error() error {
	if len(c.Errors) > 0 {
		return &MultipleError{c.Message, c.Errors}
	}
	return nil
}

type MultipleError struct {
	Message string
	Errors  []error
}

func (e *MultipleError) Error() string {
	xs := make([]string, len(e.Errors))
	for i, e := range e.Errors {
		xs[i] = e.Error()
	}
	return fmt.Sprintf("%s: %s", e.Message, strings.Join(xs, "; "))
}
