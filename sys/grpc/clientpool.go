/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2019
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/djthorpe/gopi/util/errors"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientPool struct {
	SSL        bool
	SkipVerify bool
	Timeout    time.Duration
	Util       rpc.Util
}

type clientpool struct {
	event.Publisher

	log        gopi.Logger
	discovery  gopi.RPCServiceDiscovery
	services   map[string]gopi.RPCNewClientFunc
	clients    []*clientconn
	ssl        bool
	skipverify bool
	timeout    time.Duration
	util       rpc.Util
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config ClientPool) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.clientpool>Open{ ssl=%v skipverify=%v timeout=%v }", config.SSL, config.SkipVerify, config.Timeout)

	this := new(clientpool)
	this.log = log
	this.util = config.Util
	this.ssl = config.SSL
	this.skipverify = config.SkipVerify
	this.timeout = config.Timeout
	this.services = make(map[string]gopi.RPCNewClientFunc, 10)
	this.clients = make([]*clientconn, 0)

	// Success
	return this, nil
}

func (this *clientpool) Close() error {
	this.log.Debug("<grpc.clientpool>Close{ discovery=%v }", this.discovery)

	// Unsubscribe listeners
	this.Publisher.Close()

	// Close clients
	errs := errors.CompoundError{}
	for _, client := range this.clients {
		if client.Connected() {
			errs.Add(client.Disconnect())
		}
	}

	// Release resources
	this.clients = nil
	this.discovery = nil
	this.services = nil

	// Return any errors
	return errs.ErrorOrSelf()
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *clientpool) Connect(service gopi.RPCServiceRecord, flags gopi.RPCFlag) (gopi.RPCClientConn, error) {
	this.log.Debug2("<grpc.clientpool.Connect>{ service=%v flags=%v }", service, flags)

	// Check incoming parameters
	if service == nil {
		return nil, gopi.ErrBadParameter
	}
	if flags&(gopi.RPC_FLAG_INET_V4|gopi.RPC_FLAG_INET_V6) == 0 {
		flags = gopi.RPC_FLAG_INET_V4 | gopi.RPC_FLAG_INET_V6
	}

	// Get addresses in order of preference
	if addrs, err := this.addressesFor(service, flags); err != nil {
		return nil, err
	} else if len(addrs) == 0 {
		return nil, gopi.ErrNotFound
	} else if conn, err := this.connectTo(service.Name(), addrs[0], service.Port(), this.ssl, this.skipverify, this.timeout); err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

func (this *clientpool) Disconnect(conn gopi.RPCClientConn) error {
	this.log.Debug2("<grpc.clientpool.Disconnect>{ conn=%v }", conn)
	if conn_, ok := conn.(*clientconn); ok {
		return conn_.Disconnect()
	} else {
		return gopi.ErrBadParameter
	}
}

func (this *clientpool) RegisterClient(service string, callback gopi.RPCNewClientFunc) error {
	this.log.Debug2("<grpc.clientpool.RegisterClient>{ service=%v callback=%v }", strconv.Quote(service), callback)
	if service == "" || callback == nil {
		return gopi.ErrBadParameter
	} else if _, exists := this.services[service]; exists {
		this.log.Warn("<rpc.clientpool>RegisterClient: Duplicate service: %v", service)
		return gopi.ErrBadParameter
	} else {
		this.services[service] = callback
		return nil
	}

	return gopi.ErrNotImplemented
}

func (this *clientpool) NewClient(service string, conn gopi.RPCClientConn) gopi.RPCClient {
	this.log.Debug2("<grpc.clientpool.NewClient>{ service=%v conn=%v }", strconv.Quote(service), conn)

	// Obtain the module with which to create a new client
	if callback, exists := this.services[service]; exists == false {
		this.log.Warn("<grpc.clientpool>NewClient: Not Found: %v", service)
		return nil
	} else {
		return callback(conn)
	}
}

func (this *clientpool) Lookup(ctx context.Context, service, addr string, max int) ([]gopi.RPCServiceRecord, error) {
	this.log.Debug2("<grpc.clientpool.Lookup>{ service=%v addr=%v max=%v }", strconv.Quote(service), strconv.Quote(addr), max)

	if this.discovery == nil || service == "" {
		// If there is no discovery service or the service string is empty,
		// then return the service record with the address only
		if record := this.util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB); record == nil {
			return nil, gopi.ErrBadParameter
		} else if err := record.SetAddr(addr); err != nil {
			return nil, err
		} else {
			return []gopi.RPCServiceRecord{record}, nil
		}
	} else if records, err := this.discovery.Lookup(ctx, service); err != nil {
		// TODO: Lookup should end when 'max' is reached
		// Error from lookup
		return nil, err
	} else if len(records) == 0 {
		// Return timeout error
		return nil, gopi.ErrDeadlineExceeded
	} else {
		// TODO: filter by address
		this.log.Debug("records=%v", records)
		return nil, gopi.ErrNotImplemented
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *clientpool) String() string {
	return fmt.Sprintf("<grpc.clientpool>{ discovery=%v services=%v }", this.discovery, this.services)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *clientpool) addressesFor(service gopi.RPCServiceRecord, flags gopi.RPCFlag) ([]net.IP, error) {

	// Check incoming parameters
	if service == nil {
		return nil, gopi.ErrBadParameter
	}

	// Obtain addresses from service record
	addrs := make([]net.IP, 0, len(service.IP4())+len(service.IP6()))
	if flags&gopi.RPC_FLAG_INET_V4 != 0 && len(service.IP4()) != 0 {
		addrs = append(addrs, service.IP4()...)
	}
	if flags&gopi.RPC_FLAG_INET_V6 != 0 && len(service.IP6()) != 0 {
		addrs = append(addrs, service.IP6()...)
	}

	// Where there are no addresses found then lookup by hostname
	if len(addrs) == 0 && service.Host() != "" {
		if addrs_, err := net.LookupHost(service.Host()); err != nil {
			return nil, err
		} else {
			for _, addr := range addrs_ {
				if ip := net.ParseIP(addr); ip != nil {
					if ip.To4() == nil && flags&gopi.RPC_FLAG_INET_V6 != 0 {
						addrs = append(addrs, ip)
					} else if ip.To4() != nil && flags&gopi.RPC_FLAG_INET_V4 != 0 {
						addrs = append(addrs, ip)
					}
				}
			}
		}
	}

	// Return not found or success
	if len(addrs) == 0 {
		return nil, gopi.ErrNotFound
	} else {
		return addrs, nil
	}
}

func (this *clientpool) connectTo(name string, addr net.IP, port uint, ssl, skipverify bool, timeout time.Duration) (gopi.RPCClientConn, error) {
	this.log.Debug2("<grpc.clientpool.Connect>{ name=%v addr=%v port=%v ssl=%v skipverify=%v timeout=%v }", strconv.Quote(name), addr, port, ssl, skipverify, timeout)

	if conn, err := gopi.Open(ClientConn{
		Name:       name,
		Addr:       fmt.Sprintf("[%v]:%v", addr.String(), port),
		SSL:        ssl,
		SkipVerify: skipverify,
		Timeout:    timeout,
	}, this.log); err != nil {
		return nil, err
	} else if conn_, ok := conn.(*clientconn); ok == false {
		return nil, gopi.ErrAppError
	} else if err := conn_.Connect(); err != nil {
		return nil, err
	} else {
		return conn_, nil
	}
}
