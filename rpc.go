/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"github.com/djthorpe/gopi/v2"
)

type Server interface {
	// Start the RPC server, blocks until completed
	Start() error

	// Stop signals end to server, and force end
	Stop(halt bool) error

	// Implements gopi.Unit
	gopi.Unit
}
