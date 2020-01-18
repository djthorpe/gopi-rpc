/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"fmt"
	"net"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
	grpc "google.golang.org/grpc"
	credentials "google.golang.org/grpc/credentials"
	reflection "google.golang.org/grpc/reflection"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Server struct {
	SSLKey         string
	SSLCertificate string
	Port           uint
	ServerOption   []grpc.ServerOption
}

type server struct {
	port   uint
	server *grpc.Server
	addr   net.Addr
	ssl    bool

	sync.Mutex
	base.Unit
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION gopi.Unit

func (Server) Name() string { return "gopi/grpc/server" }

func (config Server) New(log gopi.Logger) (gopi.Unit, error) {
	this := new(server)
	if err := this.Unit.Init(log); err != nil {
		return nil, err
	} else if err := this.Init(config); err != nil {
		return nil, err
	}

	// Success
	return this, nil
}

func (this *server) Init(config Server) error {
	this.port = config.Port
	this.ssl = false
	this.addr = nil

	if config.SSLKey != "" && config.SSLCertificate != "" {
		if creds, err := credentials.NewServerTLSFromFile(config.SSLCertificate, config.SSLKey); err != nil {
			return err
		} else {
			this.server = grpc.NewServer(append(config.ServerOption, grpc.Creds(creds))...)
		}
		this.ssl = true
	} else if config.SSLKey != "" {
		return gopi.ErrBadParameter.WithPrefix("SSLKey")
	} else if config.SSLCertificate != "" {
		return gopi.ErrBadParameter.WithPrefix("SSLCertificate")
	} else {
		this.server = grpc.NewServer(config.ServerOption...)
	}

	// Register reflection service on gRPC server.
	reflection.Register(this.server)

	// Success
	return nil
}

func (this *server) Close() error {

	// Ungracefully stop the server
	if err := this.Stop(true); err != nil {
		return err
	}

	// Release resources
	this.server = nil
	this.addr = nil

	return this.Unit.Close()
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *server) String() string {
	if this.Closed {
		return "<" + this.Log.Name() + ">"
	} else if this.addr != nil {
		return "<" + this.Log.Name() + " status=serving addr=" + this.Addr().String() + ">"
	} else if this.port == 0 {
		return "<" + this.Log.Name() + " status=idle" + ">"
	} else {
		return "<" + this.Log.Name() + " status=idle port=" + fmt.Sprint(this.port) + ">"
	}
}

////////////////////////////////////////////////////////////////////////////////
// SERVE AND STOP

func (this *server) Start() error {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Check for serving
	if this.Addr() != nil {
		return gopi.ErrInternalAppError.WithPrefix("Start")
	} else if lis, err := net.Listen("tcp", portString(this.port)); err != nil {
		return err
	} else {
		// Start server, wait until stopped
		this.SetAddr(lis.Addr())
		err := this.server.Serve(lis) // blocking call
		this.SetAddr(nil)
		return err
	}
}

func (this *server) Stop(halt bool) error {
	if this.Addr() != nil && this.server != nil {
		if halt {
			this.server.Stop()
		} else {
			this.server.GracefulStop()
		}
	}
	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *server) Addr() net.Addr {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	return this.addr
}

func (this *server) SetAddr(addr net.Addr) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()
	this.addr = addr
}

func (this *server) GRPCServer() *grpc.Server {
	return this.server
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func portString(port uint) string {
	if port == 0 {
		return ""
	} else {
		return fmt.Sprint(":", port)
	}
}
