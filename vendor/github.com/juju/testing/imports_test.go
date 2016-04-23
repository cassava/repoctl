// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package testing_test

import (
	"go/build"
	"os"
	"path/filepath"
	"text/template"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type importsSuite struct {
	testing.CleanupSuite
}

var _ = gc.Suite(&importsSuite{})

var pkgs = [][]string{{
	"arble.com/foo", "arble.com/bar", "arble.com/baz", "fmt",
}, {
	"arble.com/bar", "arble.com/baz",
}, {
	"arble.com/baz", "math",
}, {
	"arble.com/bar", "furble.com/fur",
}, {
	"furble.com/fur", "fmt", "C",
}}

var importsTests = []struct {
	pkgName string
	prefix  string
	expect  []string
}{{
	pkgName: "arble.com/foo",
	prefix:  "arble.com/",
	expect:  []string{"bar", "baz"},
}, {
	pkgName: "arble.com/foo",
	prefix:  "furble.com/",
	expect:  []string{"fur"},
}, {
	pkgName: "furble.com/fur",
	prefix:  "arble.com/",
	expect:  nil,
}}

func (s *importsSuite) TestImports(c *gc.C) {
	goPath := writePkgs(c)
	s.PatchValue(&build.Default.GOPATH, goPath)

	c.Logf("gopath %q", build.Default.GOPATH)
	for i, test := range importsTests {
		c.Logf("test %d: %s %s", i, test.pkgName, test.prefix)
		imports, err := testing.FindImports(test.pkgName, test.prefix)
		c.Assert(err, gc.IsNil)
		c.Assert(imports, jc.DeepEquals, test.expect)
	}
}

func writePkgs(c *gc.C) (goPath string) {
	goPath = c.MkDir()
	for _, p := range pkgs {
		dir := filepath.Join(goPath, "src", p[0])
		err := os.MkdirAll(dir, 0777)
		c.Assert(err, gc.IsNil)
		f, err := os.Create(filepath.Join(dir, "pkg.go"))
		c.Assert(err, gc.IsNil)
		defer f.Close()
		err = sourceTemplate.Execute(f, p[1:])
		c.Assert(err, gc.IsNil)
	}
	return
}

var sourceTemplate = template.Must(template.New("").Parse(`
package pkg
import ({{range $f := $}}
_ {{printf "%q" .}}
{{end}})
`))
