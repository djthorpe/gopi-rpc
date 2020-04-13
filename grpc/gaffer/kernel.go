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
	"time"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
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
	Kernel rpc.GafferKernel
}

type kernelservice struct {
	base.Unit

	server gopi.RPCServer
	kernel rpc.GafferKernel
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

	// Register with server
	pb.RegisterKernelServer(this.server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return nil
}

func (this *kernelservice) Close() error {
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.RPCService

func (this *kernelservice) CancelRequests() error {
	// Do not need to cancel
	return nil
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

func (this *kernelservice) CreateProcess(_ context.Context, pb *pb.KernelService) (*pb.KernelProcessId, error) {
	this.Log.Debug("<CreateProcess service=[", pb, "]>")

	if service := ProtoToService(pb); service.Path == "" {
		return nil, gopi.ErrBadParameter
	} else if pid, err := this.kernel.CreateProcess(service); err != nil {
		return nil, err
	} else {
		return ProtoFromProcessId(pid, service.Sid), nil
	}
}

func (this *kernelservice) Processes(_ context.Context, pb *pb.KernelProcessId) (*pb.KernelProcessList, error) {
	this.Log.Debug("<Processes filter=[", pb, "]>")

	processes := this.kernel.Processes(pb.Id, pb.Sid)
	return ProtoFromProcessList(processes), nil
}

func (this *kernelservice) Executables(context.Context, *empty.Empty) (*pb.KernelExecutableList, error) {
	this.Log.Debug("<Executables>")
	executables := this.kernel.Executables(true)
	return ProtoFromExecutablesList(executables), nil
}

func (this *kernelservice) RunProcess(_ context.Context, pb *pb.KernelProcessId) (*empty.Empty, error) {
	this.Log.Debug("<RunProcess id=", pb.GetId(), ">")
	return &empty.Empty{}, this.kernel.RunProcess(pb.GetId())
}

func (this *kernelservice) StopProcess(_ context.Context, pb *pb.KernelProcessId) (*empty.Empty, error) {
	this.Log.Debug("<StopProcess id=", pb.GetId, ">")
	return &empty.Empty{}, this.kernel.StopProcess(pb.GetId())
}

func (this *kernelservice) StreamEvents(filter *pb.KernelProcessId, stream pb.Kernel_StreamEventsServer) error {
	this.Log.Debug("<StreamEvents filter=[", filter, "]>")

	// Send an empty message once a second
	ticker := time.NewTicker(time.Second)
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			if err := stream.Send(&pb.KernelProcessEvent{}); err != nil {
				this.Log.Error(fmt.Errorf("StreamEvents: %w", err))
				break FOR_LOOP
			}
		}
	}

	// Stop ticker, unsubscribe from events
	ticker.Stop()

	// Return success
	return nil
}
