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
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type EventType uint

type Event struct {
	Type EventType
	Buf  []byte
	Err  error
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	EVENT_TYPE_NONE EventType = iota
	EVENT_TYPE_STDOUT
	EVENT_TYPE_STDERR
	EVENT_TYPE_STOPPED
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewStoppedEvent(err error) *Event {
	return &Event{EVENT_TYPE_STOPPED, nil, err}
}

func NewBufferEvent(buf []byte, t EventType) *Event {
	return &Event{t, buf, nil}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Event) String() string {
	str := "<Event " + fmt.Sprint(this.Type)
	switch this.Type {
	case EVENT_TYPE_STDERR, EVENT_TYPE_STDOUT:
		str += " buf=" + strconv.Quote(string(this.Buf))
	case EVENT_TYPE_STOPPED:
		if this.Err != nil {
			str += " err=" + strconv.Quote(fmt.Sprint(this.Err))
		}
	}
	return str + ">"
}

func (t EventType) String() string {
	switch t {
	case EVENT_TYPE_NONE:
		return "EVENT_TYPE_NONE"
	case EVENT_TYPE_STDOUT:
		return "EVENT_TYPE_STDOUT"
	case EVENT_TYPE_STDERR:
		return "EVENT_TYPE_STDERR"
	case EVENT_TYPE_STOPPED:
		return "EVENT_TYPE_STOPPED"
	default:
		return "[?? Invalid EventType value]"
	}
}
