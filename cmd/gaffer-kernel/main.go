/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
	app "github.com/djthorpe/gopi-rpc/v2/app"
	gopi "github.com/djthorpe/gopi/v2"

	// Units
	_ "github.com/djthorpe/gopi-rpc/v2/grpc/gaffer"
	_ "github.com/djthorpe/gopi-rpc/v2/unit/gaffer"
	_ "github.com/djthorpe/gopi-rpc/v2/unit/grpc"
	_ "github.com/djthorpe/gopi/v2/unit/bus"
	_ "github.com/djthorpe/gopi/v2/unit/logger"
)

////////////////////////////////////////////////////////////////////////////////
// INTERFACE

type GafferKernelEx interface {
	rpc.GafferKernel

	// Extended version of create process
	CreateProcessEx(id uint32, service rpc.GafferService) (uint32, error)
}

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP SERVICE

func StartService(app gopi.App) error {
	args := make([]string, 0)

	if kernel := app.UnitInstance("gaffer/kernel").(GafferKernelEx); kernel == nil {
		return gopi.ErrInternalAppError.WithPrefix("StartService")
	} else if service := app.Flags().GetString("gaffer.service", gopi.FLAG_NS_DEFAULT); service == "" {
		// No service to start
		return nil
	} else {
		if port := app.Flags().GetUint("gaffer.port", gopi.FLAG_NS_DEFAULT); port != 0 {
			args = append(args, "-rpc.port", fmt.Sprint(port))
		}
		if fifo := app.Flags().GetString("rpc.sock", gopi.FLAG_NS_DEFAULT); fifo != "" {
			args = append(args, "-kernel.sock", fifo)
		}
		if app.Log().IsDebug() {
			args = append(args, "-debug")
		}
		if id, err := kernel.CreateProcessEx(0, rpc.GafferService{
			Path:  service,
			User:  app.Flags().GetString("gaffer.user", gopi.FLAG_NS_DEFAULT),
			Group: app.Flags().GetString("gaffer.group", gopi.FLAG_NS_DEFAULT),
			Args:  args,
		}); err != nil {
			return err
		} else if err := kernel.RunProcess(id); err != nil {
			return err
		} else {
			app.Log().Info("Running", service, strings.Join(args, " "))
		}
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// EVENT HANDLER

func RPCEventHandler(_ context.Context, app gopi.App, evt gopi.Event) {
	rpcEvent := evt.(gopi.RPCEvent)
	switch rpcEvent.Type() {
	case gopi.RPC_EVENT_SERVER_STARTED:
		server := rpcEvent.Source().(gopi.RPCServer)
		app.Log().Info("Server started", server.Addr())
		if err := StartService(app); err != nil {
			app.Log().Error(err)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// MAIN

func Main(app gopi.App, args []string) error {
	// We require an rpc.sock argument. Remove any old socket which exists
	if fifo := app.Flags().GetString("rpc.sock", gopi.FLAG_NS_DEFAULT); fifo == "" {
		return fmt.Errorf("Missing required flag -rpc.sock")
	}

	// Add handler for server start and stop
	if err := app.Bus().NewHandler(gopi.EventHandler{
		Name:    "gopi.RPCEvent",
		Handler: RPCEventHandler,
		EventNS: gopi.EVENT_NS_DEFAULT,
	}); err != nil {
		return err
	}

	// Wait until CTRL+C pressed
	fmt.Println("Press CTRL+C to exit")
	app.WaitForSignal(context.Background(), os.Interrupt)
	fmt.Println("Received interrupt signal, exiting")

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP

func main() {
	if app, err := app.NewServer(Main, "rpc/gaffer/kernel"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		// Add bootstrap arguments
		app.Flags().FlagString("gaffer.service", "", "Gaffer service binary")
		app.Flags().FlagUint("gaffer.port", 0, "Gaffer service port")
		app.Flags().FlagString("gaffer.user", "", "Gaffer service user")
		app.Flags().FlagString("gaffer.group", "", "Gaffer service group")

		// Run and exit
		os.Exit(app.Run())
	}
}
