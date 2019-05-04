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
	"fmt"

	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// event is the implementation of gopi.RPCEvent
type event struct {
	s gopi.Driver
	t gopi.RPCEventType
	r gopi.RPCServiceRecord
}

////////////////////////////////////////////////////////////////////////////////
// EVENT IMPLEMENTATION

func (this *rpcutil) NewEvent(source gopi.Driver, type_ gopi.RPCEventType, service gopi.RPCServiceRecord) gopi.RPCEvent {
	return &event{s: source, t: type_, r: service}
}

// Return the type of event
func (this *event) Type() gopi.RPCEventType {
	return this.t
}

// Return the service record
func (this *event) ServiceRecord() gopi.RPCServiceRecord {
	return this.r
}

// Return name of event
func (*event) Name() string {
	return "RPCEvent"
}

// Return source of event
func (this *event) Source() gopi.Driver {
	return this.s
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *event) String() string {
	if this.r != nil {
		return fmt.Sprintf("<rpc.event>{ type=%v record=%v }", this.t, this.r)
	} else {
		return fmt.Sprintf("<rpc.event>{ type=%v }", this.t)
	}
}
