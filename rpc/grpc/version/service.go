/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package version

import (
	"context"
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/version"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server gopi.RPCServer
	Flags  *gopi.Flags
}

type service struct {
	log   gopi.Logger
	flags *gopi.Flags
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.version>Open{ server=%v flags=%v }", config.Server, config.Flags)

	this := new(service)
	this.log = log
	this.flags = config.Flags

	// Register service with GRPC server
	pb.RegisterVersionServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Success
	return this, nil
}

func (this *service) Close() error {
	this.log.Debug("<grpc.service.version>Close{}")

	// Release resources
	this.flags = nil

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
	return fmt.Sprintf("grpc.service.version{}")
}

////////////////////////////////////////////////////////////////////////////////
// RPC Methods

func (this *service) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.version.Ping>{ }")
	return &empty.Empty{}, nil
}

func (this *service) Version(context.Context, *empty.Empty) (*pb.VersionReply, error) {
	this.log.Debug("<grpc.service.version.Version>{ }")

	return &pb.VersionReply{}, nil
}
