/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"fmt"
	"os"

	rpc "github.com/djthorpe/gopi-rpc"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

var (
	commands = make(chan string, 10)
)

func Watch(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	googlecast := app.ModuleInstance("googlecast").(rpc.GoogleCast)
	start <- gopi.DONE

	// If there is an argument, then this is the service to lookup
	events := googlecast.Subscribe()
FOR_LOOP:
	for {
		select {
		case cmd := <-commands:
			fmt.Println("Execute: %v", cmd)
		case evt := <-events:
			if evt_, ok := evt.(rpc.GoogleChannelEvent); ok {
				if evt_.Type() == rpc.GOOGLE_CAST_EVENT_APPLICATION {
					// Connect to first application
					if apps := evt_.Channel().Applications(); len(apps) > 0 {
						if err := evt_.Channel().SetApplication(apps[0]); err != nil {
							app.Logger.Error("Error: %v", err)
						} else {
							fmt.Println("APP", apps[0])
						}
					}
				} else if evt_.Type() == rpc.GOOGLE_CAST_EVENT_VOLUME {
					fmt.Println("VOLUME", evt_.Channel().Volume())
				} else if evt_.Type() == rpc.GOOGLE_CAST_EVENT_MEDIA {
					fmt.Println("MEDIA", evt_.Channel().Media())
				}
			} else if evt_, ok := evt.(rpc.GoogleCastEvent); ok {
				if evt_.Type() == gopi.RPC_EVENT_SERVICE_ADDED {
					fmt.Println("DEVICE", evt_)
					if _, err := googlecast.Connect(evt_.Device(), gopi.RPC_FLAG_INET_V4|gopi.RPC_FLAG_INET_V6, 0); err != nil {
						app.Logger.Error("Error: %v", err)
					}
				}
			}
		case <-stop:
			break FOR_LOOP
		}
	}
	googlecast.Unsubscribe(events)
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	// Queue commands
	for _, arg := range app.AppFlags.Args() {
		commands <- arg
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
	config := gopi.NewAppConfig("googlecast", "discovery")

	config.AppFlags.FlagString("id", "", "Chromecast ID")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Watch))
}
