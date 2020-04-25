/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	"fmt"
	"sync"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type processes struct {
	sync.Mutex
	Log     gopi.Logger
	process map[uint32]rpc.GafferProcess
}

////////////////////////////////////////////////////////////////////////////////
// INIT AND CLOSE

func (this *processes) Init(config Gaffer, log gopi.Logger) error {
	// Set parameters
	this.Log = log
	this.process = make(map[uint32]rpc.GafferProcess, 0)
	// Return success
	return nil
}

func (this *processes) Close() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Release resources
	this.Log = nil
	this.process = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// UPDATE PROCESSES

func (this *processes) modify(processes []rpc.GafferProcess) bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Modified is set to true if process is added, removed or changed
	modified := false

	// TODO
	fmt.Println("processes=", processes)

	return modified
}
