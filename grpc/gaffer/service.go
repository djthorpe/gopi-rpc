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

type GafferService struct {
	Server gopi.RPCServer
	Gaffer rpc.Gaffer
}

type gafferservice struct {
	base.Unit

	server gopi.RPCServer
	gaffer rpc.Gaffer
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (GafferService) Name() string { return "rpc/gaffer/service" }

func (config GafferService) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(gafferservice)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *gafferservice) Init(config GafferService) error {
	// Set server
	if config.Server == nil {
		return gopi.ErrBadParameter.WithPrefix("Server")
	} else {
		this.server = config.Server
	}

	// Set gaffer
	if config.Gaffer == nil {
		return gopi.ErrBadParameter.WithPrefix("Gaffer")
	} else {
		this.gaffer = config.Gaffer
	}

	// Register with server
	pb.RegisterGafferServer(this.server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return nil
}

func (this *gafferservice) Close() error {
	// Release resources
	this.server = nil
	this.gaffer = nil

	// Return success
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.RPCService

func (this *gafferservice) CancelRequests() error {
	// Do not need to cancel
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *gafferservice) String() string {
	str := "<" + this.Log.Name()
	str += " server=" + fmt.Sprint(this.server)
	str += " gaffer=" + fmt.Sprint(this.gaffer)
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.GafferService

func (this *gafferservice) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.Log.Debug("<Ping>")

	return &empty.Empty{}, nil
}

func (this *gafferservice) Services(context.Context, *empty.Empty) (*pb.ServiceList, error) {
	this.Log.Debug("<Services>")

	return ProtoFromServiceList(this.gaffer.Services()), nil
}

func (this *gafferservice) Update(_ context.Context, req *pb.ServiceUpdateRequest) (*pb.ServiceList, error) {
	this.Log.Debug("<Update req=", req, ">")

	if service := ProtoToService(req.Service); service == nil {
		return nil, gopi.ErrBadParameter.WithPrefix("service")
	} else if fields := req.Fields.Paths; len(fields) == 0 {
		return nil, gopi.ErrBadParameter.WithPrefix("fields")
	} else if service, err := this.gaffer.Update(service, fields); err != nil {
		return nil, err
	} else {
		return ProtoFromServiceListOne(service), nil
	}
}

func (this *gafferservice) Start(ctx context.Context, req *pb.ServiceId) (*pb.ServiceList, error) {
	this.Log.Debug("<Start req=", req, ">")

	if process, err := this.gaffer.Start(ctx, req.Sid); err != nil {
		return nil, err
	} else {
		return ProtoFromServiceListOne(process.Service()), nil
	}
}
