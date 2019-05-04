/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
)

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

type RPCUtil interface {
	gopi.Driver

	// NewEvent creates a new event from source, type and service record
	NewEvent(gopi.Driver, gopi.RPCEventType, gopi.RPCServiceRecord) gopi.RPCEvent

	// NewServiceRecord creates an empty service record
	NewServiceRecord() gopi.RPCServiceRecord
}

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
