// Copyright 2019 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import (
	"fmt"
	"regexp"
	"strconv"
)

// ControllerAgentTagKind indicates that a tag belongs to a controller agent.
const ControllerAgentTagKind = "controller"

var validControllerAgentId = regexp.MustCompile("^" + NumberSnippet + "$")

// ControllerAgentTag represents a tag used to describe a controller.
type ControllerAgentTag struct {
	id string
}

// NewControllerAgentTag returns the tag of an controller agent with the given id.
func NewControllerAgentTag(id string) ControllerAgentTag {
	_, err := strconv.Atoi(id)
	if err != nil {
		panic(fmt.Sprintf("%q is not a valid controller agent id", id))
	}
	return ControllerAgentTag{id: id}
}

// ParseControllerAgentTag parses a controller agent tag string.
func ParseControllerAgentTag(controllerAgentTag string) (ControllerAgentTag, error) {
	tag, err := ParseTag(controllerAgentTag)
	if err != nil {
		return ControllerAgentTag{}, err
	}
	et, ok := tag.(ControllerAgentTag)
	if !ok {
		return ControllerAgentTag{}, invalidTagError(controllerAgentTag, ControllerAgentTagKind)
	}
	return et, nil
}

// Number returns the controller agent number.
func (t ControllerAgentTag) Number() int {
	n, _ := strconv.Atoi(t.Id())
	return n
}

// String implements Tag.
func (t ControllerAgentTag) String() string { return t.Kind() + "-" + t.Id() }

// Kind implements Tag.
func (t ControllerAgentTag) Kind() string { return ControllerAgentTagKind }

// Id implements Tag.
func (t ControllerAgentTag) Id() string { return t.id }

// IsValidControllerAgent returns whether id is a valid controller agent id.
func IsValidControllerAgent(id string) bool {
	return validControllerAgentId.MatchString(id)
}
