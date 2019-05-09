/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

type Gaffer interface {
	gopi.Driver

	// Return list of executables
	Executables(recursive bool) []string

	// Return an existing service
	GetServiceForName(string) GafferService

	// Return an array of service groups or nil if any name could not be found
	GetGroupsForNames([]string) []GafferServiceGroup

	// Return a new service
	AddServiceForPath(string) (GafferService, error)

	// Return a new group
	AddGroupForName(string) (GafferServiceGroup, error)

	// Remove a service
	RemoveServiceForName(string) error

	// Remove a group
	RemoveGroupForName(string) error

	// Return all services and groups
	Services() []GafferService
	Groups() []GafferServiceGroup
}

type GafferService interface {
	Name() string
	Path() string
	Groups() []string
	IsMemberOfGroup(string) bool
}

type GafferServiceGroup interface {
	Name() string
}
