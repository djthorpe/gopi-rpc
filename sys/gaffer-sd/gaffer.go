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
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Gaffer struct {
	Addr string
	Pool gopi.RPCClientPool
}

type gaffer struct {
	log  gopi.Logger
	addr string
	pool gopi.RPCClientPool
	stub rpc.DiscoveryClient

	event.Publisher
}

////////////////////////////////////////////////////////////////////////////////
// COMSTANTS

const (
	STUB_NAME = "gopi.Discovery"
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Gaffer) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<rpc.discovery.gaffer.Open>{ addr=%v clientpool=%v }", strconv.Quote(config.Addr), config.Pool)

	this := new(gaffer)
	this.log = logger
	this.addr = config.Addr
	this.pool = config.Pool
	this.stub = nil

	if this.addr == "" {
		return nil, gopi.ErrBadParameter
	}
	if this.pool == nil {
		return nil, gopi.ErrBadParameter
	}

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
	if stub, err := this.Stub(STUB_NAME); err != nil {
		return nil, err
	} else if records, err := stub.Lookup(service, rpc.DISCOVERY_TYPE_DB, 0); err != nil {
		return nil, err
	} else {
		return records, nil
	}
}

func (this *gaffer) EnumerateServices(ctx context.Context) ([]string, error) {
	this.log.Debug2("<rpc.discovery.gaffer.EnumerateServices>{ }")
	return nil, gopi.ErrNotImplemented
}

func (this *gaffer) ServiceInstances(service string) []gopi.RPCServiceRecord {
	this.log.Debug2("<rpc.discovery.gaffer.ServiceInstances>{ service=%v }", strconv.Quote(service))
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RETURN STUB

func (this *gaffer) Stub(service string) (rpc.DiscoveryClient, error) {
	if this.stub != nil {
		return this.stub, nil
	} else if conn, err := this.pool.ConnectAddr(this.addr, 0); err != nil {
		return nil, err
	} else if stub := this.pool.NewClient(service, conn); stub == nil {
		return nil, gopi.ErrNotFound
	} else if stub_, ok := stub.(rpc.DiscoveryClient); ok == false {
		return nil, gopi.ErrBadParameter
	} else {
		this.stub = stub_
		return this.stub, nil
	}
}
