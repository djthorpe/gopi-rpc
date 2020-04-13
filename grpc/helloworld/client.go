/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package helloworld

import (
	"context"
	"fmt"

	// Frameworks
	grpc "github.com/djthorpe/gopi-rpc/v2/unit/grpc"
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/v2/protobuf/helloworld"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	Conn gopi.RPCClientConn
}

type client struct {
	base.Unit
	conn   gopi.RPCClientConn
	client pb.HelloworldClient
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (Client) Name() string { return "gopi.Helloworld" }

func (config Client) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(client)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *client) Init(config Client) error {
	// Create the client
	if config.Conn == nil {
		return gopi.ErrBadParameter
	} else if grpcconn, ok := config.Conn.(grpc.GRPCClientConn); ok == false {
		return gopi.ErrBadParameter
	} else if client := pb.NewHelloworldClient(grpcconn.GRPCClient()); client == nil {
		return gopi.ErrBadParameter
	} else {
		this.conn = config.Conn
		this.client = client
	}

	// Success
	return nil
}

func (this *client) Close() error {
	return this.Unit.Close()
}

func (this *client) Conn() gopi.RPCClientConn {
	return this.conn
}

func (this *client) String() string {
	return "<gopi.Helloworld conn=" + fmt.Sprint(this.conn) + ">"
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION rpc.HelloworldClient

func (this *client) Ping(ctx context.Context) error {
	this.conn.Lock()
	defer this.conn.Unlock()
	if _, err := this.client.Ping(ctx, &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}
