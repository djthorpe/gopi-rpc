/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"context"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// GafferKernelClient represents a connection to a remote kernel service
type HelloworldStub interface {
	// Ping returns nil if the remote service is running
	Ping(context.Context) error
}
