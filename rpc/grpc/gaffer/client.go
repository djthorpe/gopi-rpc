/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"context"
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/gaffer"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	pb.GafferClient
	conn gopi.RPCClientConn
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewGafferClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewGafferClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
}

func (this *Client) NewContext() context.Context {
	if this.conn.Timeout() == 0 {
		return context.Background()
	} else {
		ctx, _ := context.WithTimeout(context.Background(), this.conn.Timeout())
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
	if _, err := this.GafferClient.Ping(this.NewContext(), &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Client) ListExecutables() ([]string, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform ping
	if reply, err := this.GafferClient.ListExecutables(this.NewContext(), &empty.Empty{}); err != nil {
		return nil, err
	} else {
		return reply.Path, nil
	}
}

func (this *Client) ListServices() ([]rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()
	if reply, err := this.GafferClient.ListServices(this.NewContext(), &pb.RequestFilter{
		Type: pb.RequestFilter_NONE,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoServiceArray(reply.Service), nil
	}
}

func (this *Client) ListServicesForGroup(group string) ([]rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()
	if reply, err := this.GafferClient.ListServices(this.NewContext(), &pb.RequestFilter{
		Type:  pb.RequestFilter_GROUP,
		Value: group,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoServiceArray(reply.Service), nil
	}
}

func (this *Client) GetService(service string) (rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()
	if reply, err := this.GafferClient.ListServices(this.NewContext(), &pb.RequestFilter{
		Type:  pb.RequestFilter_SERVICE,
		Value: service,
	}); err != nil {
		return nil, err
	} else if len(reply.Service) == 0 {
		return nil, gopi.ErrNotFound
	} else {
		return fromProtoService(reply.Service[0]), nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<rpc.service.gaffer.Client>{ conn=%v }", this.conn)
}
