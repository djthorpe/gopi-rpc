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
	"sync/atomic"
	"time"

	// Frameworks

	rpc "github.com/djthorpe/gopi-rpc/v2"
	gaffer "github.com/djthorpe/gopi-rpc/v2/unit/gaffer"
	grpc "github.com/djthorpe/gopi-rpc/v2/unit/grpc"
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/v2/protobuf/gaffer"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type KernelService struct {
	Server gopi.RPCServer
	Kernel gaffer.GafferKernel
}

type kernelservice struct {
	base.Unit
	base.PubSub

	server     gopi.RPCServer
	kernel     gaffer.GafferKernel
	cancelling atomic.Value
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (KernelService) Name() string { return "rpc/gaffer/kernel" }

func (config KernelService) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(kernelservice)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *kernelservice) Init(config KernelService) error {
	// Set server
	if config.Server == nil {
		return gopi.ErrBadParameter.WithPrefix("Server")
	} else {
		this.server = config.Server
	}

	// Set kernel
	if config.Kernel == nil {
		return gopi.ErrBadParameter.WithPrefix("Kernel")
	} else {
		this.kernel = config.Kernel
	}

	// Set cancelling to false
	this.cancelling.Store(false)

	// Register with server
	pb.RegisterKernelServer(this.server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return nil
}

func (this *kernelservice) Close() error {
	if err := this.PubSub.Close(); err != nil {
		return err
	}

	// Release resources
	this.kernel = nil
	this.server = nil

	// Return success
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.RPCService

func (this *kernelservice) CancelRequests() error {
	this.Log.Debug("<CancelRequests>")

	// Flag that we are cancelling
	this.cancelling.Store(true)

	// Send stop signal all running processes except for process zero
	this.Log.Debug("<CancelRequests CancelProcesses>")
	this.CancelProcesses(context.Background())

	// Send a NullEvent on the PubSub channel to indicate end
	this.Log.Debug("<CancelRequests Emit>")
	this.PubSub.Emit(gopi.NullEvent)

	this.Log.Debug("<CancelRequests done>")

	// Return success
	return nil
}

func (this *kernelservice) CancelProcesses(ctx context.Context) error {

	// Stop processes which are running
	var root rpc.GafferProcess
	for _, process := range this.kernel.Processes(0, 0) {
		if process.Id() == 0 {
			root = process
			continue
		} else if process.State() == rpc.GAFFER_STATE_RUNNING {
			this.Log.Debug("<CancelProcesses stopping", process.Id())
			if err := this.kernel.StopProcess(process.Id()); err != nil {
				this.Log.Error(fmt.Errorf("CancelProcess: %v: %w", process.Id(), err))
			}
		}
	}

	// Wait for all processes the end, or deadline exceeded
	ticker := time.NewTimer(100 * time.Millisecond)
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			if processesRunning(this.kernel.Processes(0, 0)) {
				ticker.Reset(500 * time.Millisecond)
			} else {
				ticker.Stop()
				break FOR_LOOP
			}
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		}
	}

	// Cancel root process
	if root != nil && root.State() == rpc.GAFFER_STATE_RUNNING {
		if err := this.kernel.StopProcess(root.Id()); err != nil {
			this.Log.Error(fmt.Errorf("CancelProcess: %v: %w", root.Id(), err))
		}
	}

	// Return success
	return nil
}

// processesRunning returns true if all processes are either new or stopped
// (ie, not running)
func processesRunning(processes []rpc.GafferProcess) bool {
	for _, process := range processes {
		if process.Id() == 0 {
			continue
		}
		if process.State() != rpc.GAFFER_STATE_STOPPED && process.State() != rpc.GAFFER_STATE_NEW {
			fmt.Println("<CancelProcesses still running", process.Id())
			return true
		}
	}
	return false
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *kernelservice) String() string {
	str := "<" + this.Log.Name()
	str += " server=" + fmt.Sprint(this.server)
	str += " kernel=" + fmt.Sprint(this.kernel)
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.GafferKernel

func (this *kernelservice) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.Log.Debug("<Ping>")

	return &empty.Empty{}, nil
}

