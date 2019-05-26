/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Services
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/discovery"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/gaffer"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/version"
)

const (
	DISCOVERY_TIMEOUT = 250 * time.Millisecond
)

////////////////////////////////////////////////////////////////////////////////

func Conn(app *gopi.AppInstance) (gopi.RPCServiceRecord, error) {
	addr, _ := app.AppFlags.GetString("addr")
	timeout, exists := app.AppFlags.GetDuration("rpc.timeout")
	if exists == false {
		timeout = DISCOVERY_TIMEOUT
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if service, _, _, err := app.Service(); err != nil {
		return nil, err
	} else {
		service_ := fmt.Sprintf("_%v._tcp", service)
		pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
		if services, err := pool.Lookup(ctx, service_, addr, 0); err != nil {
			return nil, err
		} else if len(services) == 0 {
			return nil, gopi.ErrNotFound
		} else if len(services) > 1 {
			var names []string
			for _, service := range services {
				names = append(names, strconv.Quote(service.Name()))
			}
			return nil, fmt.Errorf("More than one service returned, use -addr to choose between %v", strings.Join(names, ","))
		} else {
			return services[0], nil
		}
	}
}

func GafferStub(app *gopi.AppInstance, sr gopi.RPCServiceRecord) (rpc.GafferClient, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	if sr == nil || pool == nil {
		return nil, gopi.ErrBadParameter
	} else if conn, err := pool.Connect(sr, 0); err != nil {
		return nil, err
	} else if stub := pool.NewClient("gopi.Gaffer", conn); stub == nil {
		return nil, gopi.ErrBadParameter
	} else if stub_, ok := stub.(rpc.GafferClient); ok == false {
		return nil, fmt.Errorf("Stub is not an rpc.GafferClient")
	} else if err := stub_.Ping(); err != nil {
		return nil, err
	} else {
		return stub_, nil
	}
}

func DiscoveryStub(app *gopi.AppInstance, sr gopi.RPCServiceRecord) (rpc.DiscoveryClient, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	if sr == nil || pool == nil {
		return nil, gopi.ErrBadParameter
	} else if conn, err := pool.Connect(sr, 0); err != nil {
		return nil, err
	} else if stub := pool.NewClient("gopi.Discovery", conn); stub == nil {
		return nil, gopi.ErrBadParameter
	} else if stub_, ok := stub.(rpc.DiscoveryClient); ok == false {
		return nil, fmt.Errorf("Stub is not an rpc.DiscoveryClient")
	} else if err := stub_.Ping(); err != nil {
		return nil, err
	} else {
		return stub_, nil
	}
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if record, err := Conn(app); err != nil {
		return err
	} else if gaffer_client, err := GafferStub(app, record); err != nil {
		return err
	} else if discovery_client, err := DiscoveryStub(app, record); err != nil {
		return err
	} else if err := Run(app, gaffer_client, discovery_client); err != nil {
		return err
	}

	// Success
	return nil
}

func Usage(flags *gopi.Flags) {
	fh := os.Stdout

	fmt.Fprintf(fh, "%v: Microservice Manager\nhttps://github.com/djthorpe/gopi-rpc/\n\n", flags.Name())
	fmt.Fprintf(fh, "Syntax:\n\n")
	fmt.Fprintf(fh, "  %v (<flags>...) <service|type|@group|instance> (<command>) (<argument>...)\n\n", flags.Name())
	fmt.Fprintf(fh, "Commands:\n\n")
	for _, command := range root_commands {
		fmt.Fprintf(fh, "  %v %v\n", flags.Name(), command.Name)
		fmt.Fprintf(fh, "        %v\n", command.Description)
		fmt.Fprintf(fh, "\n")
	}
	fmt.Fprintf(fh, "Command line flags:\n\n")
	flags.PrintDefaults()
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/gaffer:client", "rpc/discovery:client", "discovery")

	// Set usage
	config.AppFlags.SetUsageFunc(Usage)

	// Set flags
	config.AppFlags.FlagString("addr", "", "Service name or gateway address")
	config.AppFlags.FlagBool("dns", false, "Use DNS for service discovery")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
