/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"context"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type GafferState uint

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Gaffer operations
type Gaffer interface {
	gopi.Unit
}

// GafferKernel operations
type GafferKernel interface {
	gopi.Unit

	// CreateProcess creates a new process, which is ready to run and returns
	// a unique id for that process
	CreateProcess(GafferService) (uint32, error)

	// RunProcess starts a process in NEW state
	RunProcess(uint32) error

	// StopProcess kills a process in RUNNING state
	StopProcess(uint32) error

	// Processes returns a list of running processes, filtered optionally by
	// process id and service id, both can be zero for 'any'
	Processes(uint32, uint32) []GafferProcess

	// Return a list of executables under the gaffer root, or returns
	// an empty list if the gaffer root is not defined. When argument
	// is set to true, then recurse into subfolders
	Executables(bool) []string
}

// GafferService represents a service to be run
type GafferService struct {
	Path        string        // Path represents the path to the service executable
	Wd          string        // Working directory
	User, Group string        // User and group for process
	Timeout     time.Duration // Timeout for maximum run time for process or zero
	Args        []string      // Process arguments
	Sid         uint32        // Service ID
}

// GafferProcess represents a running process
type GafferProcess interface {
	Id() uint32
	Service() GafferService
	State() GafferState
}

// GafferKernelClient represents a connection to a remote kernel service
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
	Processes(context.Context, uint32, uint32) ([]GafferProcess, error)

	// Executables returns the set of executable service names
	Executables(context.Context) ([]string, error)

	// Stream events until cancelled, using a filter
	StreamEvents(context.Context, uint32, uint32) error
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