func (this *kernelservice) CreateProcess(_ context.Context, pb *pb.Service) (*pb.ProcessId, error) {
	this.Log.Debug("<CreateProcess service=", pb, ">")

	if cancelling := this.cancelling.Load().(bool); cancelling {
		return nil, gopi.ErrOutOfOrder
	}

	if service := ProtoToService(pb); service.Path() == "" {
		return nil, gopi.ErrBadParameter
	} else if pid, err := this.kernel.CreateProcess(service); err != nil {
		return nil, err
	} else {
		return ProtoFromProcessId(pid), nil
	}
}

func (this *kernelservice) Processes(_ context.Context, pb *pb.ProcessId) (*pb.KernelProcessList, error) {
	this.Log.Debug("<Processes filter=[", pb, "]>")

	processes := this.kernel.Processes(pb.Id, 0)
	return ProtoFromProcessList(processes), nil
}

func (this *kernelservice) Executables(context.Context, *empty.Empty) (*pb.KernelExecutableList, error) {
	this.Log.Debug("<Executables>")

	executables := this.kernel.Executables(true)
	return ProtoFromExecutablesList(executables), nil
}

func (this *kernelservice) RunProcess(_ context.Context, pb *pb.ProcessId) (*empty.Empty, error) {
	this.Log.Debug("<RunProcess id=", pb.GetId(), ">")

	if cancelling := this.cancelling.Load().(bool); cancelling {
		return nil, gopi.ErrOutOfOrder
	}

	return &empty.Empty{}, this.kernel.RunProcess(pb.GetId())
}

func (this *kernelservice) StopProcess(_ context.Context, pb *pb.ProcessId) (*empty.Empty, error) {
	this.Log.Debug("<StopProcess id=", pb.GetId, ">")

	if cancelling := this.cancelling.Load().(bool); cancelling {
		return nil, gopi.ErrOutOfOrder
	}

	return &empty.Empty{}, this.kernel.StopProcess(pb.GetId())
}

func (this *kernelservice) StreamEvents(filter *pb.ProcessId, stream pb.Kernel_StreamEventsServer) error {
	this.Log.Debug("<StreamEvents filter=[", filter, "]>")

	if cancelling := this.cancelling.Load().(bool); cancelling {
		return gopi.ErrOutOfOrder
	}

	// Subscribe to cancels and send an empty message once a second
	cancel := this.PubSub.Subscribe()
	ticker := time.NewTicker(time.Second)

	// Subscribe to messages from kernel
	msgs := this.kernel.Subscribe()

	// Repeat until stream is canceled by server or client
FOR_LOOP:
	for {
		this.Log.Debug("StreamEvents: FOR_LOOP")
		select {
		case <-cancel:
			this.Log.Debug("StreamEvents: Cancel called")
			// Stop ticker, unsubscribe from events
			ticker.Stop()
			this.kernel.Unsubscribe(msgs)
			this.PubSub.Unsubscribe(cancel)
			// Break loop
			break FOR_LOOP
		case msg := <-msgs:
			if event, ok := msg.(GafferKernelEvent); ok {
				if this.applyFilter(filter, event) {
					proto := ProtoFromEvent(event)
					if err := stream.Send(proto); err != nil {
						if grpc.IsErrUnavailable(err) == false {
							this.Log.Error(fmt.Errorf("StreamEvents: %w", err))
						}
						break FOR_LOOP
					}
				}
			}
		case <-ticker.C:
			if err := stream.Send(&pb.KernelProcessEvent{}); err != nil {
				if grpc.IsErrUnavailable(err) == false {
					this.Log.Error(fmt.Errorf("StreamEvents: %w", err))
				}
				break FOR_LOOP
			}
		}
	}

	this.Log.Debug("StreamEvents: Ended")

	// Return success
	return nil
}

func (this *kernelservice) applyFilter(filter *pb.ProcessId, event GafferKernelEvent) bool {
	// Never allow filters on process 0
	if event.Process().Id() == 0 {
		this.Log.Debug(event)
		return false
	}
	// Doesn't match unless the Id is zero or equal to the process id
	if filter.Id != 0 && filter.Id != event.Process().Id() {
		return false
	}
	// Matches
	return true
}
