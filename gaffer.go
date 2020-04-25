/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"context"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type GafferState uint
type Error uint

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Gaffer operations
type Gaffer interface {
	gopi.Unit

	// Return services
	Services() []GafferService

	// Update a service
	Update(GafferService, []string) (GafferService, error)

	// Start a service process
	Start(context.Context, uint32) (GafferProcess, error)
}

// GafferService represents a service which may or may not be running
type GafferService interface {
	Name() string
	Sid() uint32
	Path() string
	Cwd() string
	Args() []string
	User() string
	Group() string
	Enabled() bool
}

// MutableGafferService represents a service which can have fields updated
type MutableGafferService interface {
	GafferService

	SetEnabled(bool) MutableGafferService
	SetName(string) MutableGafferService
}

// GafferProcess represents a running process
type GafferProcess interface {
	Id() uint32
	Service() GafferService
	State() GafferState
}

// GafferKernelStub represents a connection to a remote kernel service
type GafferKernelStub interface {
	gopi.RPCClientStub

	// Ping returns without error if the remote service is running
	Ping(context.Context) error

	// CreateProcess creates a new process from a service and
	// returns a unique process id
	CreateProcess(context.Context, GafferService) (uint32, error)

	// RunProcess runs a created process
	RunProcess(context.Context, uint32) error

	// StopProcess stops a running process
	StopProcess(context.Context, uint32) error

	// Processes returns a filtered set of processes
	Processes(context.Context, uint32) ([]GafferProcess, error)

	// Executables returns the set of executable service names
	Executables(context.Context) ([]string, error)

	// Stream events until cancelled, using a filter
	StreamEvents(context.Context, uint32) error
}

// GafferClientStub represents a connection to a remote gaffer service
type GafferClientStub interface {
	gopi.RPCClientStub

	// Make a mutable service
	Mutable(GafferService) MutableGafferService

	// Ping returns without error if the remote service is running
	Ping(context.Context) error

	// Services returns a list of services registered
	Services(context.Context) ([]GafferService, error)

	// Update a service name, cwd, args, user, group or enabled field
	Update(context.Context, MutableGafferService) (GafferService, error)

	// Start a service
	Start(context.Context, uint32) (GafferService, error)
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	GAFFER_STATE_NONE GafferState = iota
	GAFFER_STATE_NEW
	GAFFER_STATE_RUNNING
	GAFFER_STATE_STOPPING
	GAFFER_STATE_STOPPED
	GAFFER_STATE_STDOUT
	GAFFER_STATE_STDERR
)

const (
	ERROR_NONE Error = iota
	ERROR_NOT_MODIFIED
	ERROR_NOT_ENABLED
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s GafferState) String() string {
	switch s {
	case GAFFER_STATE_NONE:
		return "GAFFER_STATE_NONE"
	case GAFFER_STATE_NEW:
		return "GAFFER_STATE_NEW"
	case GAFFER_STATE_RUNNING:
		return "GAFFER_STATE_RUNNING"
	case GAFFER_STATE_STOPPING:
		return "GAFFER_STATE_STOPPING"
	case GAFFER_STATE_STOPPED:
		return "GAFFER_STATE_STOPPED"
	case GAFFER_STATE_STDOUT:
		return "GAFFER_STATE_STDOUT"
	case GAFFER_STATE_STDERR:
		return "GAFFER_STATE_STDERR"
	default:
		return "[?? Invalid GafferStatus value]"
	}
}

func (e Error) Error() string {
	switch e {
	case ERROR_NONE:
		return "No Error"
	case ERROR_NOT_MODIFIED:
		return "Not Modified"
	case ERROR_NOT_ENABLED:
		return "Not Enabled"
	default:
		return "[?? Invalid Error value]"
	}
}
