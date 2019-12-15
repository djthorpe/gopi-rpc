/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"context"
	"io"
	"net"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/miekg/dns"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	// DiscoveryType is either DNS (using DNS-SD) or DB (using internal database)
	DiscoveryType       uint
	GoogleCastEventType uint
)

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

const (
	GOOGLE_CAST_EVENT_NONE    GoogleCastEventType = 0
	GOOGLE_CAST_EVENT_CONNECT GoogleCastEventType = iota
	GOOGLE_CAST_EVENT_DISCONNECT
	GOOGLE_CAST_EVENT_DEVICE
	GOOGLE_CAST_EVENT_VOLUME
	GOOGLE_CAST_EVENT_APPLICATION
	GOOGLE_CAST_EVENT_MEDIA
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

	// Set parameters
	SetService(service, subtype string) error
	SetName(name string) error
	SetHostPort(addr string) error
	SetTTL(time.Duration) error
	AppendIP(...net.IP) error
	AppendTXT(...string) error

	// Get DNS answers
	PTR(zone string, ttl time.Duration) *dns.PTR
	SRV(zone string, ttl time.Duration) *dns.SRV
	TXT(zone string, ttl time.Duration) *dns.TXT
	A(zone string, ttl time.Duration) []*dns.A
	AAAA(zone string, ttl time.Duration) []*dns.AAAA

	// Set parameters from DNS answers
	SetPTR(zone string, rr *dns.PTR) error
	SetSRV(zone string, rr *dns.SRV) error
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
	gopi.Publisher

	// Ping the remote service instance
	Ping() error

	// Register a service record in database
	Register(gopi.RPCServiceRecord) error

	// Enumerate service names
	Enumerate(DiscoveryType, time.Duration) ([]string, error)

	// Lookup service instances
	Lookup(string, DiscoveryType, time.Duration) ([]gopi.RPCServiceRecord, error)

	// Stream discovery events. filtering by service name
	StreamEvents(ctx context.Context, service string) error
}

////////////////////////////////////////////////////////////////////////////////
// GOOGLECAST

type GoogleCast interface {
	gopi.Driver
	gopi.Publisher

	Devices() []GoogleCastDevice

	// Connect to the control channel for a device, with timeout
	Connect(GoogleCastDevice, gopi.RPCFlag, time.Duration) (GoogleCastChannel, error)
	Disconnect(GoogleCastChannel) error
}

type GoogleCastDevice interface {
	Id() string
	Name() string
	Model() string
	Service() string
	State() uint
}

type GoogleCastChannel interface {
	// Address of channel
	RemoteAddr() string

	// Get Properties
	Applications() []GoogleCastApplication
	Volume() GoogleCastVolume
	Media() GoogleCastMedia

	// Set Properties
	SetApplication(GoogleCastApplication) error // Application to watch or nil
	SetPlay(bool) (int, error)                  // Play or stop
	SetPause(bool) (int, error)                 // Pause or play
	SetVolume(float32) (int, error)             // Set volume level
	SetMuted(bool) (int, error)                 // Set muted
	//SetTrackNext() (int, error)
	//SetTrackPrev() (int, error)
	//StreamUrl(string)
}

type GoogleCastApplication interface {
	ID() string
	Name() string
	Status() string
}

type GoogleCastVolume interface {
	Level() float32
	Muted() bool
}

type GoogleCastMedia interface {
}

type GoogleCastEvent interface {
	gopi.Event

	Type() gopi.RPCEventType
	Device() GoogleCastDevice
}

type GoogleChannelEvent interface {
	gopi.Event

	Type() GoogleCastEventType
	Channel() GoogleCastChannel
}

type GoogleCastClient interface {
	gopi.RPCClient
	gopi.Publisher

	// Ping remote service
	Ping() error

	// Return devices from the remote service
	Devices() ([]GoogleCastDevice, error)

	// Stream discovery events
	StreamEvents(ctx context.Context) error
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t GoogleCastEventType) String() string {
	switch t {
	case GOOGLE_CAST_EVENT_NONE:
		return "GOOGLE_CAST_EVENT_NONE"
	case GOOGLE_CAST_EVENT_CONNECT:
		return "GOOGLE_CAST_EVENT_CONNECT"
	case GOOGLE_CAST_EVENT_DISCONNECT:
		return "GOOGLE_CAST_EVENT_DISCONNECT"
	case GOOGLE_CAST_EVENT_DEVICE:
		return "GOOGLE_CAST_EVENT_DEVICE"
	case GOOGLE_CAST_EVENT_VOLUME:
		return "GOOGLE_CAST_EVENT_VOLUME"
	case GOOGLE_CAST_EVENT_APPLICATION:
		return "GOOGLE_CAST_EVENT_APPLICATION"
	case GOOGLE_CAST_EVENT_MEDIA:
		return "GOOGLE_CAST_EVENT_MEDIA"
	default:
		return "[?? Invalid GoogleCastEventType value]"
	}
}
