// Copyright 2020 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import (
	"fmt"
	"regexp"
)

const OperationTagKind = "operation"

// OperationSnippet defines the regexp for a valid Operation Id.
// Operations are identified by a unique, incrementing number.
const OperationSnippet = NumberSnippet

var validOperation = regexp.MustCompile("^" + OperationSnippet + "$")

type OperationTag struct {
	// Tags that are serialized need to have fields exported.
	ID string
}

// NewOperationTag returns the tag of an operation with the given id.
func NewOperationTag(id string) OperationTag {
	if !validOperation.MatchString(id) {
		panic(fmt.Sprintf("invalid operation id %q", id))
	}
	return OperationTag{ID: id}
}

// ParseOperationTag parses an operation tag string.
func ParseOperationTag(operationTag string) (OperationTag, error) {
	tag, err := ParseTag(operationTag)
	if err != nil {
		return OperationTag{}, err
	}
	at, ok := tag.(OperationTag)
	if !ok {
		return OperationTag{}, invalidTagError(operationTag, OperationTagKind)
	}
	return at, nil
}

func (t OperationTag) String() string { return t.Kind() + "-" + t.Id() }
func (t OperationTag) Kind() string   { return OperationTagKind }
func (t OperationTag) Id() string     { return t.ID }

// IsValidOperation returns whether id is a valid operation id.
func IsValidOperation(id string) bool {
	return validOperation.MatchString(id)
}
