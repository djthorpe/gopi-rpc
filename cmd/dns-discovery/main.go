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
	"os"
	"strings"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	tablewriter "github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////

func RenderEvent(evt gopi.RPCEvent) {
	t := strings.TrimPrefix(fmt.Sprint(evt.Type()), "RPC_EVENT_SERVICE_")
	s := evt.ServiceRecord()
	if evt.Type() == gopi.RPC_EVENT_SERVICE_NAME {
		fmt.Printf("%-10s %-30s\n", t, s.Name())
	} else if s.Port() == 0 {
		fmt.Printf("%-10s %-30s %s\n", t, s.Service(), s.Name())
	} else {
		fmt.Printf("%-10s %-30s %s (%s:%v)\n", t, s.Service(), s.Name(), s.Host(), s.Port())
	}
}

func RenderHost(service gopi.RPCServiceRecord) string {
	if service.Port() == 0 {
		return service.Host()
	} else {
		return fmt.Sprintf("%v:%v", service.Host(), service.Port())
	}
}

func RenderIP(service gopi.RPCServiceRecord) string {
	ips := make([]string, 0)
	for _, ip := range service.IP4() {
		ips = append(ips, fmt.Sprint(ip))
	}
	for _, ip := range service.IP6() {
		ips = append(ips, fmt.Sprint(ip))
	}
	return strings.Join(ips, " ")
}

func Lookup(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	discovery := app.ModuleInstance("discovery").(gopi.RPCServiceDiscovery)
	start <- gopi.DONE

	// If there is an argument, then this is the service to lookup
	args := app.AppFlags.Args()
	watch, _ := app.AppFlags.GetBool("watch")
	service := ""
	if len(args) == 1 {
		service = args[0]
	} else if len(args) > 1 {
		return gopi.ErrHelp
	} else if len(args) == 0 && watch {
		events := discovery.Subscribe()
	FOR_LOOP:
		for {
			select {
			case evt := <-events:
				if evt_, ok := evt.(gopi.RPCEvent); ok {
					RenderEvent(evt_)
				}
			case <-stop:
				break FOR_LOOP
			}
		}
		discovery.Unsubscribe(events)
		return nil
	}

	// Create context with timeout
	timeout, _ := app.AppFlags.GetDuration("timeout")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// Wait for cancel in the background
	go func() {
		<-stop
		cancel()
	}()

	if service == "" {
		// Enumerate service names
		if services, err := discovery.EnumerateServices(ctx); err != nil {
			return err
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Service"})
			for _, service := range services {
				table.Append([]string{service})
			}
			table.Render()
		}

	} else {
		// Perform lookup
		if services, err := discovery.Lookup(ctx, args[0]); err != nil {
			return err
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Service", "Name", "Host", "IP"})
			for _, service := range services {
				table.Append([]string{
					service.Service(),
					service.Name(),
					RenderHost(service),
					RenderIP(service),
				})
			}
			table.Render()
		}
	}

	// Signal end
	return app.SendSignal()
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
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
	config.AppFlags.FlagDuration("timeout", time.Millisecond*750, "Timeout for lookup")
	config.AppFlags.FlagBool("watch", false, "Watch for stream of service records")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Lookup))
}
