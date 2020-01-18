/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"crypto/tls"
	"net"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
	grpc "google.golang.org/grpc"
	credentials "google.golang.org/grpc/credentials"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientConn struct {
	Addr       net.IP
	Port       uint16
	Flags      gopi.RPCFlag
	SSL        bool
	SkipVerify bool
	Timeout    time.Duration
}

type clientconn struct {
	base.Unit
	sync.Mutex
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (ClientConn) Name() string { return "gopi/grpc/clientconn" }

func (config ClientConn) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(clientconn)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *clientconn) Init(config ClientConn) error {
	// Success
	return nil
}

func (this *clientconn) Close() error {
	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.ClientConn

func (this *clientconn) Connect() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Create connection options
	opts := make([]grpc.DialOption, 0, 1)

	// SSL options
	if this.ssl {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: this.skipverify})))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// Connection timeout options
	if this.timeout > 0 {
		opts = append(opts, grpc.WithTimeout(this.timeout))
	}

	// Dial connection
	if conn, err := grpc.Dial(this.addr, opts...); err != nil {
		return err
	} else {
		this.conn = conn
	}

	// Success
	return nil
}

func (this *clientconn) Disconnect() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if this.conn != nil {
		err := this.conn.Close()
		this.conn = nil
		return err
	} else {
		return nil
	}
}
