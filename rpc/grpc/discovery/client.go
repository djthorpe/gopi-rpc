/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"context"
	"fmt"
	"io"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/discovery"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	pb.DiscoveryClient
	conn gopi.RPCClientConn
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewDiscoveryClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewDiscoveryClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
}

func (this *Client) NewContext(timeout time.Duration) context.Context {
	if timeout == 0 {
		timeout = this.conn.Timeout()
	}
	if timeout == 0 {
		return context.Background()
	} else {
		ctx, _ := context.WithTimeout(context.Background(), timeout)
		return ctx
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *Client) Conn() gopi.RPCClientConn {
	return this.conn
}

////////////////////////////////////////////////////////////////////////////////
// CALLS

func (this *Client) Ping() error {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform ping
	if _, err := this.DiscoveryClient.Ping(this.NewContext(0), &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

// Register a service record
func (this *Client) Register(service gopi.RPCServiceRecord) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform register
	if _, err := this.DiscoveryClient.Register(this.NewContext(0), protoFromServiceRecord(service)); err != nil {
		return err
	} else {
		return nil
	}
}

// Enumerate service names
func (this *Client) Enumerate(t rpc.DiscoveryType, timeout time.Duration) ([]string, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	// If timeout is zero, use the connection timeout, but it can't be zero
	if timeout == 0 && this.conn.Timeout() == 0 {
		return nil, gopi.ErrBadParameter
	}

	// Perform enumerate
	if reply, err := this.DiscoveryClient.Enumerate(this.NewContext(timeout), &pb.EnumerateRequest{
		Type: protoFromDiscoveryType(t),
	}); err != nil {
		return nil, err
	} else {
		return reply.Service, nil
	}
}

// Lookup service instances
func (this *Client) Lookup(service string, t rpc.DiscoveryType, timeout time.Duration) ([]gopi.RPCServiceRecord, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	// If timeout is zero, use the connection timeout, but it can't be zero
	if timeout == 0 && this.conn.Timeout() == 0 {
		return nil, gopi.ErrBadParameter
	}

	// Perform lookup
	if reply, err := this.DiscoveryClient.Lookup(this.NewContext(timeout), &pb.LookupRequest{
		Service: service,
		Type:    protoFromDiscoveryType(t),
	}); err != nil {
		return nil, err
	} else {
		return protoToServiceRecords(reply.Service), nil
	}
}

// Stream events
func (this *Client) StreamEvents(service string, events chan<- gopi.RPCEvent) error {
	// One request per connection
	this.conn.Lock()
	defer this.conn.Unlock()

	// Keep reading from stream
	if stream, err := this.DiscoveryClient.StreamEvents(this.NewContext(0), &pb.StreamEventsRequest{
		Service: service,
	}); err != nil {
		return err
	} else {
		for {
			if msg, err := stream.Recv(); err == io.EOF {
				break
			} else if err != nil {
				return err
			} else if evt := protoToEvent(msg); evt != nil {
				events <- evt
			}
		}
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<rpc.service.discovery.Client>{ conn=%v }", this.conn)
}
