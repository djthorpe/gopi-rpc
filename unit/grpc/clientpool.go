/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientPool struct {
	SSL        bool
	SkipVerify bool
	Timeout    time.Duration
}

type clientpool struct {
	base.Unit

	discovery       gopi.RPCServiceDiscovery
	timeout         time.Duration
	ssl, skipverify bool
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	reValidHostname = regexp.MustCompile("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$")
)

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (ClientPool) Name() string { return "gopi/grpc/client/pool" }

func (config ClientPool) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(clientpool)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *clientpool) Init(config ClientPool) error {
	// Set basic parameters
	this.ssl = config.SSL
	this.skipverify = config.SkipVerify
	this.timeout = config.Timeout

	// Success
	return nil
}

func (this *clientpool) Close() error {

	// Release resources
	this.discovery = nil

	// Return success
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *clientpool) String() string {
	if this.Closed {
		return "<" + this.Log.Name() + ">"
	} else if this.discovery != nil {
		return "<" + this.Log.Name() +
			" discovery=" + fmt.Sprint(this.discovery) +
			">"
	} else {
		return "<" + this.Log.Name() +
			" discovery=nil" +
			">"
	}
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.RPCClientPool

// Lookup service records by address, hostname or service name
func (this *clientpool) Lookup(ctx context.Context, addr string, max uint) ([]gopi.RPCServiceRecord, error) {
	if addr == "" {
		return nil, gopi.ErrBadParameter.WithPrefix("addr")
	} else if record, err := ParseIP(addr); err == nil && record.Addrs != nil {
		return []gopi.RPCServiceRecord{record}, nil
	} else if this.discovery == nil {
		return []gopi.RPCServiceRecord{}, gopi.ErrNotFound.WithPrefix(addr)
	} else {
		return this.discovery.Lookup(ctx, addr)
	}
}

// Connect to remote host by service record
func (this *clientpool) Connect(service gopi.RPCServiceRecord, flags gopi.RPCFlag) (gopi.RPCClientConn, error) {
	if addrs := SelectAddr(service, flags); len(addrs) == 0 {
		return nil, gopi.ErrNotFound.WithPrefix("service")
	} else if flags&gopi.RPC_FLAG_SERVICE_FIRST == gopi.RPC_FLAG_SERVICE_FIRST {
		return this.ConnectAddr(addrs[0], service.Port)
	} else {
		i := rand.Int() % len(addrs)
		return this.ConnectAddr(addrs[i], service.Port)
	}
}

// Connect to remote host by IP address and port
func (this *clientpool) ConnectAddr(addr net.IP, port uint16) (gopi.RPCClientConn, error) {
	this.Log.Debug("ConnectAddr", addr, ":", port)
	return nil, gopi.ErrNotImplemented
}

// Connect to remote host by Fifo
func (this *clientpool) ConnectFifo(path string) (gopi.RPCClientConn, error) {
	// Create a connection
	if conn, err := gopi.New(ClientConn{
		Fifo:    path,
		Timeout: this.timeout,
	}, this.Log.Clone(ClientConn{}.Name())); err != nil {
		return nil, err
	} else if err := conn.(*clientconn).Connect(); err != nil {
		return nil, err
	} else {
		return conn.(gopi.RPCClientConn), nil
	}
}

// Disconnect from a client connection
func (this *clientpool) Disconnect(conn gopi.RPCClientConn) error {
	if conn == nil {
		return gopi.ErrBadParameter
	} else {
		return conn.Close()
	}
}

// Create client stub for connection
func (this *clientpool) CreateStub(name string, conn gopi.RPCClientConn) gopi.RPCClientStub {
	if modules := gopi.UnitsByName(name); len(modules) == 0 {
		this.Log.Error(fmt.Errorf("%w: Client Stub %v", gopi.ErrNotFound, strconv.Quote(name)))
		return nil
	} else if modules[0].Type != gopi.UNIT_RPC_CLIENT || modules[0].Stub == nil {
		this.Log.Error(fmt.Errorf("%w: Client Stub %v", gopi.ErrBadParameter, strconv.Quote(name)))
		return nil
	} else if stub, err := modules[0].Stub(conn); err != nil {
		this.Log.Error(fmt.Errorf("%s: %w", modules[0].Name, err))
		return nil
	} else {
		return stub
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// ParseIP turns
func ParseIP(addr string) (gopi.RPCServiceRecord, error) {
	if host, port, err := net.SplitHostPort(addr); err != nil {
		return gopi.RPCServiceRecord{}, err
	} else if port_, err := strconv.ParseUint(port, 10, 16); err != nil {
		return gopi.RPCServiceRecord{}, gopi.ErrBadParameter.WithPrefix("port")
	} else if ip := net.ParseIP(host); ip != nil {
		return gopi.RPCServiceRecord{
			Addrs: []net.IP{ip},
			Port:  uint16(port_),
		}, nil
	} else if reValidHostname.MatchString(host) == false {
		return gopi.RPCServiceRecord{}, gopi.ErrBadParameter.WithPrefix("addr")
	} else if addrs, err := net.LookupHost(host); err != nil {
		return gopi.RPCServiceRecord{}, err
	} else {
		ips := make([]net.IP, 0, len(addrs))
		for _, addr := range addrs {
			if ip := net.ParseIP(addr); ip != nil {
				ips = append(ips, ip)
			}
		}
		return gopi.RPCServiceRecord{
			Host:  host,
			Addrs: ips,
			Port:  uint16(port_),
		}, nil
	}
}

// Convert RPCServiceRecord into a set of addresses
func SelectAddr(service gopi.RPCServiceRecord, flag gopi.RPCFlag) []net.IP {
	ip4 := flag&gopi.RPC_FLAG_INET_V4 == gopi.RPC_FLAG_INET_V4 || flag&(gopi.RPC_FLAG_INET_V4|gopi.RPC_FLAG_INET_V6) == 0
	ip6 := flag&gopi.RPC_FLAG_INET_V6 == gopi.RPC_FLAG_INET_V6 || flag&(gopi.RPC_FLAG_INET_V4|gopi.RPC_FLAG_INET_V6) == 0
	pool := make([]net.IP, 0, len(service.Addrs))
	for _, addr := range service.Addrs {
		if addr.To16() != nil && ip6 {
			pool = append(pool, addr)
		} else if addr.To4() != nil && ip4 {
			pool = append(pool, addr)
		}
	}
	return pool
}
