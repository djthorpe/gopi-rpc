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
	State      string             // Folder for storing state information
}

type gaffer struct {
	base.Unit
	sync.Mutex
	sync.WaitGroup
	services
	processes

	kernel1, kernel2 rpc.GafferKernelStub // Kernel clients
	stop             chan struct{}        // stop service signal
	cancel           context.CancelFunc
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Time to look for new services
	DURATION_DISCOVER = 20 * time.Second
	// Time to update process state
	DURATION_PROCESS = 5 * time.Second
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
	if err := this.services.Init(config, log); err != nil {
		return nil, err
	}
	if err := this.processes.Init(config, log); err != nil {
		return nil, err
	}

	// Stream all events from the kernel on the second channel
	ctx, cancel := context.WithCancel(context.Background())
	this.cancel = cancel
	go func() {
		if err := this.kernel2.StreamEvents(ctx, 0); err != nil {
			this.Unit.Log.Error(fmt.Errorf("StreamEvents: %w", err))
		}
	}()

	// Background orchestrator
	go this.BackgroundProcess()

	// Return success
	return this, nil
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gaffer.Kernel

func (this *gaffer) Init(config Gaffer) error {
	if config.Fifo == "" {
		return gopi.ErrBadParameter.WithPrefix("-kernel.sock")
	} else if config.Clientpool == nil {
		return gopi.ErrBadParameter.WithPrefix("clientpool")
	}

	if kernel, err := NewKernelStub(config); err != nil {
		return err
	} else {
		this.kernel1 = kernel
	}

	if kernel, err := NewKernelStub(config); err != nil {
		return err
	} else {
		this.kernel2 = kernel
	}

	this.stop = make(chan struct{})

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
	this.kernel1 = nil
	this.kernel2 = nil
	this.stop = nil

	// Close Processes
	if err := this.processes.Close(); err != nil {
		return err
	}

	// Close Services
	if err := this.services.Close(); err != nil {
		return err
	}

	// Return success
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *gaffer) String() string {
	return "<" + this.Unit.Log.Name() + " kernel=" + fmt.Sprint(this.kernel1) + ">"
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *gaffer) Start(ctx context.Context, sid uint32) (rpc.GafferProcess, error) {
	// Get service for ID
	if service := this.services.Get(sid); service == nil {
		return nil, gopi.ErrNotFound.WithPrefix("sid")
	} else if service.Enabled() == false {
		return nil, rpc.ERROR_NOT_ENABLED
	} else if pid, err := this.kernel1.CreateProcess(ctx, service); err != nil {
		return nil, err
	} else if process, err := this.kernel1.Processes(ctx, pid); err != nil {
		return nil, err
	} else if len(process) != 1 {
		return nil, gopi.ErrInternalAppError
	} else {
		return process[0], nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND PROCESS

func (this *gaffer) BackgroundProcess() {
	this.WaitGroup.Add(1)
	defer this.WaitGroup.Done()

	// ticker to discover new services and processes
	ticker1 := time.NewTimer(100 * time.Millisecond)
	ticker2 := time.NewTimer(time.Second)

FOR_LOOP:
	for {
		select {
		case <-this.stop:
			ticker1.Stop()
			ticker2.Stop()
			break FOR_LOOP
		case <-ticker1.C:
			// Discover new services every 30 seconds
			if modified, err := this.discoverServices(); err != nil {
				this.Unit.Log.Error(err)
				ticker1.Reset(DURATION_DISCOVER)
			} else if modified {
				ticker1.Reset(DURATION_DISCOVER / 2)
			} else {
				ticker1.Reset(DURATION_DISCOVER)
			}
		case <-ticker2.C:
			// Discover process state changes every 5 seconds
			if modified, err := this.discoverProcesses(); err != nil {
				this.Unit.Log.Error(err)
				ticker2.Reset(DURATION_PROCESS)
			} else if modified {
				ticker2.Reset(DURATION_PROCESS / 2)
			} else {
				ticker2.Reset(DURATION_PROCESS)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// DISCOVER SERVICES

func (this *gaffer) discoverServices() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if executables, err := this.kernel1.Executables(ctx); err != nil {
		return false, err
	} else {
		return this.services.modify(executables), nil
	}
}

func (this *gaffer) discoverProcesses() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if processes, err := this.kernel1.Processes(ctx, 0); err != nil {
		return false, err
	} else {
		return this.processes.modify(processes), nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func NewKernelStub(config Gaffer) (rpc.GafferKernelStub, error) {
	if conn, err := config.Clientpool.ConnectFifo(config.Fifo); err != nil {
		return nil, err
	} else if stub, ok := config.Clientpool.CreateStub("gaffer.Kernel", conn).(rpc.GafferKernelStub); ok == false {
		return nil, gopi.ErrInternalAppError.WithPrefix("CreateStub")
	} else if err := stub.Ping(context.Background()); err != nil {
		return nil, err
	} else {
		return stub, nil
	}
}
