// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package checkers_test

import (
	gc "gopkg.in/check.v1"

	jc "github.com/juju/testing/checkers"
)

type Inner struct {
	First  string
	Second int             `json:",omitempty" yaml:",omitempty"`
	Third  map[string]bool `json:",omitempty" yaml:",omitempty"`
}

type Outer struct {
	First  float64
	Second []*Inner `json:"Last,omitempty" yaml:"last,omitempty"`
}

func (s *CheckerSuite) TestJSONEquals(c *gc.C) {
	tests := []struct {
		descr    string
		obtained string
		expected *Outer
		result   bool
		msg      string
	}{
		{
			descr:    "very simple",
			obtained: `{"First": 47.11}`,
			expected: &Outer{
				First: 47.11,
			},
			result: true,
		}, {
			descr:    "nested",
			obtained: `{"First": 47.11, "Last": [{"First": "Hello", "Second": 42}]}`,
			expected: &Outer{
				First: 47.11,
				Second: []*Inner{
					{First: "Hello", Second: 42},
				},
			},
			result: true,
		}, {
			descr: "nested with newline",
			obtained: `{"First": 47.11, "Last": [{"First": "Hello", "Second": 42},
			{"First": "World", "Third": {"T": true, "F": false}}]}`,
			expected: &Outer{
				First: 47.11,
				Second: []*Inner{
					{First: "Hello", Second: 42},
					{First: "World", Third: map[string]bool{
						"F": false,
						"T": true,
					}},
				},
			},
			result: true,
		}, {
			descr:    "illegal field",
			obtained: `{"NotThere": 47.11}`,
			expected: &Outer{
				First: 47.11,
			},
			result: false,
			msg:    `mismatch at .*: validity mismatch; .*`,
		}, {
			descr:    "illegal optained content",
			obtained: `{"NotThere": `,
			result:   false,
			msg:      `cannot unmarshal obtained contents: unexpected end of JSON input; .*`,
		},
	}
	for i, test := range tests {
		c.Logf("test #%d) %s", i, test.descr)
		result, msg := jc.JSONEquals.Check([]interface{}{test.obtained, test.expected}, nil)
		c.Check(result, gc.Equals, test.result)
		c.Check(msg, gc.Matches, test.msg)
	}

	// Test non-string input.
	result, msg := jc.JSONEquals.Check([]interface{}{true, true}, nil)
	c.Check(result, gc.Equals, false)
	c.Check(msg, gc.Matches, "expected string, got bool")
}

func (s *CheckerSuite) TestYAMLEquals(c *gc.C) {
	tests := []struct {
		descr    string
		obtained string
		expected *Outer
		result   bool
		msg      string
	}{
		{
			descr:    "very simple",
			obtained: `first: 47.11`,
			expected: &Outer{
				First: 47.11,
			},
			result: true,
		}, {
			descr:    "nested",
			obtained: `{first: 47.11, last: [{first: 'Hello', second: 42}]}`,
			expected: &Outer{
				First: 47.11,
				Second: []*Inner{
					{First: "Hello", Second: 42},
				},
			},
			result: true,
		}, {
			descr: "nested with newline",
			obtained: `{first: 47.11, last: [{first: 'Hello', second: 42},
			{first: 'World', third: {t: true, f: false}}]}`,
			expected: &Outer{
				First: 47.11,
				Second: []*Inner{
					{First: "Hello", Second: 42},
					{First: "World", Third: map[string]bool{
						"f": false,
						"t": true,
					}},
				},
			},
			result: true,
		}, {
			descr:    "illegal field",
			obtained: `{"NotThere": 47.11}`,
			expected: &Outer{
				First: 47.11,
			},
			result: false,
			msg:    `mismatch at .*: validity mismatch; .*`,
		}, {
			descr:    "illegal obtained content",
			obtained: `{"NotThere": `,
			result:   false,
			msg:      `cannot unmarshal obtained contents: yaml: line 1: .*`,
		},
	}
	for i, test := range tests {
		c.Logf("test #%d) %s", i, test.descr)
		result, msg := jc.YAMLEquals.Check([]interface{}{test.obtained, test.expected}, nil)
		c.Check(result, gc.Equals, test.result)
		c.Check(msg, gc.Matches, test.msg)
	}

	// Test non-string input.
	result, msg := jc.YAMLEquals.Check([]interface{}{true, true}, nil)
	c.Check(result, gc.Equals, false)
	c.Check(msg, gc.Matches, "expected string, got bool")
}
