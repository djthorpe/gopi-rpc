/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

// An example RPC Client tool
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	"github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/gaffer-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Services
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/helloworld"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/version"
)

////////////////////////////////////////////////////////////////////////////////

func RunVersion(app *gopi.AppInstance, conn gopi.RPCClientConn) error {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	if client_ := pool.NewClient("gopi.Version", conn); client_ == nil {
		return gopi.ErrAppError
	} else if client, ok := client_.(rpc.VersionClient); ok == false {
		return gopi.ErrAppError
	} else if err := client.Ping(); err != nil {
		return err
	} else if params, err := client.Version(); err != nil {
		return err
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Key", "Value"})
		for k, v := range params {
			table.Append([]string{
				k, v,
			})
		}
		table.Render()
	}
	return nil
}

func RunHelloworld(app *gopi.AppInstance, conn gopi.RPCClientConn) error {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	name, _ := app.AppFlags.GetString("name")
	if client_ := pool.NewClient("gopi.Greeter", conn); client_ == nil {
		return gopi.ErrAppError
	} else if client, ok := client_.(rpc.GreeterClient); ok == false {
		return gopi.ErrAppError
	} else if err := client.Ping(); err != nil {
		return err
	} else if reply, err := client.SayHello(name); err != nil {
		return err
	} else {
		fmt.Println("Service says:", strconv.Quote(reply))
	}
	return nil
}

func Conn(app *gopi.AppInstance) (gopi.RPCClientConn, error) {
	// Return a single connection
	addr, _ := app.AppFlags.GetString("addr")
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
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

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if conn, err := Conn(app); err != nil {
		return err
	} else if services, err := conn.Services(); err != nil {
		return err
	} else if args := app.AppFlags.Args(); len(args) == 0 {
		if err := RunHelloworld(app, conn); err != nil {
			return err
		}
	} else if len(args) == 1 && args[0] == "version" {
		if err := RunVersion(app, conn); err != nil {
			return err
		}
	} else if len(args) == 1 && args[0] == "services" {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Service"})
		for _, s := range services {
			table.Append([]string{
				s,
			})
		}
		table.Render()
	} else {
		return fmt.Errorf("Invoke the command with zero or one argument (which can be 'version' or 'services')")
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func Conn2(app *gopi.AppInstance) (gopi.RPCServiceRecord, error) {
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

func Main2(app *gopi.AppInstance, done chan<- struct{}) error {
	if record, err := Conn2(app); err != nil {
		return err
	} else {
		fmt.Println(record)
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/helloworld:client", "rpc/version:client", "discovery")

	// Set flags
	config.AppFlags.FlagString("name", "", "Your name")
	config.AppFlags.FlagString("addr", "", "Service name or gateway address")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main2))
}
