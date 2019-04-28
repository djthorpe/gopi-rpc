/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"context"
	"fmt"
	"strconv"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/discovery"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server    gopi.RPCServer
	Discovery gopi.RPCServiceDiscovery
}

type service struct {
	log       gopi.Logger
	discovery gopi.RPCServiceDiscovery
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.discovery>Open{ server=%v discovery=%v }", config.Server, config.Discovery)

	if config.Server == nil || config.Discovery == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(service)
	this.log = log
	this.discovery = config.Discovery

	// Register service with GRPC server
	pb.RegisterDiscoveryServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return this, nil
}

func (this *service) Close() error {
	this.log.Debug("<grpc.service.discovery>Close{}")

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RPCService implementation

func (this *service) CancelRequests() error {
	// No need to cancel any requests since none are streaming
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Stringify

func (this *service) String() string {
	return fmt.Sprintf("grpc.service.discovery{}")
}

////////////////////////////////////////////////////////////////////////////////
// RPC Methods

func (this *service) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.discovery.Ping>{ }")
	return &empty.Empty{}, nil
}

func (this *service) Register(ctx context.Context, service *pb.ServiceRecord) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.discovery.Register>{ service=%v }", service)
	if service == nil {
		return nil, gopi.ErrNotImplemented
	} else if err := this.discovery.Register(protoToServiceRecord(service)); err != nil {
		return nil, err
	} else {
		return &empty.Empty{}, nil
	}
}

func (this *service) Enumerate(ctx context.Context, req *pb.EnumerateRequest) (*pb.EnumerateReply, error) {
	this.log.Debug("<grpc.service.discovery.Enumerate>{ type=%v }", req.Type)

	// Adjust the deadline to allow for communicating back to the client
	if timeout, ok := ctx.Deadline(); ok {
		deadline := timeout.Sub(time.Now()) - 100*time.Millisecond
		if deadline > 0 {
			ctx, _ = context.WithTimeout(ctx, deadline)
		}
	}

	// Enumerate using DNS
	if req.Type == pb.DiscoveryType_DISCOVERY_DNS {
		if services, err := this.discovery.EnumerateServices(ctx); err != nil {
			return nil, err
		} else {
			return &pb.EnumerateReply{Service: services}, nil
		}
	}

	// Enumerate using Cache
	if req.Type == pb.DiscoveryType_DISCOVERY_DB {
		services := make(map[string]bool)
		for _, instance := range this.discovery.ServiceInstances("") {
			services[instance.Service()] = true
		}
		services_ := make([]string, 0, len(services))
		for key := range services {
			services_ = append(services_, key)
		}
		return &pb.EnumerateReply{Service: services_}, nil
	}

	// Return error
	return nil, gopi.ErrBadParameter
}

func (this *service) Lookup(ctx context.Context, req *pb.LookupRequest) (*pb.LookupReply, error) {
	this.log.Debug("<grpc.service.discovery.Lookup>{ type=%v service=%v }", req.Type, strconv.Quote(req.Service))

	// Check incoming parameters
	if req.Service == "" {
		return nil, gopi.ErrBadParameter
	}

	// Adjust the deadline to allow for communicating back to the client
	if timeout, ok := ctx.Deadline(); ok {
		deadline := timeout.Sub(time.Now()) - 100*time.Millisecond
		if deadline > 0 {
			ctx, _ = context.WithTimeout(ctx, deadline)
		}
	}

	if req.Type == pb.DiscoveryType_DISCOVERY_DNS {
		if reply, err := this.discovery.Lookup(ctx, req.Service); err != nil {
			return nil, err
		} else {
			return &pb.LookupReply{
				Service: protoFromServiceRecords(reply),
			}, nil
		}
	}

	if req.Type == pb.DiscoveryType_DISCOVERY_DB {
		return &pb.LookupReply{
			Service: protoFromServiceRecords(this.discovery.ServiceInstances(req.Service)),
		}, nil
	}
	// Return error
	return nil, gopi.ErrBadParameter
}
