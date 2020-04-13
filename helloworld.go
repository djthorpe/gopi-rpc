/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"context"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// HelloworldStub represents a connection to a remote helloworld service
type HelloworldStub interface {
	gopi.RPCClientStub

	// Ping returns without error if the remote service is running
	Ping(context.Context) error
}
