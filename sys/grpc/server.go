/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2016-2018
	All Rights Reserved

	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
	grpc "google.golang.org/grpc"
	credentials "google.golang.org/grpc/credentials"
	reflection "google.golang.org/grpc/reflection"
)

// Server is the RPC server configuration
type Server struct {
	SSLKey         string
	SSLCertificate string
	Port           uint
	ServerOption   []grpc.ServerOption
	Util           rpc.Util
}

type server struct {
	log    gopi.Logger
	port   uint
	server *grpc.Server
	addr   net.Addr
	ssl    bool
	util   rpc.Util

	event.Publisher
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	reService = regexp.MustCompile("[A-za-z][A-Za-z0-9\\-]*")
)

////////////////////////////////////////////////////////////////////////////////
// SERVER OPEN AND CLOSE

// Open the server
func (config Server) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.Server>Open(port=%v,sslcert=\"%v\",sslkey=\"%v\")", config.Port, config.SSLCertificate, config.SSLKey)

	this := new(server)
	this.log = log
	this.port = config.Port
	this.ssl = false
	this.util = config.Util

	if this.util == nil {
		return nil, gopi.ErrBadParameter
	}

	if config.SSLKey != "" && config.SSLCertificate != "" {
		if creds, err := credentials.NewServerTLSFromFile(config.SSLCertificate, config.SSLKey); err != nil {
			return nil, err
		} else {
			this.server = grpc.NewServer(append(config.ServerOption, grpc.Creds(creds))...)
		}
		this.ssl = true
	} else {
		this.server = grpc.NewServer(config.ServerOption...)
	}

	this.addr = nil

	// Register reflection service on gRPC server.
	reflection.Register(this.server)

	// success
	return this, nil
}

// Close server
func (this *server) Close() error {
	this.log.Debug("<grpc.Server>Close( addr=%v )", this.addr)

	// Ungracefully stop the server
	err := this.Stop(true)
	if err != nil {
		this.log.Warn("grpc.Server: %v", err)
	}

	// Close publisher
	this.Publisher.Close()

	// Release resources
	this.addr = nil
	this.server = nil

	// Return any error that occurred
	return err
}

////////////////////////////////////////////////////////////////////////////////
// SERVE

func (this *server) Start() error {
	this.log.Debug2("<grpc.Server>Start()")

	// Check for serving
	if this.addr != nil {
		return errors.New("Cannot call Start() when server already started")
	} else if lis, err := net.Listen("tcp", portString(this.port)); err != nil {
		return err
	} else {
		// Start server
		this.addr = lis.Addr()
		this.Emit(this.util.NewEvent(this, gopi.RPC_EVENT_SERVER_STARTED, nil))
		this.log.Info("Listening on addresss: %v", this.addr)
		err := this.server.Serve(lis) // blocking call
		this.Emit(this.util.NewEvent(this, gopi.RPC_EVENT_SERVER_STOPPED, nil))
		this.addr = nil
		return err
	}
}

