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
	"time"

	rpc "github.com/djthorpe/gopi-rpc"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

var (
	channels = make(map[string]rpc.GoogleCastChannel)
	commands = make(chan string, 10)
)

func Execute(app *gopi.AppInstance, command string) error {
	id, _ := app.AppFlags.GetString("id")
	// get channel which is connected
	if channel, exists := channels[id]; exists == false {
		return fmt.Errorf("Missing or invalid -id flag")
	} else {
		switch command {
		case "pause":
			if channel.Media() == nil {
				return fmt.Errorf("Nothing playing")
			} else if _, err := channel.SetPause(true); err != nil {
				return err
			}
		case "play":
			if channel.Media() == nil {
				return fmt.Errorf("Nothing playing")
			} else if _, err := channel.SetPlay(true); err != nil {
				return err
			}
		case "stop":
			if _, err := channel.SetPlay(false); err != nil {
				return err
			}
		case "mute":
			if _, err := channel.SetMuted(true); err != nil {
				return err
			}
		case "unmute":
			if _, err := channel.SetMuted(false); err != nil {
				return err
			}
		case "volume_00":
			if _, err := channel.SetVolume(0.0); err != nil {
				return err
			}
		case "volume_25":
			if _, err := channel.SetVolume(0.25); err != nil {
				return err
			}
		case "volume_50":
			if _, err := channel.SetVolume(0.50); err != nil {
				return err
			}
		case "volume_75":
			if _, err := channel.SetVolume(0.75); err != nil {
				return err
			}
		case "volume_100":
			if _, err := channel.SetVolume(1.0); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Invalid command: should be mute, unmute, volume_DD, pause, play or stop")
		}
	}
	// success
	return nil
}

func Watch(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	googlecast := app.ModuleInstance("googlecast").(rpc.GoogleCast)
	start <- gopi.DONE

	// If there is an argument, then this is the service to lookup
	events := googlecast.Subscribe()
FOR_LOOP:
	for {
		select {
		case cmd := <-commands:
			if err := Execute(app, cmd); err != nil {
				app.Logger.Error("Error: %v", err)
				app.SendSignal()
			}
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
					if channel, err := googlecast.Connect(evt_.Device(), gopi.RPC_FLAG_INET_V4|gopi.RPC_FLAG_INET_V6, 0); err != nil {
						app.Logger.Error("Error: %v", err)
					} else {
						channels[evt_.Device().Id()] = channel
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

	// Wait for a second for the devices to be discovered
	time.Sleep(time.Second * 2)

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
