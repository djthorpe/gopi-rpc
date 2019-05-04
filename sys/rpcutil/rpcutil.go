/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpcutil

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

type RPCUtil struct {
	// No members
}

type rpcutil struct {
	// No members
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open a client
func (config RPCUtil) Open(log gopi.Logger) (gopi.Driver, error) {
	return new(rpcutil), nil
}

func (this *rpcutil) Close() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *rpcutil) String() string {
	return "<rpc.util>{}"
}
