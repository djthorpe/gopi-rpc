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
	gopi.Publisher

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

	// Tuples
	SetServiceFlagsForName(string, Tuples) error
	SetGroupFlagsForName(string, Tuples) error
	SetGroupEnvForName(string, Tuples) error

	// Instances
	GetInstanceForId(id uint32) GafferServiceInstance
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
	Flags() Tuples
	IsMemberOfGroup(string) bool
}

type GafferServiceGroup interface {
	Name() string
	Flags() Tuples
	Env() Tuples
}

type GafferServiceInstance interface {
	Id() uint32
	Service() GafferService
	Flags() Tuples
	Env() Tuples
	Start() time.Time
	Stop() time.Time
	ExitCode() int64
}

type GafferEvent interface {
	gopi.Event

	Type() GafferEventType
	Service() GafferService
	Group() GafferServiceGroup
	Instance() GafferServiceInstance
	Data() []byte
}

type GafferClient interface {
	gopi.RPCClient

	// Ping remote microservice
	Ping() error

	// Return list of executables which can be used as microservices
	ListExecutables() ([]string, error)

	// Return services
	ListServices() ([]GafferService, error)
	ListServicesForGroup(string) ([]GafferService, error)
	GetService(string) (GafferService, error)

	// Return groups
	ListGroups() ([]GafferServiceGroup, error)
	ListGroupsForService(string) ([]GafferServiceGroup, error)
	GetGroup(string) (GafferServiceGroup, error)

	// Return instances
	ListInstances() ([]GafferServiceInstance, error)

	// Add services and groups
	AddServiceForPath(path string) (GafferService, error)
	AddGroupForName(name string) (GafferServiceGroup, error)

	// Remove services and groups
	RemoveServiceForName(name string) error
	RemoveGroupForName(name string) error

	// Start instances
	GetInstanceId() (uint32, error)
	StartInstance(string, uint32) (GafferServiceInstance, error)
	StopInstance(uint32) (GafferServiceInstance, error)
}

type GafferServiceMode uint

type GafferEventType uint

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	GAFFER_MODE_NONE GafferServiceMode = iota
	GAFFER_MODE_MANUAL
	GAFFER_MODE_AUTO
)

const (
	GAFFER_EVENT_NONE GafferEventType = iota
	GAFFER_EVENT_SERVICE_ADD
	GAFFER_EVENT_SERVICE_CHANGE
	GAFFER_EVENT_SERVICE_REMOVE
	GAFFER_EVENT_GROUP_ADD
	GAFFER_EVENT_GROUP_CHANGE
	GAFFER_EVENT_GROUP_REMOVE
	GAFFER_EVENT_INSTANCE_ADD
	GAFFER_EVENT_INSTANCE_START
	GAFFER_EVENT_INSTANCE_RUN
	GAFFER_EVENT_INSTANCE_STOP_OK
	GAFFER_EVENT_INSTANCE_STOP_ERROR
	GAFFER_EVENT_INSTANCE_STOP_ZOMBIE
	GAFFER_EVENT_LOG_STDOUT
	GAFFER_EVENT_LOG_STDERR
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

func (t GafferEventType) String() string {
	switch t {
	case GAFFER_EVENT_SERVICE_ADD:
		return "GAFFER_EVENT_SERVICE_ADD"
	case GAFFER_EVENT_SERVICE_CHANGE:
		return "GAFFER_EVENT_SERVICE_CHANGE"
	case GAFFER_EVENT_SERVICE_REMOVE:
		return "GAFFER_EVENT_SERVICE_REMOVE"
	case GAFFER_EVENT_GROUP_ADD:
		return "GAFFER_EVENT_GROUP_ADD"
	case GAFFER_EVENT_GROUP_CHANGE:
		return "GAFFER_EVENT_GROUP_CHANGE"
	case GAFFER_EVENT_GROUP_REMOVE:
		return "GAFFER_EVENT_GROUP_REMOVE"
	case GAFFER_EVENT_INSTANCE_ADD:
		return "GAFFER_EVENT_INSTANCE_ADD"
	case GAFFER_EVENT_LOG_STDOUT:
		return "GAFFER_EVENT_LOG_STDOUT"
	case GAFFER_EVENT_LOG_STDERR:
		return "GAFFER_EVENT_LOG_STDERR"
	case GAFFER_EVENT_INSTANCE_START:
		return "GAFFER_EVENT_INSTANCE_START"
	case GAFFER_EVENT_INSTANCE_RUN:
		return "GAFFER_EVENT_INSTANCE_RUN"
	case GAFFER_EVENT_INSTANCE_STOP_OK:
		return "GAFFER_EVENT_INSTANCE_STOP_OK"
	case GAFFER_EVENT_INSTANCE_STOP_ERROR:
		return "GAFFER_EVENT_INSTANCE_STOP_ERROR"
	case GAFFER_EVENT_INSTANCE_STOP_ZOMBIE:
		return "GAFFER_EVENT_INSTANCE_STOP_ZOMBIE"
	default:
		return "[?? Invalid GafferEventType value]"
	}
}
