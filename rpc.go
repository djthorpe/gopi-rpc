/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"fmt"
	"regexp"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
)

// RPCEvent implementation
type Event struct {
	s gopi.Driver
	t gopi.RPCEventType
	r gopi.RPCServiceRecord
}

// DiscoveryType is either DNS (using DNS-SD) or DB (using internal database)
type DiscoveryType uint

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DISCOVERY_TYPE_NONE DiscoveryType = 0
	DISCOVERY_TYPE_DNS  DiscoveryType = 1
	DISCOVERY_TYPE_DB   DiscoveryType = 2
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

type GreeterClient interface {
	gopi.RPCClient

	// Ping the remote service instance
	Ping() error

	// Return a message from the remote service
	SayHello(name string) (string, error)
}

type VersionClient interface {
	gopi.RPCClient

	// Ping the remote service instance
	Ping() error

	// Return version parameters from the remote service
	Version() (map[string]string, error)
}

type DiscoveryClient interface {
	gopi.RPCClient

	// Ping the remote service instance
	Ping() error

	// Register a service record
	Register(gopi.RPCServiceRecord) error

	// Enumerate service names
	Enumerate(DiscoveryType, time.Duration) ([]string, error)

	// Lookup service instances
	Lookup(string, DiscoveryType, time.Duration) ([]gopi.RPCServiceRecord, error)

	// Stream discovery events. filtering by service name
	StreamEvents(string, chan<- gopi.RPCEvent) error
}

////////////////////////////////////////////////////////////////////////////////
// EVENT IMPLEMENTATION

func NewEvent(source gopi.Driver, type_ gopi.RPCEventType, service gopi.RPCServiceRecord) *Event {
	return &Event{s: source, t: type_, r: service}
}

// Return the type of event
func (this *Event) Type() gopi.RPCEventType {
	return this.t
}

// Return the service record
func (this *Event) ServiceRecord() gopi.RPCServiceRecord {
	return this.r
}

// Return name of event
func (*Event) Name() string {
	return "RPCEvent"
}

// Return source of event
func (this *Event) Source() gopi.Driver {
	return this.s
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Event) String() string {
	if this.r != nil {
		return fmt.Sprintf("<rpc.event>{ type=%v record=%v }", this.t, this.r)
	} else {
		return fmt.Sprintf("<rpc.event>{ type=%v }", this.t)
	}
}

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	reService = regexp.MustCompile("[A-za-z][A-Za-z0-9\\-]*")
)

////////////////////////////////////////////////////////////////////////////////

// RPCServiceType returns a service type from a name and
// optional subtype
func RPCServiceType(name, subtype string, flags gopi.RPCFlag) (string, error) {
	if reService.MatchString(name) == false {
		return "", gopi.ErrBadParameter
	}
	if flags&gopi.RPC_FLAG_INET_UDP != 0 {
		name = "_" + name + "._udp"
	} else {
		name = "_" + name + "._tcp"
	}
	if subtype == "" {
		return name, nil
	}
	if reService.MatchString(subtype) == false {
		return "", gopi.ErrBadParameter
	}
	return "_" + subtype + "._sub." + name, nil
}
