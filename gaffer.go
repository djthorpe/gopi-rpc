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

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

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

	// Set service mode to manual or auto
	SetServiceModeByName(string, GafferServiceMode) error

	// Return all services, groups and instances
	Services() []GafferService
	Groups() []GafferServiceGroup
	Instances() []GafferServiceInstance
}

type GafferService interface {
	Name() string
	Path() string
	Groups() []string
	Mode() GafferServiceMode
	Instances() uint

	IsMemberOfGroup(string) bool
}

type GafferServiceGroup interface {
	Name() string
}

type GafferServiceInstance interface {
	Id() uint
	Service() string
}

type GafferServiceMode uint

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	GAFFER_MODE_NONE GafferServiceMode = iota
	GAFFER_MODE_MANUAL
	GAFFER_MODE_AUTO
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m GafferServiceMode) String() string {
	switch m {
	case GAFFER_MODE_MANUAL:
		return "GAFFER_MODE_MANUAL"
	case GAFFER_MODE_AUTO:
		return "GAFFER_MODE_AUTO"
	default:
		return "[?? Invalid GafferServiceMode value]"
	}
}
