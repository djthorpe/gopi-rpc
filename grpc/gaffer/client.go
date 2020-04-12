/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	// Frameworks
	"context"
	"fmt"

	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
	"github.com/golang/protobuf/ptypes/empty"

	// Protocol buffers
	rpc "github.com/djthorpe/gopi-rpc/v2"
	pb "github.com/djthorpe/gopi-rpc/v2/protobuf/gaffer"
	grpc "github.com/djthorpe/gopi-rpc/v2/unit/grpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type KernelClient struct {
	Conn gopi.RPCClientConn
}

type kernelclient struct {
	base.Unit
	conn   gopi.RPCClientConn
	client pb.KernelClient
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (KernelClient) Name() string { return "gaffer.Kernel" }

func (config KernelClient) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(kernelclient)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *kernelclient) Init(config KernelClient) error {
	// Create the client
	if config.Conn == nil {
		return gopi.ErrBadParameter
	} else if grpcconn, ok := config.Conn.(grpc.GRPCClientConn); ok == false {
		return gopi.ErrBadParameter
	} else if client := pb.NewKernelClient(grpcconn.GRPCClient()); client == nil {
		return gopi.ErrBadParameter
	} else {
		this.conn = config.Conn
		this.client = client
	}

	// Success
	return nil
}

func (this *kernelclient) Close() error {
	return this.Unit.Close()
}

func (this *kernelclient) Conn() gopi.RPCClientConn {
	return this.conn
}

func (this *kernelclient) String() string {
	return "<gaffer.Kernel conn=" + fmt.Sprint(this.conn) + ">"
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION rpc.GafferKernelClient

func (this *kernelclient) Ping(ctx context.Context) error {
	this.conn.Lock()
	defer this.conn.Unlock()
	if _, err := this.client.Ping(ctx, &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *kernelclient) CreateProcess(ctx context.Context, service rpc.GafferService) (uint32, error) {
	this.conn.Lock()
	defer this.conn.Unlock()
	if id, err := this.client.CreateProcess(ctx, ProtoFromService(service)); err != nil {
		return 0, err
	} else if id == nil {
		return 0, gopi.ErrInternalAppError
	} else {
		return id.GetId(), nil
	}
}

func (this *kernelclient) Processes(ctx context.Context, id, sid uint32) ([]rpc.GafferProcess, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if processes, err := this.client.Processes(ctx, ProtoFromProcessId(id, sid)); err != nil {
		return nil, err
	} else {
		list := make([]rpc.GafferProcess, len(processes.Process))
		for i, process := range processes.Process {
			list[i] = ProtoToProcess(process)
		}
		return list, nil
	}
}
