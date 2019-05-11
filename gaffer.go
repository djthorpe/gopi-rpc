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
	"time"

	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

type Gaffer interface {
	gopi.Driver

	// Return all services, groups, instances and executables
	GetServices() []GafferService
	GetGroups() []GafferServiceGroup
	GetInstances() []GafferServiceInstance
	GetExecutables(recursive bool) []string

	// Services
	AddServiceForPath(string) (GafferService, error)
	GetServiceForName(string) GafferService
	RemoveServiceForName(string) error
	SetServiceNameForName(service string, new string) error
	SetServiceModeForName(string, GafferServiceMode) error
	SetServiceInstanceCountForName(service string, count uint) error
	AddServiceGroupForName(service string, group string, position uint) error
	RemoveServiceGroupForName(service string, group string) error

	// Groups
	GetGroupsForNames([]string) []GafferServiceGroup
	AddGroupForName(string) (GafferServiceGroup, error)
	SetGroupNameForName(group string, new string) error
	RemoveGroupForName(string) error

	// Instances
	GenerateInstanceId() uint32
	StartInstanceForServiceName(service string, id uint32) (GafferServiceInstance, error)
	StopInstanceForId(id uint32) error
}

////////////////////////////////////////////////////////////////////////////////
// INSTANCES

type GafferService interface {
	Name() string
	Path() string
	Groups() []string
	Mode() GafferServiceMode
	InstanceCount() uint
	RunTime() time.Duration
	IdleTime() time.Duration

	// Flags
	SetFlag(key, value string) error
	Flags() []string

	// Groups
	IsMemberOfGroup(string) bool
}

type GafferServiceGroup interface {
	Name() string

	// Flags
	SetFlag(key, value string) error
	Flags() []string

	// Env
	SetEnv(key, value string) error
	Env() []string
}

type GafferServiceInstance interface {
	Id() uint32
	Service() GafferService
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
