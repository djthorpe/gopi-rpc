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

type Util struct {
	// No members
}

type util struct {
	// No members
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open a client
func (config Util) Open(log gopi.Logger) (gopi.Driver, error) {
	return new(util), nil
}

func (this *util) Close() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *util) String() string {
	return "<rpc.util>{}"
}
