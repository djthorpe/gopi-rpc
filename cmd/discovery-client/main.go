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
	rpc "github.com/djthorpe/gopi-rpc"
	tablewriter "github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/discovery"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/googlecast"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/version"
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////

func RenderHost(service gopi.RPCServiceRecord) string {
	if service.Port() == 0 {
		return service.Host()
	} else {
		return fmt.Sprintf("%v:%v", service.Host(), service.Port())
	}
}

func RenderService(service gopi.RPCServiceRecord) string {
	if service.Subtype() == "" {
		return service.Service()
	} else {
		return fmt.Sprintf("%v, %v", service.Subtype(), service.Service())
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
	return strings.Join(ips, "\n")
}

func RenderTxt(service gopi.RPCServiceRecord) string {
	return strings.Join(service.Text(), "\n")
}

func Conn(app *gopi.AppInstance) (gopi.RPCClientConn, error) {
	// Return a single connection
	addr, _ := app.AppFlags.GetString("addr")
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	// If service is nil, then use the address
	if records, err := pool.Lookup(ctx, "", addr, 1); err != nil {
		return nil, err
	} else if len(records) == 0 {
		return nil, gopi.ErrDeadlineExceeded
	} else if conn, err := pool.Connect(records[0], 0); err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

func DiscoveryClient(app *gopi.AppInstance) (rpc.DiscoveryClient, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := Conn(app); err != nil {
		return nil, err
	} else if client := pool.NewClient("gopi.Discovery", conn); client == nil {
		return nil, gopi.ErrNotFound
	} else if client_, ok := client.(rpc.DiscoveryClient); client_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client_, nil
	}
}

func VersionClient(app *gopi.AppInstance) (rpc.VersionClient, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := Conn(app); err != nil {
		return nil, err
	} else if client := pool.NewClient("gopi.Version", conn); client == nil {
		return nil, gopi.ErrNotFound
	} else if client_, ok := client.(rpc.VersionClient); client_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client_, nil
	}
}

func GoogleCastClient(app *gopi.AppInstance) (rpc.GoogleCastClient, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := Conn(app); err != nil {
		return nil, err
	} else if client := pool.NewClient("gopi.GoogleCast", conn); client == nil {
		return nil, gopi.ErrNotFound
	} else if client_, ok := client.(rpc.GoogleCastClient); client_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client_, nil
	}
}

func Watch(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	start <- gopi.DONE
	service := ""

	if watch, _ := app.AppFlags.GetBool("watch"); watch == false {
		return nil
	}
	if args := app.AppFlags.Args(); len(args) == 1 {
		service = args[0]
	}

	events := make(chan gopi.RPCEvent)
	go func() {
		fmt.Println("START")
	FOR_LOOP:
		for {
			select {
			case evt := <-events:
				fmt.Println(evt)
			case <-stop:
				break FOR_LOOP
			}
		}
		fmt.Println("STOP")
	}()

	if client, err := DiscoveryClient(app); err != nil {
		return err
	} else if err := client.StreamEvents(service, events); err != nil {
		return err
	}

	// Success
	return nil
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	t := rpc.DISCOVERY_TYPE_DB
	service := ""
	if dns, _ := app.AppFlags.GetBool("dns"); dns {
		t = rpc.DISCOVERY_TYPE_DNS
	}
	if args := app.AppFlags.Args(); len(args) == 1 {
		service = args[0]
	} else if len(args) > 1 {
		return gopi.ErrHelp
	}

	if client, err := DiscoveryClient(app); err != nil {
		return err
	} else if err := client.Ping(); err != nil {
		return err
	} else if service == "" {
		if services, err := client.Enumerate(t, 0); err != nil {
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
		if instances, err := client.Lookup(service, t, 0); err != nil {
			return err
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Service", "Name", "Host", "Addr", "TXT"})
			for _, instance := range instances {
				table.Append([]string{
					RenderService(instance),
					instance.Name(),
					RenderHost(instance),
					RenderIP(instance),
					RenderTxt(instance),
				})
			}
			table.Render()
		}
	}

	if client, err := GoogleCastClient(app); err != nil {
		return err
	} else if err := client.Ping(); err != nil {
		return err
	} else {
		if devices, err := client.Devices(); err != nil {
			return err
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Service", "State", "Model"})
			for _, device := range devices {
				table.Append([]string{
					device.Name(),
					device.Service(),
					fmt.Sprint(device.State()),
					device.Model(),
				})
			}
			table.Render()
		}
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/discovery:client", "rpc/version:client", "googlecast:client", "discovery")

	// Set flags
	config.AppFlags.FlagString("addr", "localhost:8080", "Gateway address")
	config.AppFlags.FlagBool("dns", false, "Use DNS lookup rather than cache")
	config.AppFlags.FlagBool("watch", false, "Watch for discovery changes")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main, Watch))
}
