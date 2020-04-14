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
	Process rpc.GafferProcess
	State   rpc.GafferState
	Buf     []byte
	Err     error
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewRunningEvent(process rpc.GafferProcess) *Event {
	return &Event{process, rpc.GAFFER_STATE_RUNNING, nil, nil}
}

func NewStoppedEvent(process rpc.GafferProcess, err error) *Event {
	return &Event{process, rpc.GAFFER_STATE_STOPPED, nil, err}
}

func NewBufferEvent(process rpc.GafferProcess, buf []byte, t rpc.GafferState) *Event {
	return &Event{process, t, buf, nil}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Event) String() string {
	str := "<Event " + fmt.Sprint(this.State)
	if this.Process != nil {
		str += " process_id=" + fmt.Sprint(this.Process.Id())
	}
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