func (this *server) Stop(halt bool) error {
	// Stop server
	if this.addr != nil {
		if halt {
			this.log.Debug2("<grpc.Server>Stop()")
			this.server.Stop()
		} else {
			this.log.Debug2("<grpc.Server>GracefulStop()")
			this.server.GracefulStop()
		}
	}

	// Return success
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Addr returns the currently listening address or will return
// nil if the server is not serving requests
func (this *server) Addr() net.Addr {
	return this.addr
}

// Port returns the expected port the server will be listening on
func (this *server) Port() uint {
	if this.addr == nil {
		return this.port
	} else if _, port, err := net.SplitHostPort(this.addr.String()); err != nil {
		return 0
	} else if port_, err := strconv.ParseUint(port, 10, 64); err != nil {
		return 0
	} else {
		return uint(port_)
	}
}

// Return the gRPC server object
func (this *server) GRPCServer() *grpc.Server {
	return this.server
}

///////////////////////////////////////////////////////////////////////////////
// SERVICE

func (this *server) Service(service, subtype, name string, text ...string) gopi.RPCServiceRecord {
	this.log.Debug2("<grpc.Service>{ service=%v subtype=%v name=%v text=%v }", strconv.Quote(service), strconv.Quote(subtype), strconv.Quote(name), text)

	// Can't return a service unless the server is started
	if this.addr == nil {
		this.log.Warn("grpc.Service: No address")
		return nil
	}

	// Can't register if name is blank
	if strings.TrimSpace(name) == "" {
		this.log.Warn("grpc.Service: No name")
		return nil
	}

	// Check service name
	if matched, err := regexp.MatchString("^[A-Za-z][A-Za-z0-9\\-]*$", service); err != nil {
		this.log.Warn("grpc.Service: SetService: %v", err)
		return nil
	} else if matched == false {
		this.log.Warn("grpc.Service: SetService: Invalid service type")
		return nil
	} else {
		service = fmt.Sprintf("_%v._%v", service, this.Addr().Network())
	}

	// Return service
	if _, ok := this.addr.(*net.TCPAddr); ok == false {
		return nil
	} else {
		r := this.util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB)

		// Set service, subtype, etc.
		if err := r.SetService(service, subtype); err != nil {
			this.log.Warn("grpc.Service: SetService: %v", err)
			return nil
		} else if err := r.SetName(name); err != nil {
			this.log.Warn("grpc.Service: SetName: %v", err)
			return nil
		} else if hostname, err := os.Hostname(); err != nil {
			this.log.Warn("grpc.Service: SetAddr: %v", err)
			return nil
		} else if err := r.SetAddr(fmt.Sprintf("%v:%v", hostname, this.Port())); err != nil {
			this.log.Warn("grpc.Service: SetAddr: %v", err)
			return nil
		} else if v4, v6, err := addrsForInterfaces(); err != nil {
			this.log.Warn("grpc.Service: SetAddr: %v", err)
			return nil
		} else if err := r.AppendIP(v4...); err != nil {
			this.log.Warn("grpc.Service: AppendIP: IPv4: %v", err)
			return nil
		} else if err := r.AppendIP(v6...); err != nil {
			this.log.Warn("grpc.Service: AppendIP: IPv6: %v", err)
			return nil
		}

		// Set a TXT record for SSL
		if err := r.AppendTXT(toSslTXT(this.ssl)); err != nil {
			this.log.Warn("grpc.Service: AppendTXT: %v", err)
			return nil
		}

		return r
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *server) String() string {
	if this.addr != nil {
		return fmt.Sprintf("<grpc.Server>{ serving,addr=%v }", this.addr)
	} else if this.port == 0 {
		return fmt.Sprintf("<grpc.Server>{ idle }")
	} else {
		return fmt.Sprintf("<grpc.Server>{ idle,port=%v }", this.port)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toSslTXT(ssl bool) string {
	if ssl {
		return "ssl=1"
	} else {
		return "ssl=0"
	}
}

func portString(port uint) string {
	if port == 0 {
		return ""
	} else {
		return fmt.Sprint(":", port)
	}
}

func addrsForInterfaces() ([]net.IP, []net.IP, error) {
	if ifaces, err := net.Interfaces(); err != nil {
		return nil, nil, err
	} else {
		var v4, v6 []net.IP
		for _, iface := range ifaces {
			if iface.Flags&net.FlagLoopback != 0 {
				continue
			} else if iface.Flags&net.FlagUp == 0 {
				continue
			}
			if v4_, v6_, err := addrsForInterface(iface); err != nil {
				return nil, nil, err
			} else {
				v4 = append(v4, v4_...)
				v6 = append(v6, v6_...)
			}
		}
		return v4, v6, nil
	}
}

func addrsForInterface(iface net.Interface) ([]net.IP, []net.IP, error) {
	if addrs, err := iface.Addrs(); err != nil {
		return nil, nil, err
	} else {
		var v4, v6 []net.IP
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok == false {
				continue
			} else if ipnet.IP.IsLoopback() == true {
				continue
			} else if ipnet.IP.To4() != nil {
				v4 = append(v4, ipnet.IP)
			} else if ip := ipnet.IP.To16(); ip != nil && ip.IsGlobalUnicast() {
				v6 = append(v6, ipnet.IP)
			}
		}
		return v4, v6, nil
	}
}
