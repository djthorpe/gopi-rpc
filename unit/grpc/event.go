/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type event struct {
	source_ gopi.Unit
	type_   gopi.RPCEventType
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewEvent(source_ gopi.Unit, type_ gopi.RPCEventType) gopi.RPCEvent {
	return &event{
		source_: source_,
		type_:   type_,
	}
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Event

func (*event) NS() gopi.EventNS {
	return gopi.EVENT_NS_DEFAULT
}

func (*event) Name() string {
	return "gopi.RPCEvent"
}

func (this *event) Source() gopi.Unit {
	return this.source_
}

func (this *event) Value() interface{} {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.RPCEvent

func (*event) Service() gopi.RPCServiceRecord {
	return gopi.RPCServiceRecord{}
}

func (this *event) Type() gopi.RPCEventType {
	return this.type_
}

func (this *event) TTL() time.Duration {
	return 0
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *event) String() string {
	return "<" + this.Name() +
		" type=" + fmt.Sprint(this.type_) +
		" service=" + fmt.Sprint(this.Service()) +
		">"
}
