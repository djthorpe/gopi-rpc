/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"context"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
)

/////////////////////////////////////////////////////////////////////
// INTERFACES

// GRPCServer interface is an RPCServer which also
// returns gRPC-specific properties
type GRPCServer interface {
	gopi.RPCServer

	// Return the gRPC Server object
	GRPCServer() *grpc.Server
}

// GRPCClientConn interface is an RPCClientConn which also
// returns gRPC-specific properties
type GRPCClientConn interface {
	gopi.RPCClientConn

	// Return the gRPC Server object
	GRPCClient() grpc.ClientConnInterface
}

/////////////////////////////////////////////////////////////////////
// UTILITY FUNCTIONS

func IsErrCanceled(err error) bool {
	if err == nil {
		return false
	}
	if err == context.Canceled {
		return true
	}
	return grpc.Code(err) == codes.Canceled
}

func IsErrUnavailable(err error) bool {
	if err == nil {
		return false
	}
	return grpc.Code(err) == codes.Unavailable
}

func IsErrDeadlineExceeded(err error) bool {
	if err == nil {
		return false
	}
	if err == context.DeadlineExceeded {
		return true
	}
	return grpc.Code(err) == codes.DeadlineExceeded
}
