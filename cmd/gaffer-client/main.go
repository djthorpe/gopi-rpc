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
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/gaffer"
	_ "github.com/djthorpe/gopi/sys/logger"
)

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	gaffer := app.ModuleInstance("gaffer").(rpc.Gaffer)

	groups := []string{"prod", "dev", "staging"}
	for _, name := range groups {
		if group, err := gaffer.AddGroupForName(name); err != nil {
			app.Logger.Warn("%v: %v", name, err)
		} else {
			app.Logger.Info("%v: %v", name, group)
		}
	}

	for _, exec := range gaffer.Executables(true) {
		if service, err := gaffer.AddServiceForPath(exec); err != nil {
			app.Logger.Warn("%v: %v", exec, err)
		} else {
			app.Logger.Info("%v: %v", exec, service)
		}
	}

	for _, name := range groups {
		if err := gaffer.RemoveGroupForName(name); err != nil {
			app.Logger.Warn("%v: %v", name, err)
		}
	}

	app.Logger.Info("Waiting for CTRL+C")
	app.WaitForSignal()
	done <- gopi.DONE

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("gaffer")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
