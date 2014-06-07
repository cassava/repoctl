// Copyright (c) 2014, Ben Morgan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pacman

import (
	"strconv"
	"strings"
)

// VerCmp compares two version strings in a way mostly compatible to the Arch
// Linux vercmp utility.
//
// Caveat: there is currently one break in compatibility: floating point
// releases are rounded down instead of being maintained. Please just don't.
//
// The following is from the vercmp man page:
//  vercmp is used to determine the relationship between two given version
//  numbers. It outputs values as follows:
//
//   < 0 : if ver1 < ver2
//   = 0 : if ver1 == ver2
//   > 0 : if ver1 > ver2
//
//  Version comparison operates as follows:
//
//  Alphanumeric:
//   1.0a < 1.0b < 1.0beta < 1.0p < 1.0pre < 1.0rc < 1.0 < 1.0.a < 1.0.1
//
//  Numeric:
//   1 < 1.0 < 1.1 < 1.1.1 < 1.2 < 2.0 < 3.0.0
//
//  Additionally, version strings can have an epoch value defined that will
//  overrule any version comparison (unless the epoch values are equal). This is
//  specified in an epoch:version-rel format. For example, 2:1.0-1 is always
//  greater than 1:3.6-1.
//
//  Keep in mind that the pkgrel is only compared if it is available on both
//  versions given to this tool. For example, comparing 1.5-1 and 1.5 will yield
//  0; comparing 1.5-1 and 1.5-2 will yield < 0 as expected. This is mainly for
//  supporting versioned dependencies that do not include the pkgrel.
//
func VerCmp(a, b string) int {
	// Shortcut if they are the same or nil.
	if a == b {
		return 0
	} else if a == "" {
		return -1
	} else if b == "" {
		return 1
	}

	e1, v1, r1 := parseIntoEVR(strings.ToLower(a))
	e2, v2, r2 := parseIntoEVR(strings.ToLower(b))

	// Compare epoch values
	if e1 != e2 {
		if e1 < e2 {
			return -1
		}
		return 1
	}

	// Compare versions and if the same, then compare release values
	c := compareVersions(v1, v2)
	if c == 0 && r1 != r2 {
		// If one of r1 and r2 are -1, then we can't compare
		if r1 < 0 || r2 < 0 {
			return 0
		} else if r1 < r2 {
			return -1
		}
		return 1
	}

	return c
}

// parseIntoEVR splits the version string into [epoch:]version[-release]
// If the release portion is not available, then -1 is returned for r.
func parseIntoEVR(a string) (e int, v string, r int) {
	var s, t int

	// Find out our e
	n := len(a)
	for i := 0; i < n; i++ {
		if !isdigit(a[i]) {
			if a[i] == ':' {
				e, _ = strconv.Atoi(a[:i])
				s = i + 1
			}
			break
		}
	}

	// Find out our r
	r, t = -1, len(a)-1
	for i := t; i >= 0; i-- {
		if !isdigit(a[i]) {
			if a[i] == '-' {
				r, _ = strconv.Atoi(a[i+1:])
				t = i
			}
			break
		}
	}

	return e, a[s:t], r
}

// compareVersions compares the version portion of [epoch:]version[-release]
func compareVersions(v1, v2 string) int {
	sep := func(r rune) bool {
		return r == '.' || r == '_' || r == '+'
	}
	s1 := strings.FieldsFunc(v1, sep)
	s2 := strings.FieldsFunc(v2, sep)
	m, n := len(s1), len(s2)
	z := min(m, n)

	//fmt.Printf("cmpver: %s - %s: %v - %v\n", v1, v2, s1, s2)
	for i := 0; i < z; i++ {
		c := comparePart(s1[i], s2[i])
		if c != 0 {
			return c
		}
	}

	return intcmp(m, n)
}

func comparePart(s1, s2 string) int {
	// Shortcut if they are the same.
	if s1 == s2 {
		return 0
	}

	for s1 != "" && s2 != "" {
		p1, num1 := nextSection(&s1)
		p2, num2 := nextSection(&s2)

		//fmt.Printf("cmpsec: %s - %s: %v - %v\n", p1, p2, num1, num2)

		if num1 != num2 {
			if num1 {
				return 1
			}
			return -1
		}

		if num1 {
			a, _ := strconv.Atoi(p1)
			b, _ := strconv.Atoi(p2)
			c := intcmp(a, b)
			if c != 0 {
				return c
			}
		} else {
			c := strcmp(p1, p2)
			if c != 0 {
				return c
			}
		}
	}

	// The part that is longer is considered less mature, because it is
	// generally something like 1.0rc1 < 1.0.
	return intcmp(len(s2), len(s1))
}

func nextSection(a *string) (s string, num bool) {
	n := len(*a)
	if n == 0 {
		return "", false
	}

	f := isalpha
	if isdigit((*a)[0]) {
		f = isdigit
		num = true
	}

	var i int
	for i = 1; i < n; i++ {
		if !f((*a)[i]) {
			break
		}
	}

	//fmt.Printf("nxtsec: %s -> %s, %s\n", *a, (*a)[0:i], (*a)[i:])
	s = (*a)[0:i]
	*a = (*a)[i:]
	return s, num
}

// isdigit returns true if c is a digit.
func isdigit(c byte) bool {
	return '0' <= c && c <= '9'
}

// isalpha returns true if c is part of [a-z].
func isalpha(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// intcmp returns the comparison of two integers.
func intcmp(i1, i2 int) int {
	if i1 < i2 {
		return -1
	} else if i1 > i2 {
		return 1
	}
	return 0
}

func strcmp(a, b string) int {
	m, n := len(a), len(b)
	z := min(m, n)
	for i := 0; i < z; i++ {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return intcmp(m, n)
}

// min returns the lesser of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
