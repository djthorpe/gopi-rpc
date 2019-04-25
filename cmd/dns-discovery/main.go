/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/discovery"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////
/*
func EnumerateServices(app *gopi.AppInstance, done <-chan struct{}) error {
	discovery := app.ModuleInstance("discovery").(rpc.Discovery)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		app.Logger.Info("EnumerateServices")
		discovery.EnumerateServiceNames(ctx)
	}()

FOR_LOOP:
	for {
		select {
		case <-done:
			cancel()
			break FOR_LOOP
		}
	}
	return nil
}
*/

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	discovery := app.ModuleInstance("discovery").(gopi.RPCServiceDiscovery)
	app.Logger.Info("discovery=%v", discovery)

	app.Logger.Info("Waiting for CTRL+C")
	app.WaitForSignal()

	done <- gopi.DONE

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("discovery")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
