/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// SERVER

func Server(config gopi.AppConfig, background_tasks ...gopi.BackgroundTask) int {
	// Append on "rpc/server" onto module configuration
	var err error
	if config.Modules, err = gopi.AppendModulesByName(config.Modules, "rpc/server"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return -1
	}

	// Create the application
	app, err := gopi.NewAppInstance(config)
	if err != nil {
		if err != gopi.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return -1
		}
		return 0
	}
	defer app.Close()

	// Run the application with a main task and background tasks
	if err := app.Run2(MainTask, ServerTask, RegisterTask); err == gopi.ErrHelp {
		config.AppFlags.PrintUsage()
		return 0
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return -1
	} else {
		return 0
	}
}

func MainTask(app *gopi.AppInstance, done chan<- struct{}) error {
	// Wait for CTRL+C or SIGTERM
	app.Logger.Info("Waiting for CTRL+C or SIGTERM to stop server")
	app.WaitForSignal()
	done <- gopi.DONE
	return nil
}

func ServerTask(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	if server, ok := app.ModuleInstance("rpc/server").(gopi.RPCServer); ok == false {
		return fmt.Errorf("rpc/server missing")
	} else {
		go func() {
			<-stop

			// Cancel on-going requests for all services
			for _, module := range gopi.ModulesByType(gopi.MODULE_TYPE_SERVICE) {
				app.Logger.Debug("CancelRequests: %v", module.Name)
				if instance, ok := app.ModuleInstance(module.Name).(gopi.RPCService); ok {
					if err := instance.CancelRequests(); err != nil {
						app.Logger.Warn("CancelRequests: %v: %v", module.Name, err)
					}
				}
			}

			// Stop the server
			if err := server.Stop(false); err != nil {
				app.Logger.Error("Stop: %v", err)
			}
		}()
		start <- gopi.DONE
		app.Logger.Debug("Server starting")
		if err_ := server.Start(); err_ != nil {
			return err_
		}
		app.Logger.Debug("Server stopped")
	}

	// Success
	return nil
}

func RegisterTask(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	discovery := app.ModuleInstance("discovery")
	if server, ok := app.ModuleInstance("rpc/server").(gopi.RPCServer); ok == false {
		return fmt.Errorf("rpc/server missing")
	} else {
		discovery_, _ := discovery.(gopi.RPCServiceDiscovery)
		if discovery_ == nil {
			app.Logger.Info("Service Discovery not enabled")
		}
		start <- gopi.DONE
		evts := server.Subscribe()
	FOR_LOOP:
		for {
			select {
			case <-stop:
				break FOR_LOOP
			case evt := <-evts:
				if evt_, ok := evt.(gopi.RPCEvent); ok == false {
					app.Logger.Warn("Not processing: %v", evt)
				} else if err := ProcessEvent(app, server, discovery_, evt_); err != nil {
					app.Logger.Warn("%v", err)
				}
			}
		}
		server.Unsubscribe(evts)
	}

	// Success
	return nil
}

func ProcessEvent(app *gopi.AppInstance, server gopi.RPCServer, discovery gopi.RPCServiceDiscovery, evt gopi.RPCEvent) error {
	// Obtain discovery unit
	switch evt.Type() {
	case gopi.RPC_EVENT_SERVER_STARTED:
		if service, subtype, name, err := app.Service(); err != nil {
			return err
		} else if service_ := server.Service(service, subtype, name); service_ == nil {
			return fmt.Errorf("Unable to create service record")
		} else if discovery == nil {
			return nil
		} else if err := discovery.Register(service_); err != nil {
			return err
		}
	}
	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// CLIENT

type ClientTask func(app *gopi.AppInstance, services []gopi.RPCServiceRecord, done chan<- struct{}) error

func Client(config gopi.AppConfig, timeout time.Duration, task ClientTask) int {
	// Add the -addr flag
	config.AppFlags.FlagString("addr", "", "Service name or address:port")

	// Append on "clientpool" onto module configuration
	var err error
	if config.Modules, err = gopi.AppendModulesByName(config.Modules, "clientpool"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return -1
	}

	// Create the application
	app, err := gopi.NewAppInstance(config)
	if err != nil {
		if err != gopi.ErrHelp {
			fmt.Fprintln(os.Stderr, err)
			return -1
		}
		return 0
	}
	defer app.Close()

	// Lookup service records
	if err := app.Run2(func(app *gopi.AppInstance, done chan<- struct{}) error {
		addr, _ := app.AppFlags.GetString("addr")
		if records, err := LookupServiceRecords(app, addr, timeout); err != nil {
			return err
		} else {
			return task(app, records, done)
		}
	}); err == gopi.ErrHelp {
		config.AppFlags.PrintUsage()
		return 0
	} else if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return -1
	} else {
		return 0
	}
}

// LookupServiceRecords returns a remote service records, or nil if not found
func LookupServiceRecords(app *gopi.AppInstance, addr string, timeout time.Duration) ([]gopi.RPCServiceRecord, error) {
	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if service, subtype, _, err := app.Service(); err != nil {
		return nil, err
	} else if HasHostPort(addr) {
		// Where addr is <host>:<port> return the service record
		if r, err := app.ClientPool.Lookup(ctx, "", addr, 1); err != nil {
			return nil, err
		} else {
			return r, nil
		}
	} else {
		// Return "unlimited" service records (parameter 0)
		if services, err := app.ClientPool.Lookup(ctx, fmt.Sprintf("_%v._tcp", service), addr, 0); err != nil {
			return nil, err
		} else if len(services) == 0 {
			return nil, gopi.ErrNotFound
		} else if services = FilterBySubtype(services, subtype); len(services) == 0 {
			return nil, gopi.ErrNotFound
		} else {
			return services, nil
		}
	}
}

// HasHostPort returns true if string is of type <hostname>:<port>
func HasHostPort(addr string) bool {
	if host, port, err := net.SplitHostPort(addr); err != nil {
		return false
	} else if host != "" && port != "" {
		return true
	} else {
		return false
	}
}

// FilterBySubtype returns set of records which match subtype
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
