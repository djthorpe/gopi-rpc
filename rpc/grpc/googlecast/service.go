/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	"context"
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"
	event "github.com/djthorpe/gopi/util/event"
	"github.com/golang/protobuf/ptypes/empty"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/googlecast"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server     gopi.RPCServer
	GoogleCast rpc.GoogleCast
}

type service struct {
	event.Publisher

	log        gopi.Logger
	googlecast rpc.GoogleCast
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.googlecast>Open{ server=%v googlecast=%v }", config.Server, config.GoogleCast)

	if config.Server == nil || config.GoogleCast == nil {
		return nil, gopi.ErrBadParameter
	}

	this := new(service)
	this.log = log
	this.googlecast = config.GoogleCast

	// Register service with GRPC server
	pb.RegisterGoogleCastServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return this, nil
}

func (this *service) Close() error {
	this.log.Debug("<grpc.service.googlecast>Close{}")

	// Close event channel
	this.Publisher.Close()

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RPCService implementation

func (this *service) CancelRequests() error {
	this.log.Debug("<grpc.service.googlecast>CancelRequests{}")

	// Put empty event onto the channel to indicate any on-going
	// requests should be ended
	this.Emit(event.NullEvent)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Stringify

func (this *service) String() string {
	return fmt.Sprintf("grpc.service.googlecast{ %v }", this.googlecast)
}

////////////////////////////////////////////////////////////////////////////////
// RPC Methods

func (this *service) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.googlecast.Ping>{ }")
	return &empty.Empty{}, nil
}

func (this *service) Devices(context.Context, *empty.Empty) (*pb.GoogleCastDeviceReply, error) {
	this.log.Debug("<grpc.service.googlecast.Devices>{ }")
	if reply := toProtobufGoogleCastDeviceReply(this.googlecast.Devices()); reply != nil {
		return reply, nil
	} else {
		return nil, gopi.ErrAppError
	}
}

func (this *service) StreamEvents(_ *empty.Empty, stream pb.GoogleCast_StreamEventsServer) error {
	this.log.Debug2("<grpc.service.googlecast.StreamEvents>{ }")

	// Subscribe to channel for incoming events, and continue until cancel request is received, send
	// empty events occasionally to ensure the channel is still alive
	events := this.googlecast.Subscribe()
	cancel := this.Subscribe()
	ticker := time.NewTicker(time.Second)

FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt == nil {
				break FOR_LOOP
			} else if evt_, ok := evt.(rpc.GoogleCastEvent); ok {
				if err := stream.Send(toProtoGoogleCastEvent(evt_)); err != nil {
					this.log.Warn("StreamEvents: %v", err)
					break FOR_LOOP
				}
			} else {
				this.log.Warn("StreamEvents: Ignoring event: %v", evt)
			}
		case <-ticker.C:
			if err := stream.Send(&pb.GoogleCastEvent{}); err != nil {
				this.log.Warn("StreamEvents: %v", err)
				break FOR_LOOP
			}
		case <-cancel:
			break FOR_LOOP
		}
	}

	// Stop ticker, unsubscribe from events
	ticker.Stop()
	this.googlecast.Unsubscribe(events)
	this.Unsubscribe(cancel)

	this.log.Debug2("StreamEvents: Ended")

	// Return success
	return nil
}
