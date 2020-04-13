/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	"context"
	"fmt"
	"sync"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	gopi.Config
	Fifo       string             // Location for the UNIX socket
	Clientpool gopi.RPCClientPool // Clientpool to create connnections and stubs
}

type service struct {
	base.Unit
	sync.Mutex

	kernel rpc.GafferKernelStub // Kernel client
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (Service) Name() string { return "gaffer/service" }

func (config Service) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(service)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	}
	if err := this.Init(config); err != nil {
		return nil, err
	}

	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gaffer.Kernel

func (this *service) Init(config Service) error {
	if config.Fifo == "" {
		return gopi.ErrBadParameter.WithPrefix("gaffer.fifo")
	} else if config.Clientpool == nil {
		return gopi.ErrBadParameter.WithPrefix("clientpool")
	}

	if conn, err := config.Clientpool.ConnectFifo(config.Fifo); err != nil {
		return err
	} else if stub, ok := config.Clientpool.CreateStub("gaffer.Kernel", conn).(rpc.GafferKernelStub); ok == false {
		return gopi.ErrInternalAppError
	} else if err := stub.Ping(context.Background()); err != nil {
		return err
	} else {
		this.kernel = stub
	}

	// Return success
	return nil
}

func (this *service) Close() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Release resources
	this.kernel = nil

	// Return success
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *service) String() string {
	return "<" + this.Log.Name() + " kernel=" + fmt.Sprint(this.kernel) + ">"
}
