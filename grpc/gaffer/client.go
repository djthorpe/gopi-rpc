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
	"io"

	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/genproto/protobuf/field_mask"

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

type GafferClient struct {
	Conn gopi.RPCClientConn
}

type kernelclient struct {
	base.Unit
	conn   gopi.RPCClientConn
	client pb.KernelClient
}

type gafferclient struct {
	base.Unit
	conn   gopi.RPCClientConn
	client pb.GafferClient
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit for Kernel client

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
// IMPLEMENTATION gopi.Unit for Gaffer client

func (GafferClient) Name() string { return "gaffer.Gaffer" }

func (config GafferClient) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(gafferclient)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *gafferclient) Init(config GafferClient) error {
	// Create the client
	if config.Conn == nil {
		return gopi.ErrBadParameter
	} else if grpcconn, ok := config.Conn.(grpc.GRPCClientConn); ok == false {
		return gopi.ErrBadParameter
	} else if client := pb.NewGafferClient(grpcconn.GRPCClient()); client == nil {
		return gopi.ErrBadParameter
	} else {
		this.conn = config.Conn
		this.client = client
	}

	// Success
	return nil
}

func (this *gafferclient) Close() error {
	return this.Unit.Close()
}

func (this *gafferclient) Conn() gopi.RPCClientConn {
	return this.conn
}

func (this *gafferclient) String() string {
	return "<gaffer.Gaffer conn=" + fmt.Sprint(this.conn) + ">"
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
		return 0, gopi.ErrInternalAppError.WithPrefix("CreateProcess")
	} else {
		return id.GetId(), nil
	}
}

func (this *kernelclient) Processes(ctx context.Context, id uint32) ([]rpc.GafferProcess, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if processes, err := this.client.Processes(ctx, ProtoFromProcessId(id)); err != nil {
		return nil, err
	} else {
		list := make([]rpc.GafferProcess, len(processes.Process))
		for i, process := range processes.Process {
			list[i] = ProtoToProcess(process)
		}
		return list, nil
	}
}

func (this *kernelclient) Executables(ctx context.Context) ([]string, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if executables, err := this.client.Executables(ctx, &empty.Empty{}); err != nil {
		return nil, err
	} else {
		list := make([]string, len(executables.Executable))
		for i, exec := range executables.Executable {
			list[i] = exec
		}
		return list, nil
	}
}

func (this *kernelclient) RunProcess(ctx context.Context, id uint32) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	if _, err := this.client.RunProcess(ctx, ProtoFromProcessId(id)); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *kernelclient) StopProcess(ctx context.Context, id uint32) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	if _, err := this.client.StopProcess(ctx, ProtoFromProcessId(id)); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *kernelclient) StreamEvents(ctx context.Context, id uint32) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	stream, err := this.client.StreamEvents(ctx, ProtoFromProcessId(id))
	if err != nil {
		return err
	}

	// Errors channel receives errors from recv
	errors := make(chan error)

	// Receive messages in the background
	go func() {
	FOR_LOOP:
		for {
			if evt_, err := stream.Recv(); err == io.EOF {
				break FOR_LOOP
			} else if err != nil {
				errors <- err
				break FOR_LOOP
			} else if evt_.State == pb.KernelProcessEvent_NONE {
				// Null event - mark timestamp
				//ts = time.Now()
			} else {
				fmt.Println("*pb.KernelProcessEvent", evt_)
			}
		}
		close(errors)
	}()

	// Continue until error or io.EOF is returned
	for {
		select {
		case <-ctx.Done():
			if err := <-errors; err == nil || grpc.IsErrCanceled(err) {
				return nil
			} else {
				return err
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION rpc.GafferClient

func (this *gafferclient) Mutable(service rpc.GafferService) rpc.MutableGafferService {
	return NewMutable(service)
}

func (this *gafferclient) Ping(ctx context.Context) error {
	this.conn.Lock()
	defer this.conn.Unlock()
	if _, err := this.client.Ping(ctx, &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *gafferclient) Services(ctx context.Context) ([]rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()
	if services_, err := this.client.Services(ctx, &empty.Empty{}); err != nil {
		return nil, err
	} else {
		services := make([]rpc.GafferService, len(services_.Service))
		for i, pb := range services_.Service {
			services[i] = ProtoToService(pb)
		}
		return services, nil
	}
}

func (this *gafferclient) Update(ctx context.Context, service rpc.MutableGafferService) (rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if proto, ok := service.(*pbservice); ok == false {
		return nil, gopi.ErrInternalAppError.WithPrefix("service")
	} else if services_, err := this.client.Update(ctx, &pb.ServiceUpdateRequest{
		Service: proto.proto,
		Fields: &field_mask.FieldMask{
			Paths: proto.Fields(),
		},
	}); err != nil {
		return nil, err
	} else if len(services_.Service) != 1 {
		return nil, gopi.ErrInternalAppError
	} else {
		return ProtoToService(services_.Service[0]), nil
	}
}

func (this *gafferclient) Start(ctx context.Context, sid uint32) (rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if services, err := this.client.Start(ctx, ProtoFromServiceId(sid)); err != nil {
		return nil, err
	} else if len(services.Service) != 1 {
		return nil, gopi.ErrInternalAppError
	} else {
		return ProtoToService(services.Service[0]), nil
	}
}
