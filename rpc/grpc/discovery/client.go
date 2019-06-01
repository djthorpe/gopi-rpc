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
	event "github.com/djthorpe/gopi/util/event"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/discovery"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	pb.DiscoveryClient
	conn gopi.RPCClientConn
	event.Publisher
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewDiscoveryClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewDiscoveryClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn, event.Publisher{}}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<grpc.service.discovery.Client>{ conn=%v }", this.conn)
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

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
func (this *Client) StreamEvents(ctx context.Context, service string) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	stream, err := this.DiscoveryClient.StreamEvents(ctx, &pb.StreamEventsRequest{
		Service: service,
	})
	if err != nil {
		return err
	}

	// Errors channel receives errors from recv
	errors := make(chan error)

	// Receive messages in the background
	go func() {
	FOR_LOOP:
		for {
			if message_, err := stream.Recv(); err == io.EOF {
				break FOR_LOOP
			} else if err != nil {
				errors <- err
				break FOR_LOOP
			} else if evt := protoToEvent(message_, this.conn); evt != nil {
				this.Emit(evt)
			}
		}
	}()

	// Continue until error or io.EOF is returned
	for {
		select {
		case err := <-errors:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
