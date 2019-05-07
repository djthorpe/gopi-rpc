/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	"context"
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/googlecast"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	pb.GoogleCastClient
	conn gopi.RPCClientConn
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewGoogleCastClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewGoogleCastClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
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
	if _, err := this.GoogleCastClient.Ping(this.NewContext(0), &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Client) Devices() ([]rpc.GoogleCastDevice, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GoogleCastClient.Devices(this.NewContext(0), &empty.Empty{}); err != nil {
		return nil, err
	} else {
		return fromProtobufGoogleCastDeviceReply(reply.Device), nil
	}
}

func (this *Client) StreamEvents(string, chan<- rpc.GoogleCastEvent) error {
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<rpc.service.googlecast.Client>{ conn=%v }", this.conn)
}
