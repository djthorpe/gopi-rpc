/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package helloworld

import (
	// Frameworks

	rpc "github.com/djthorpe/gopi-rpc/v2"
	base "github.com/djthorpe/gopi/v2/base"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server rpc.Server
}

type service struct {
	base.Unit
}
