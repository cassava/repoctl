// Copyright 2015 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package names

import (
	"fmt"
	"regexp"
)

const SubnetTagKind = "subnet"

var validSubnet = regexp.MustCompile("^" + NumberSnippet + "$")

// IsValidSubnet returns whether id is a valid subnet id.
func IsValidSubnet(id string) bool {
	return validSubnet.MatchString(id)
}

type SubnetTag struct {
	id string
}

func (t SubnetTag) String() string { return t.Kind() + "-" + t.id }
func (t SubnetTag) Kind() string   { return SubnetTagKind }
func (t SubnetTag) Id() string     { return t.id }

// NewSubnetTag returns the tag for subnet with the given ID.
func NewSubnetTag(id string) SubnetTag {
	if !IsValidSubnet(id) {
		panic(fmt.Sprintf("%s is not a valid subnet ID", id))
	}
	return SubnetTag{id: id}
}

// ParseSubnetTag parses a subnet tag string.
func ParseSubnetTag(subnetTag string) (SubnetTag, error) {
	tag, err := ParseTag(subnetTag)
	if err != nil {
		return SubnetTag{}, err
	}
	subt, ok := tag.(SubnetTag)
	if !ok {
		return SubnetTag{}, invalidTagError(subnetTag, SubnetTagKind)
	}
	return subt, nil
}
