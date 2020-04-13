/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	"fmt"
	"strconv"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Event struct {
	State rpc.GafferState
	Buf   []byte
	Err   error
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewRunningEvent() *Event {
	return &Event{rpc.GAFFER_STATE_RUNNING, nil, nil}
}

func NewStoppedEvent(err error) *Event {
	return &Event{rpc.GAFFER_STATE_STOPPED, nil, err}
}

func NewBufferEvent(buf []byte, t rpc.GafferState) *Event {
	return &Event{t, buf, nil}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Event) String() string {
	str := "<Event " + fmt.Sprint(this.State)
	switch this.State {
	case rpc.GAFFER_STATE_STDERR, rpc.GAFFER_STATE_STDOUT:
		str += " buf=" + strconv.Quote(string(this.Buf))
	case rpc.GAFFER_STATE_STOPPED:
		if this.Err != nil {
			str += " err=" + strconv.Quote(fmt.Sprint(this.Err))
		}
	}
	return str + ">"
}
