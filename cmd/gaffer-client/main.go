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
	"net"
	"os"
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

////////////////////////////////////////////////////////////////////////////////

func Conn(app *gopi.AppInstance) (gopi.RPCServiceRecord, error) {
	addr, _ := app.AppFlags.GetString("addr")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if addr == "" {
		if addr_, _, _, err := app.Service(); err != nil {
			return nil, err
		} else {
			addr = addr_
		}
	}

	if _, _, err := net.SplitHostPort(addr); err == nil {
		if service, err := pool.Lookup(ctx, "", addr, 1); err != nil {
			return nil, err
		} else {
			return service[0], nil
		}
	} else if service, err := pool.Lookup(ctx, fmt.Sprintf("_%v._tcp", addr), "", 1); err != nil {
		return nil, err
	} else {
		return service[0], nil
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
	config := gopi.NewAppConfig("rpc/gaffer:client", "rpc/discovery:client")

	// Set usage
	config.AppFlags.SetUsageFunc(Usage)

	// Set flags
	config.AppFlags.FlagString("addr", "", "Service name or gateway address")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool2(config, Main))
}
