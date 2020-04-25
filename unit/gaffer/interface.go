/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	// Frameworks
	"time"

	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

// Kernel operations
type GafferKernel interface {
	gopi.PubSub

	// CreateProcess creates a new process, which is ready to run and returns
	// a unique id for that process
	CreateProcess(rpc.GafferService) (uint32, error)

	// Extended version of create process
	CreateProcessEx(uint32, rpc.GafferService, time.Duration) (uint32, error)

	// RunProcess starts a process in NEW state
	RunProcess(uint32) error

	// StopProcess kills a process in RUNNING state
	StopProcess(uint32) error

	// Processes returns a list of running processes, filtered optionally by
	// process id and service id, both can be zero for 'any'
	Processes(uint32, uint32) []rpc.GafferProcess

	// Return a list of executables under the gaffer root, or returns
	// an empty list if the gaffer root is not defined. When argument
	// is set to true, then recurse into subfolders
	Executables(bool) []string
}
