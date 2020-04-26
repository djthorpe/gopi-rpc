/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	base "github.com/djthorpe/gopi/v2/base"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

type server struct {
	base.App
	sync.WaitGroup

	main gopi.MainCommandFunc
}

////////////////////////////////////////////////////////////////////////////////
// gopi.App implementation for command-line tool

func NewServer(main gopi.MainCommandFunc, units ...string) (gopi.App, error) {
	this := new(server)

	// Name of the server
	name := filepath.Base(os.Args[0])

	// Append required units
	units = append(units, "server")

	// Check parameters
	if err := this.App.Init(name, units); err != nil {
		return nil, err
	} else {
		this.main = main
	}

	// Success
	return this, nil
}

func (this *server) Run() int {
	// Initialize the application
	if err := this.App.Start(this, os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) == false {
			fmt.Fprintln(os.Stderr, this.App.Flags().Name()+":", err)
			return -1
		}
	}

	// Defer closing of instances to exit
	defer func() {
		if err := this.App.Close(); err != nil {
			fmt.Fprintln(os.Stderr, this.App.Flags().Name()+":", err)
		}
	}()

	// Add handler for server start and stop
	if err := this.Bus().NewHandler(gopi.EventHandler{
		Name:    "gopi.RPCEvent",
		Handler: this.RPCEventHandler,
		EventNS: gopi.EVENT_NS_DEFAULT,
	}); err != nil {
		fmt.Fprintln(os.Stderr, this.App.Flags().Name()+":", err)
		return -1
	}

	server := this.UnitInstance("server").(gopi.RPCServer)
	if server == nil {
		fmt.Fprintln(os.Stderr, this.App.Flags().Name()+":", gopi.ErrInternalAppError.WithPrefix("server"))
		return -1
	}

	// Gather services
	services := make([]gopi.RPCService, 0, len(gopi.UnitsByType(gopi.UNIT_RPC_SERVICE)))
	for _, unit := range gopi.UnitsByType(gopi.UNIT_RPC_SERVICE) {
		if instance, ok := this.UnitInstance(unit.Name).(gopi.RPCService); instance == nil {
			// No instance created for service
			continue
		} else if ok == false {
			// Cannot cast to gopi.RPCService
			fmt.Fprintln(os.Stderr, this.App.Flags().Name()+":", fmt.Errorf("Invalid RPCService: %v", strconv.Quote(unit.Name)))
			return -1
		} else {
			services = append(services, instance)
			this.Log().Debug("Run: Service", instance)
		}
	}

	// Run main function in background
	go func() {
		this.WaitGroup.Add(1)
		defer this.WaitGroup.Done()
		if err := this.main(this, this.Flags().Args()); err != nil {
			fmt.Fprintln(os.Stderr, this.App.Flags().Name()+":", err)
		}

		// Send cancels
		for _, service := range services {
			this.Log().Debug("Run: Sending cancel request to", service)
			if err := service.CancelRequests(); err != nil {
				fmt.Fprintf(os.Stderr, "  %s\n", err)
			}
			this.Log().Debug("Run: Finished sending cancel request to", service)
		}

		// Stop server gracefully
		server.Stop(false)
	}()

	// Start server in foreground
	if err := server.Start(); err != nil {
		fmt.Fprintln(os.Stderr, this.App.Flags().Name()+":", err)
		return -1
	}

	// Wait for main to end and for server to be stopped
	this.Log().Debug("Run: Waiting for server to stop")
	this.WaitGroup.Wait()
	this.Log().Debug("Run: Server stopped")

	// Success
	return 0
}

func (this *server) RPCEventHandler(_ context.Context, _ gopi.App, evt gopi.Event) {
	rpcEvent := evt.(gopi.RPCEvent)
	switch rpcEvent.Type() {
	case gopi.RPC_EVENT_SERVER_STARTED:
		server := rpcEvent.Source().(gopi.RPCServer)
		this.WaitGroup.Add(1)
		this.Log().Debug("Server started", server.Addr())
		// TODO: Register service
	case gopi.RPC_EVENT_SERVER_STOPPED:
		// TODO: Unregister service
		this.Log().Debug("Server stopped")
		this.WaitGroup.Done()
	}
}
