// Package shortry simplifies action assignment.
//
// Please note that the API and the functionality are not yet set, things may
// change in the future.
//
// No validation is performed at the moment. If you input is crap, crap comes out.
package shortry

import (
	"errors"
	"strings"
)

var (
	ErrNotExists = errors.New("no pattern found for key")
	ErrAmbiguous = errors.New("key matches multiple patterns")
	ErrForbidden = errors.New("key contains forbidden characters")

	errFieldsMismatch = errors.New("key and pattern have different number of fields")
	errMismatch       = errors.New("key and pattern differ in field selector")
)

type Shortry struct {
	// Contains the mapping from key to value map.
	// It is not recommended to insert a value that is nil.
	kv map[string]interface{}

	// Character to split category on, currently a point.
	split rune
}

// New returns a new Shortry.
func New(m map[string]interface{}) *Shortry {
	return &Shortry{
		kv:    m,
		split: '.',
	}
}

// Get value for key, if possible.
func (s *Shortry) Get(key string) (interface{}, error) {
	m := s.Matches(key)
	if len(m) == 0 {
		return nil, ErrNotExists
	} else if len(m) > 1 {
		return nil, ErrAmbiguous
	}
	return s.kv[m[0]], nil
}

// Get values for all keys given, in order.
func (s *Shortry) GetAll(key string) []interface{} {
	list := make([]interface{}, 0)
	ms := s.Matches(key)
	for _, m := range ms {
		list = append(list, s.kv[m])
	}
	return list
}

// Exists returns true if there exists a pattern that matches to key.
func (s *Shortry) Exists(key string) bool {
	return len(s.Matches(key)) >= 1
}

// Unique returns whether there exists a only one pattern that matches to key.
func (s *Shortry) Unique(key string) bool {
	return len(s.Matches(key)) == 1
}

// Matches returns a list of matching patterns.
func (s *Shortry) Matches(key string) []string {
	list := make([]string, 0)
	for pattern := range s.kv {
		if matches(key, pattern, s.split) == nil {
			list = append(list, pattern)
		}
	}
	return list
}

// matches returns nil when key matches pattern, given split rune.
//
// It needs to match on every field. If we have db.pending, then we need at least d.p.
// Currently it is case-sensitive.
func matches(key, pattern string, split rune) error {
	kf := strings.FieldsFunc(key, func(r rune) bool { return r == split })
	pf := strings.FieldsFunc(pattern, func(r rune) bool { return r == split })

	if len(kf) > len(pf) {
		return errFieldsMismatch
	}

	for i, k := range kf {
		if !strings.HasPrefix(pf[i], k) {
			return errMismatch
		}
	}

	// I guess everything is ok
	return nil
}
