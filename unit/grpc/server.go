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
	"os"
	"os/user"
	"strconv"
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
	File           string
	FileGroup      string
	ServerOption   []grpc.ServerOption
	Bus            gopi.Bus
}

type server struct {
	bus     gopi.Bus
	fifo    string
	fifogid *user.Group
	port    uint
	server  *grpc.Server
	addr    net.Addr
	ssl     bool

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

	if config.Bus == nil {
		return gopi.ErrBadParameter.WithPrefix("bus")
	} else {
		this.bus = config.Bus
	}

	if config.File != "" {
		if this.port != 0 || config.SSLKey != "" || config.SSLCertificate != "" {
			return fmt.Errorf("%w: Cannot combine -rpc.fifo with -rpc.port, -rpc.sslkey or -rpc.sslcert", gopi.ErrBadParameter)
		} else if _, err := os.Stat(config.File); err == nil {
			return fmt.Errorf("%w: -rpc.fifo already exists", gopi.ErrBadParameter)
		} else if fifogid, err := lookupGroup(config.FileGroup); err != nil {
			return err
		} else {
			this.fifo = config.File
			this.fifogid = fifogid
			this.server = grpc.NewServer(config.ServerOption...)
		}
	} else if config.SSLKey != "" && config.SSLCertificate != "" {
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
	this.bus = nil
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
	} else {
		return "<" + this.Log.Name() + " status=idle" + ">"
	}
}

////////////////////////////////////////////////////////////////////////////////
// SERVE AND STOP

func (this *server) Start() error {
	// Check for serving
	if this.Addr() != nil {
		return gopi.ErrInternalAppError.WithPrefix("Start")
	}

	var lis net.Listener
	var err error
	if this.fifo != "" {
		if lis, err = net.Listen("unix", this.fifo); err == nil && this.fifogid != nil {
			// Set group and allow group to write
			if err := setFifoGroup(this.fifo, this.fifogid); err != nil {
				lis.Close()
				return err
			}
		}
	} else {
		lis, err = net.Listen("tcp", portString(this.port))
	}
	if err != nil {
		return err
	}

	// Start server, wait until stopped
	this.SetAddr(lis.Addr())
	this.bus.Emit(NewEvent(this, gopi.RPC_EVENT_SERVER_STARTED))

	// blocking call
	err = this.server.Serve(lis)

	// Send message on bus that server has stopped
	this.bus.Emit(NewEvent(this, gopi.RPC_EVENT_SERVER_STOPPED))
	this.SetAddr(nil)

	// Return any error
	return err
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

func lookupGroup(name string) (*user.Group, error) {
	if name == "" {
		return nil, nil
	} else if group, err := user.LookupGroup(name); err == nil {
		return group, nil
	} else if group, err := user.LookupGroupId(name); err == nil {
		return group, nil
	} else {
		return nil, fmt.Errorf("%w: Invalid group", gopi.ErrBadParameter)
	}
}

func setFifoGroup(path string, group *user.Group) error {
	if gid, err := strconv.ParseUint(group.Gid, 10, 32); err != nil {
		return err
	} else if err := os.Chown(path, -1, int(gid)); err != nil {
		return err
	} else if err := os.Chmod(path, 0775); err != nil {
		return err
	}

	// Success
	return nil
}
