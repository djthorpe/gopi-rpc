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

type GafferStatus uint

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

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
}

// GafferService represents a service to be run
type GafferService struct {
	Path        string        // Path represents the path to the process executable
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
	Status() GafferStatus
}

// GafferKernelClient represents a connection to a remote kernel service
type GafferKernelStub interface {
	// Ping returns nil if the remote service is running
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

	// Stream events until cancelled, using a filter
	StreamEvents(context.Context, uint32, uint32) error
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	GAFFER_STATUS_NONE GafferStatus = iota
	GAFFER_STATUS_NEW
	GAFFER_STATUS_RUNNING
	GAFFER_STATUS_STOPPING
	GAFFER_STATUS_STOPPED
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s GafferStatus) String() string {
	switch s {
	case GAFFER_STATUS_NONE:
		return "GAFFER_STATUS_NONE"
	case GAFFER_STATUS_NEW:
		return "GAFFER_STATUS_NEW"
	case GAFFER_STATUS_RUNNING:
		return "GAFFER_STATUS_RUNNING"
	case GAFFER_STATUS_STOPPING:
		return "GAFFER_STATUS_STOPPING"
	case GAFFER_STATUS_STOPPED:
		return "GAFFER_STATUS_STOPPED"
	default:
		return "[?? Invalid GafferStatus value]"
	}
}
