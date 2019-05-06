/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package version

import (
	"context"
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/version"
	ptypes "github.com/golang/protobuf/ptypes"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Client struct {
	pb.VersionClient
	conn gopi.RPCClientConn
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewVersionClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewVersionClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
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
	if _, err := this.VersionClient.Ping(this.NewContext(), &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Client) Version() (map[string]string, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform SayHello
	if reply, err := this.VersionClient.Version(this.NewContext(), &empty.Empty{}); err != nil {
		return nil, err
	} else {
		if reply.Hostname != "" {
			reply.Param["hostname"] = reply.Hostname
		}
		if uptime, err := ptypes.Duration(reply.HostUptime); err == nil && uptime != 0 {
			reply.Param["hostuptime"] = fmt.Sprint(uptime.Truncate(time.Second))
		}
		if uptime, err := ptypes.Duration(reply.ServiceUptime); err == nil && uptime != 0 {
			reply.Param["serviceuptime"] = fmt.Sprint(uptime.Truncate(time.Second))
		}
		return reply.Param, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<rpc.service.version.Client>{ conn=%v }", this.conn)
}
