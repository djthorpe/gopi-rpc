/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"fmt"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Event struct {
	Source_   gopi.Driver
	Type_     rpc.GafferEventType
	Service_  rpc.GafferService
	Group_    rpc.GafferServiceGroup
	Instance_ rpc.GafferServiceInstance
	Data_     []byte
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func NewEventWithService(source gopi.Driver, type_ rpc.GafferEventType, service rpc.GafferService) *Event {
	this := new(Event)
	this.Source_ = source
	this.Type_ = type_
	this.Service_ = service
	return this
}

func NewEventWithGroup(source gopi.Driver, type_ rpc.GafferEventType, group rpc.GafferServiceGroup) *Event {
	this := new(Event)
	this.Source_ = source
	this.Type_ = type_
	this.Group_ = group
	return this
}

func NewEventWithInstanceData(source gopi.Driver, type_ rpc.GafferEventType, instance rpc.GafferServiceInstance, data []byte) *Event {
	this := new(Event)
	this.Source_ = source
	this.Type_ = type_
	this.Instance_ = instance
	this.Data_ = data
	return this
}

func NewEventWithInstance(source gopi.Driver, type_ rpc.GafferEventType, instance rpc.GafferServiceInstance) *Event {
	return NewEventWithInstanceData(source, type_, instance, nil)
}

////////////////////////////////////////////////////////////////////////////////
// GOPI EVENT IMPLEMENTATION

func (this *Event) Source() gopi.Driver {
	return this.Source_
}

func (this *Event) Name() string {
	return "GafferEvent"
}

func (this *Event) Type() rpc.GafferEventType {
	return this.Type_
}

func (this *Event) Service() rpc.GafferService {
	if this.Service_ != nil {
		return this.Service_
	} else if this.Instance_ != nil {
		return this.Instance_.Service()
	} else {
		return nil
	}
}

func (this *Event) Group() rpc.GafferServiceGroup {
	return this.Group_
}

func (this *Event) Instance() rpc.GafferServiceInstance {
	return this.Instance_
}

func (this *Event) Data() []byte {
	return this.Data_
}

func (this *Event) String() string {
	if this.Service_ != nil {
		return fmt.Sprintf("<%v>{ %v %v }", this.Name(), this.Type_, this.Service_)
	} else if this.Group_ != nil {
		return fmt.Sprintf("<%v>{ %v %v }", this.Name(), this.Type_, this.Group_)
	} else if this.Instance_ != nil {
		if this.Data_ != nil {
			log := strings.TrimSuffix(string(this.Data_), "\n")
			return fmt.Sprintf("<%v>{ %v %v %v }", this.Name(), this.Type_, this.Instance_, strconv.Quote(log))
		} else {
			return fmt.Sprintf("<%v>{ %v %v }", this.Name(), this.Type_, this.Instance_)
		}
	} else {
		return fmt.Sprintf("<%v>{ %v }", this.Name(), this.Type_)
	}
}
