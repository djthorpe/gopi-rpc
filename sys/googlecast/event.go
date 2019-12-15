/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type castevent struct {
	source gopi.Driver
	type_  gopi.RPCEventType
	device *castdevice
	ts     time.Time
}

type channelevent struct {
	source  gopi.Driver
	type_   rpc.GoogleCastEventType
	channel rpc.GoogleCastChannel
}

////////////////////////////////////////////////////////////////////////////////
// CAST EVENT

func NewCastEvent(s gopi.Driver, t gopi.RPCEventType, d *castdevice) gopi.Event {
	return &castevent{
		s, t, d, time.Now(),
	}
}

func (this *castevent) Name() string {
	return "GoogleCastEvent"
}

func (this *castevent) Source() gopi.Driver {
	return this.source
}

func (this *castevent) Type() gopi.RPCEventType {
	return this.type_
}

func (this *castevent) Device() rpc.GoogleCastDevice {
	return this.device
}

func (this *castevent) Timestamp() time.Time {
	return this.ts
}

func (this *castevent) String() string {
	return fmt.Sprintf("<%v>{ type=%v device=%v}", this.Name(), this.Type(), this.Device())
}

////////////////////////////////////////////////////////////////////////////////
// CHANNEL EVENT

func NewChannelEvent(source gopi.Driver, t rpc.GoogleCastEventType, channel rpc.GoogleCastChannel) gopi.Event {
	return &channelevent{source, t, channel}
}

func (this *channelevent) Name() string {
	return "GoogleChannelEvent"
}

func (this *channelevent) Source() gopi.Driver {
	return this.source
}

func (this *channelevent) Type() rpc.GoogleCastEventType {
	return this.type_
}

func (this *channelevent) Channel() rpc.GoogleCastChannel {
	return this.channel
}

func (this *channelevent) String() string {
	return fmt.Sprintf("<%v>{ type=%v channel=%v}", this.Name(), this.Type(), this.Channel())
}
