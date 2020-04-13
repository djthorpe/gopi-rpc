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
	"time"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Gaffer struct {
	gopi.Config
	Fifo       string             // Location for the UNIX socket
	Clientpool gopi.RPCClientPool // Clientpool to create connnections and stubs
}

type gaffer struct {
	base.Unit
	sync.Mutex
	sync.WaitGroup
	Services

	kernel rpc.GafferKernelStub // Kernel client
	stop   chan struct{}        // stop service signal
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Time to look for new services
	DURATION_DISCOVER = 20 * time.Second
)

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (Gaffer) Name() string { return "gaffer/service" }

func (config Gaffer) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(gaffer)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	}
	if err := this.Init(config); err != nil {
		return nil, err
	}
	if err := this.Services.Init(config, log); err != nil {
		return nil, err
	}

	// Background orchestrator
	go this.BackgroundProcess()

	// Return success
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gaffer.Kernel

func (this *gaffer) Init(config Gaffer) error {
	if config.Fifo == "" {
		return gopi.ErrBadParameter.WithPrefix("gaffer.fifo")
	} else if config.Clientpool == nil {
		return gopi.ErrBadParameter.WithPrefix("clientpool")
	}

	if conn, err := config.Clientpool.ConnectFifo(config.Fifo); err != nil {
		return err
	} else if stub, ok := config.Clientpool.CreateStub("gaffer.Kernel", conn).(rpc.GafferKernelStub); ok == false {
		return gopi.ErrInternalAppError.WithPrefix("CreateStub")
	} else if err := stub.Ping(context.Background()); err != nil {
		return err
	} else {
		this.kernel = stub
	}

	// Return success
	return nil
}

func (this *gaffer) Close() error {
	// signal stop and wait for end
	close(this.stop)
	this.WaitGroup.Wait()

	// Lock to release resources
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Release resources
	this.kernel = nil
	this.stop = nil

	// Close Services
	if err := this.Services.Close(); err != nil {
		return err
	}

	// Return success
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *gaffer) String() string {
	return "<" + this.Unit.Log.Name() + " kernel=" + fmt.Sprint(this.kernel) + ">"
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND PROCESS

func (this *gaffer) BackgroundProcess() {
	this.WaitGroup.Add(1)
	defer this.WaitGroup.Done()

	// ticker to discover new services
	ticker := time.NewTimer(100 * time.Millisecond)

FOR_LOOP:
	for {
		select {
		case <-this.stop:
			ticker.Stop()
			break FOR_LOOP
		case <-ticker.C:
			// Discover new services every 30 seconds
			if modified, err := this.discoverServices(); err != nil {
				this.Unit.Log.Error(err)
				ticker.Reset(DURATION_DISCOVER)
			} else if modified {
				ticker.Reset(DURATION_DISCOVER / 2)
			} else {
				ticker.Reset(DURATION_DISCOVER)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// DISCOVER SERVICES

func (this *gaffer) discoverServices() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if executables, err := this.kernel.Executables(ctx); err != nil {
		return false, err
	} else {
		return this.Services.Modify(executables), nil
	}
}
