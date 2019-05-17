/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"io"
	"net"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/miekg/dns"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// DiscoveryType is either DNS (using DNS-SD) or DB (using internal database)
type DiscoveryType uint

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DISCOVERY_TYPE_NONE DiscoveryType = 0
	DISCOVERY_TYPE_DNS  DiscoveryType = 1
	DISCOVERY_TYPE_DB   DiscoveryType = 2
)

const (
	DISCOVERY_SERVICE_QUERY = "_services._dns-sd._udp"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

type Util interface {
	gopi.Driver

	// NewEvent creates a new event from source, type and service record
	NewEvent(gopi.Driver, gopi.RPCEventType, gopi.RPCServiceRecord) gopi.RPCEvent

	// NewServiceRecord creates an empty service record
	NewServiceRecord(source DiscoveryType) ServiceRecord

	// Read and write array of service records
	Writer(fh io.Writer, records []ServiceRecord, indent bool) error
	Reader(fh io.Reader) ([]ServiceRecord, error)
}

type ServiceRecord interface {
	gopi.RPCServiceRecord

	// Key returns the PTR record for the service record
	Key() string

	// Expired returns true if TTL has been reached
	Expired() bool

	// Source returns the source of the record
	Source() DiscoveryType

	// Get DNS answers
	PTR(zone string, ttl time.Duration) *dns.PTR
	SRV(zone string, ttl time.Duration) *dns.SRV
	TXT(zone string, ttl time.Duration) *dns.TXT
	A(zone string, ttl time.Duration) []*dns.A
	AAAA(zone string, ttl time.Duration) []*dns.AAAA

	// Set parameters
	SetService(service, subtype string) error
	SetName(name string) error
	SetAddr(addr string) error
	SetPTR(zone string, rr *dns.PTR) error
	SetSRV(zone string, rr *dns.SRV) error
	SetTTL(time.Duration) error
	AppendIP(...net.IP) error
	AppendTXT(...string) error
}

////////////////////////////////////////////////////////////////////////////////
// RPC CLIENTS

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
// GOOGLECAST

type GoogleCast interface {
	gopi.Driver
	gopi.Publisher

	Devices() []GoogleCastDevice
}

type GoogleCastDevice interface {
	Id() string
	Name() string
	Model() string
	Service() string
	State() uint
}

type GoogleCastEvent interface {
	gopi.Event

	Type() gopi.RPCEventType
	Device() GoogleCastDevice
	Timestamp() time.Time
}

type GoogleCastClient interface {
	// Ping remote service
	Ping() error

	// Return devices from the remote service
	Devices() ([]GoogleCastDevice, error)

	// Stream discovery events
	StreamEvents(string, chan<- GoogleCastEvent) error
}
