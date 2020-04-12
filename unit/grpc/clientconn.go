/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
	grpc "google.golang.org/grpc"
	credentials "google.golang.org/grpc/credentials"
	reflection_pb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientConn struct {
	Fifo       string
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

	opts []grpc.DialOption
	conn *grpc.ClientConn
	addr string
	port uint16
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (ClientConn) Name() string { return "gopi/grpc/client/conn" }

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
	this.opts = make([]grpc.DialOption, 0, 1)

	// SSL options
	if config.SSL {
		this.opts = append(this.opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: config.SkipVerify})))
	} else {
		this.opts = append(this.opts, grpc.WithInsecure())
	}

	// Timeout options
	if config.Timeout > 0 {
		this.opts = append(this.opts, grpc.WithTimeout(config.Timeout))
	}

	// Set address
	// See https://github.com/grpc/grpc/blob/master/doc/naming.md
	if config.Fifo != "" {
		if config.Addr != nil {
			return fmt.Errorf("%w: Cannot specify both fifo and address", gopi.ErrBadParameter)
		} else {
			this.addr = "unix:" + config.Fifo
		}
	} else if config.Addr == nil || config.Port == 0 {
		return fmt.Errorf("%w: Missing address or port", gopi.ErrBadParameter)
	} else if v6 := config.Addr.To16(); v6 != nil {
		this.addr = fmt.Sprintf("ipv6:[%v]:%v", v6.String(), config.Port)
		this.port = config.Port
	} else if v4 := config.Addr.To4(); v4 != nil {
		this.addr = fmt.Sprintf("ipv4:%v:%v", v4.String(), config.Port)
		this.port = config.Port
	} else {
		return gopi.ErrInternalAppError
	}

	// Success
	return nil
}

func (this *clientconn) Close() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Close gRPC connection
	if this.conn != nil {
		if err := this.conn.Close(); err != nil {
			return err
		}
	}

	// Free any resources
	this.conn = nil

	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// CONNECT and DISCONNECT

func (this *clientconn) Connect() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if this.conn != nil {
		return gopi.ErrOutOfOrder
	}

	// Dial connection
	if conn, err := grpc.Dial(this.addr, this.opts...); err != nil {
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

	// Already disconnected
	if this.conn == nil {
		return nil
	}

	// Perform disconnection and return any error
	err := this.conn.Close()
	this.conn = nil
	return err
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *clientconn) Addr() string {
	return this.addr
}

func (this *clientconn) Services() ([]string, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	if this.conn == nil {
		return nil, gopi.ErrOutOfOrder
	} else if client, err := reflection_pb.NewServerReflectionClient(this.conn).ServerReflectionInfo(context.Background()); err != nil {
		return nil, err
	} else {
		defer client.CloseSend()
		if services, err := this.listServices(client); err != nil {
			return nil, err
		} else {
			return services, nil
		}
	}
}

func (this *clientconn) Lock() {
	this.Mutex.Lock()
}

func (this *clientconn) Unlock() {
	this.Mutex.Unlock()
}

func (this *clientconn) GRPCClient() grpc.ClientConnInterface {
	return this.conn
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *clientconn) listServices(c reflection_pb.ServerReflection_ServerReflectionInfoClient) ([]string, error) {
	if err := c.Send(&reflection_pb.ServerReflectionRequest{
		MessageRequest: &reflection_pb.ServerReflectionRequest_ListServices{},
	}); err != nil {
		return nil, err
	}
	if resp, err := c.Recv(); err != nil {
		return nil, err
	} else if modules := resp.GetListServicesResponse(); modules == nil {
		return nil, fmt.Errorf("%s: GetListServicesResponse", gopi.ErrUnexpectedResponse)
	} else {
		module_services := modules.GetService()
		module_names := make([]string, len(module_services))
		for i, service := range module_services {
			// Full name of a registered service, including its package name
			module_names[i] = service.Name
		}
		return module_names, nil
	}
}
