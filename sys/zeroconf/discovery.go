/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2017
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package zeroconf

import (
	"context"
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi/sys/rpc"
	event "github.com/djthorpe/gopi/util/event"
	mdns "github.com/djthorpe/zeroconf"
)

////////////////////////////////////////////////////////////////////////////////
// STRUCTS

// The configuration
type Config struct {
	Domain string
}

// The driver for the logging
type driver struct {
	log      gopi.Logger
	domain   string
	servers  []*mdns.Server
	resolver *mdns.Resolver
	event.Publisher
}

///////////////////////////////////////////////////////////////////////////////
// CONSTS

const (
	MDNS_DEFAULT_DOMAIN = "local."
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Create discovery object
func (config Config) Open(log gopi.Logger) (gopi.Driver, error) {

	this := new(driver)
	this.log = log
	if config.Domain == "" {
		this.domain = MDNS_DEFAULT_DOMAIN
	} else {
		this.domain = config.Domain
	}

	log.Debug("<sys.rpc.mdns>Open{ domain='%v' }", this.domain)

	this.servers = make([]*mdns.Server, 0, 1)

	if resolver, err := mdns.NewResolver(); err != nil {
		return nil, err
	} else {
		this.resolver = resolver
	}

	// success
	return this, nil
}

// Close discovery object
func (this *driver) Close() error {
	this.log.Debug("<sys.rpc.mdns>Close{ domain='%v' }", this.domain)

	// Close servers
	for _, server := range this.servers {
		server.Shutdown()
	}

	// Unsubscribe
	this.Publisher.Close()

	// Empty methods
	this.servers = nil
	this.resolver = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// DRIVER INTERFACE METHODS

// Register a service and announce the service when queries occur
func (this *driver) Register(service *gopi.RPCServiceRecord) error {
	this.log.Debug2("<sys.rpc.mdns>Register{ service=%v }", service)
	if server, err := mdns.Register(service.Name, service.Type, this.domain, int(service.Port), service.Text, nil); err != nil {
		return err
	} else {
		this.servers = append(this.servers, server)
		return nil
	}
}

// Browse will find service entries, will block until ctx timeout
// or cancel
func (this *driver) Browse(ctx context.Context, serviceType string) error {
	this.log.Debug2("<sys.rpc.mdns>Browse{ service_type='%v' }", serviceType)
	entries := make(chan *mdns.ServiceEntry)
	if err := this.resolver.Browse(ctx, serviceType, this.domain, entries); err != nil {
		return err
	} else {
		for entry := range entries {
			this.Emit(&gopi.RPCServiceRecord{
				Name: entry.Instance,
				Type: entry.Service,
				Port: uint(entry.Port),
				Text: entry.Text,
				Host: entry.HostName,
				IP4:  entry.AddrIPv4,
				IP6:  entry.AddrIPv6,
				TTL:  time.Duration(entry.TTL) * time.Second,
			})
		}
		return nil
	}
}

func (this *driver) DefaultServiceType(network string) string {
	return "_gopi._" + network
}

////////////////////////////////////////////////////////////////////////////////
// PUBSUB

// Emit an event
func (this *driver) Emit(record *gopi.RPCServiceRecord) {
	this.Publisher.Emit(rpc.NewEvent(this, gopi.RPC_EVENT_SERVICE_RECORD, record))
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *driver) String() string {
	return fmt.Sprintf("<sys.rpc.mdns>{ domain=\"%v\" registrations=%v }", this.domain, this.servers)
}
