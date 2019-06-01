/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/discovery"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/googlecast"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/version"
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi-rpc/sys/rpcutil"
	_ "github.com/djthorpe/gopi/sys/logger"
)

const (
	DISCOVERY_TIMEOUT = 400 * time.Millisecond
)

////////////////////////////////////////////////////////////////////////////////

type App struct {
	app      *gopi.AppInstance
	log      gopi.Logger
	services map[string]gopi.RPCServiceRecord
}

////////////////////////////////////////////////////////////////////////////////

func NewApp(app *gopi.AppInstance) *App {
	this := new(App)
	this.app = app
	this.log = app.Logger
	this.services = make(map[string]gopi.RPCServiceRecord)
	return this
}

// Pool returns the client pool or nil
func (this *App) Pool() gopi.RPCClientPool {
	if pool, ok := this.app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool); ok == false || pool == nil {
		return nil
	} else {
		return pool
	}
}

// Return timeout or default
func (this *App) TimeoutOrDefault(timeout time.Duration) time.Duration {
	if timeout > 0 {
		return timeout
	} else {
		return DISCOVERY_TIMEOUT
	}
}

func HasHostPort(addr string) bool {
	if host, port, err := net.SplitHostPort(addr); err != nil {
		return false
	} else if host != "" && port != "" {
		return true
	} else {
		return false
	}
}

func FilterBySubtype(services []gopi.RPCServiceRecord, subtype string) []gopi.RPCServiceRecord {
	if len(services) == 0 || subtype == "" {
		return services
	}
	services_ := make([]gopi.RPCServiceRecord, 0, len(services))
	for _, service := range services {
		if service.Subtype() == subtype {
			services_ = append(services_, service)
		}
	}
	return services_
}

func StubExists(stub string, stubs []string) bool {
	for _, stub_ := range stubs {
		if stub == stub_ {
			return true
		}
	}
	return false
}

// ServiceRecord returns a remote service record, or nil if not found
func (this *App) ServiceRecord(addr string, timeout time.Duration) (gopi.RPCServiceRecord, error) {
	this.log.Debug("ServiceRecord{ addr=%v timeout=%v }", strconv.Quote(addr), this.TimeoutOrDefault(timeout))

	// If we already have the service record, return it
	if service, exists := this.services[addr]; exists {
		return service, nil
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), this.TimeoutOrDefault(timeout))
	defer cancel()

	if pool := this.Pool(); pool == nil {
		return nil, fmt.Errorf("Missing rpc/clientpool")
	} else if service, subtype, _, err := this.app.Service(); err != nil {
		return nil, err
	} else if HasHostPort(addr) {
		// Where addr is <host>:<port> return the service record
		if r, err := pool.Lookup(ctx, "", addr, 1); err != nil {
			return nil, err
		} else if len(r) != 1 {
			return nil, gopi.ErrAppError
		} else {
			// Cache and return
			this.services[addr] = r[0]
			return r[0], nil
		}
	} else {
		// Return "unlimited" service records (parameter 0)
		if services, err := pool.Lookup(ctx, fmt.Sprintf("_%v._tcp", service), addr, 0); err != nil {
			return nil, err
		} else if len(services) == 0 {
			return nil, gopi.ErrNotFound
		} else if services = FilterBySubtype(services, subtype); len(services) == 0 {
			return nil, gopi.ErrNotFound
		} else if len(services) > 1 {
			var names []string
			for _, service := range services {
				names = append(names, strconv.Quote(service.Name()))
			}
			return nil, fmt.Errorf("More than one service returned, use -addr to choose between %v", strings.Join(names, ","))
		} else {
			// Cache and return
			this.services[addr] = services[0]
			return services[0], nil
		}
	}
}

// Conn returns a connection, or nil if connection cannot be made
func (this *App) Conn(addr string, timeout time.Duration) (gopi.RPCClientConn, error) {
	this.log.Debug("Conn{ addr=%v timeout=%v }", strconv.Quote(addr), this.TimeoutOrDefault(timeout))

	if pool := this.Pool(); pool == nil {
		return nil, fmt.Errorf("Missing rpc/clientpool")
	} else if service, err := this.ServiceRecord(addr, timeout); err != nil {
		return nil, err
	} else if conn, err := pool.Connect(service, gopi.RPC_FLAG_NONE); err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

// Client returns a service stub, or nil
func (this *App) Stub(stub string, addr string, timeout time.Duration) (gopi.RPCClient, error) {
	this.log.Debug("Stub{ stub=%v addr=%v timeout=%v }", strconv.Quote(stub), strconv.Quote(addr), this.TimeoutOrDefault(timeout))

	if pool := this.Pool(); pool == nil {
		return nil, fmt.Errorf("Missing rpc/clientpool")
	} else if conn, err := this.Conn(addr, timeout); err != nil {
		return nil, err
	} else if stubs, err := conn.Services(); err != nil {
		return nil, err
	} else if StubExists(stub, stubs) == false {
		return nil, fmt.Errorf("%v: Unknown service stub %v. Possible stubs are: %v", conn.Addr(), stub, strings.Join(stubs, ", "))
	} else if stub := pool.NewClient(stub, conn); stub == nil {
		return nil, fmt.Errorf("%v: Unknown service stub %v. Possible stubs are: %v", conn.Addr(), stub, strings.Join(stubs, ", "))
	} else {
		return stub, nil
	}
}

func (this *App) Run(stub rpc.DiscoveryClient) error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	runner := NewApp(app)
	addr, _ := app.AppFlags.GetString("addr")
	timeout, _ := app.AppFlags.GetDuration("timeout")
	if stub, err := runner.Stub("gopi.Discovery", addr, timeout); err != nil {
		return err
	} else if err := runner.Run(stub.(rpc.DiscoveryClient)); err != nil {
		return err
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/discovery:client", "rpc/version:client", "googlecast:client", "discovery")

	// Set subtype as "discovery"
	//config.AppFlags.SetParam(gopi.PARAM_SERVICE_SUBTYPE, "discovery")

	// Set flags
	config.AppFlags.FlagString("addr", "", "Gateway address")
	config.AppFlags.FlagDuration("timeout", 0, "Gateway discovery timeout")

	/*
		config.AppFlags.FlagBool("dns", false, "Use DNS lookup rather than cache")
		config.AppFlags.FlagBool("watch", false, "Watch for discovery changes")
		config.AppFlags.FlagString("register", "", "Register a service with name")
		config.AppFlags.FlagString("subtype", "", "Service subtype")
		config.AppFlags.FlagUint("port", 0, "Service port")
	*/
	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
