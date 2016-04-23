// Copyright 2013, 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	stdtesting "testing"

	"github.com/juju/testing"
)

func Test(t *stdtesting.T) {
	testing.MgoTestPackage(t, nil)
}
