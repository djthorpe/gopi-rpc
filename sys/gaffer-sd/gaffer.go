/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	"context"
	"fmt"
	"strconv"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// CONFIG

type Gaffer struct {
	Addr string
}

type gaffer struct {
	log  gopi.Logger
	addr string

	event.Publisher
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Gaffer) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<rpc.discovery.gaffer.Open>{ addr=%v }", strconv.Quote(config.Addr))

	this := new(gaffer)
	this.log = logger
	this.addr = config.Addr

	// Success
	return this, nil
}

func (this *gaffer) Close() error {
	this.log.Debug("<rpc.discovery.gaffer.Close>{ addr=%v }", strconv.Quote(this.addr))

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *gaffer) String() string {
	return fmt.Sprintf("<rpc.discovery.gaffer>{ addr=%v }", strconv.Quote(this.addr))
}

////////////////////////////////////////////////////////////////////////////////
// RPCServiceDiscovery implementation

func (this *gaffer) Register(r gopi.RPCServiceRecord) error {
	this.log.Debug2("<rpc.discovery.gaffer.Register>{ r=%v }", r)
	return gopi.ErrNotImplemented
}

func (this *gaffer) Lookup(ctx context.Context, service string) ([]gopi.RPCServiceRecord, error) {
	this.log.Debug2("<rpc.discovery.gaffer.Lookup>{ service=%v }", strconv.Quote(service))
	return nil, gopi.ErrNotImplemented
}

func (this *gaffer) EnumerateServices(ctx context.Context) ([]string, error) {
	this.log.Debug2("<rpc.discovery.gaffer.EnumerateServices>{ }")
	return nil, gopi.ErrNotImplemented
}

func (this *gaffer) ServiceInstances(service string) []gopi.RPCServiceRecord {
	this.log.Debug2("<rpc.discovery.gaffer.ServiceInstances>{ service=%v }", strconv.Quote(service))
	return nil
}
