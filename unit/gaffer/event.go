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
	process rpc.GafferProcess
	state   rpc.GafferState
	buf     []byte
	err     error
}

////////////////////////////////////////////////////////////////////////////////
// NEW

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
// PUBLIC METHODS

func (this *Event) Process() rpc.GafferProcess {
	return this.process
}

func (this *Event) State() rpc.GafferState {
	return this.state
}

func (this *Event) Buf() []byte {
	return this.buf
}

func (this *Event) Error() error {
	return this.err
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Event) String() string {
	str := "<Event " + fmt.Sprint(this.state)
	if this.Process != nil {
		str += " process_id=" + fmt.Sprint(this.process.Id())
	}
	switch this.state {
	case rpc.GAFFER_STATE_STDERR, rpc.GAFFER_STATE_STDOUT:
		str += " buf=" + strconv.Quote(string(this.buf))
	case rpc.GAFFER_STATE_STOPPED:
		if this.err != nil {
			str += " err=" + strconv.Quote(fmt.Sprint(this.err))
		}
	}
	return str + ">"
}
