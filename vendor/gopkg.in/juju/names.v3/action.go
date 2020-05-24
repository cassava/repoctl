// Copyright 2014 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import (
	"fmt"
	"regexp"

	"github.com/juju/errors"
	"github.com/juju/utils"
)

const ActionTagKind = "action"

// ActionSnippet defines the regexp for a valid Action Id.
// Actions are identified by a unique, incrementing number.
const ActionSnippet = NumberSnippet

var validActionV2 = regexp.MustCompile("^" + ActionSnippet + "$")

type ActionTag struct {
	// Tags that are serialized need to have fields exported.
	ID string
}

// NewActionTag returns the tag of an action with the given id (UUID).
func NewActionTag(id string) ActionTag {
	// Actions v1 use a UUID for the id.
	if uuid, err := utils.UUIDFromString(id); err == nil {
		return ActionTag{ID: uuid.String()}
	}

	// Actions v2 use a number.
	if !validActionV2.MatchString(id) {
		panic(fmt.Sprintf("invalid action id %q", id))
	}
	return ActionTag{ID: id}
}

// ParseActionTag parses an action tag string.
func ParseActionTag(actionTag string) (ActionTag, error) {
	tag, err := ParseTag(actionTag)
	if err != nil {
		return ActionTag{}, err
	}
	at, ok := tag.(ActionTag)
	if !ok {
		return ActionTag{}, invalidTagError(actionTag, ActionTagKind)
	}
	return at, nil
}

func (t ActionTag) String() string { return t.Kind() + "-" + t.Id() }
func (t ActionTag) Kind() string   { return ActionTagKind }
func (t ActionTag) Id() string     { return t.ID }

// IsValidAction returns whether id is a valid action id.
func IsValidAction(id string) bool {
	// UUID is for actions v1
	// N is for actions V2.
	return utils.IsValidUUIDString(id) ||
		validActionV2.MatchString(id)
}

// ActionReceiverTag returns an ActionReceiver Tag from a
// machine or unit name.
func ActionReceiverTag(name string) (Tag, error) {
	if IsValidUnit(name) {
		return NewUnitTag(name), nil
	}
	if IsValidApplication(name) {
		// TODO(jcw4) enable when leader elections complete
		//return NewApplicationTag(name), nil
	}
	if IsValidMachine(name) {
		return NewMachineTag(name), nil
	}
	return nil, fmt.Errorf("invalid actionreceiver name %q", name)
}

// ActionReceiverFrom Tag returns an ActionReceiver tag from
// a machine or unit tag.
func ActionReceiverFromTag(tag string) (Tag, error) {
	unitTag, err := ParseUnitTag(tag)
	if err == nil {
		return unitTag, nil
	}
	machineTag, err := ParseMachineTag(tag)
	if err == nil {
		return machineTag, nil
	}
	return nil, errors.Errorf("invalid actionreceiver tag %q", tag)
}
