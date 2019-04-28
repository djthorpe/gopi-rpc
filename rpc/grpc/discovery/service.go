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
	event "github.com/djthorpe/gopi/util/event"

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
	event.Publisher

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

	// Close event channel
	this.Publisher.Close()

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RPCService implementation

func (this *service) CancelRequests() error {
	this.log.Debug("<grpc.service.discovery>CancelRequests{}")

	// Put empty event onto the channel to indicate any on-going
	// requests should be ended
	this.Emit(event.NullEvent)

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

// Stream events
func (this *service) StreamEvents(req *pb.StreamEventsRequest, stream pb.Discovery_StreamEventsServer) error {
	this.log.Debug2("<grpc.service.discovery.StreamEvents>{ service=%v }", strconv.Quote(req.Service))

	// Subscribe to channel for incoming events, and continue until cancel request is received, send
	// empty events occasionally to ensure the channel is still alive
	events := this.discovery.Subscribe()
	cancel := this.Subscribe()
	ticker := time.NewTicker(time.Second)

FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt == nil {
				break FOR_LOOP
			} else if evt_, ok := evt.(gopi.RPCEvent); ok {
				// TODO: FILTER
				if err := stream.Send(protoFromEvent(evt_)); err != nil {
					this.log.Warn("StreamEvents: %v", err)
					break FOR_LOOP
				}
			} else {
				this.log.Warn("StreamEvents: Ignoring event: %v", evt)
			}
		case <-ticker.C:
			if err := stream.Send(&pb.Event{}); err != nil {
				this.log.Warn("StreamEvents: %v", err)
				break FOR_LOOP
			}
		case <-cancel:
			break FOR_LOOP
		}
	}

	// Stop ticker, unsubscribe from events
	ticker.Stop()
	this.discovery.Unsubscribe(events)
	this.Unsubscribe(cancel)

	this.log.Debug2("StreamEvents: Ended")

	// Return success
	return nil
}
