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
	"os"
	"strconv"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	"github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Services
	hw "github.com/djthorpe/gopi-rpc/rpc/grpc/helloworld"
	version "github.com/djthorpe/gopi-rpc/rpc/grpc/version"
)

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if conn, err := Conn(app); err != nil {
		return err
	} else if services, err := conn.Services(); err != nil {
		return err
	} else {
		app.Logger.Info("conn=%v", conn)
		app.Logger.Info("services=%v", services)
		if err := RunVersion(app, conn); err != nil {
			return err
		}
		if err := RunHelloworld(app, conn); err != nil {
			return err
		}
	}

	// Success
	return nil
}

func RunVersion(app *gopi.AppInstance, conn gopi.RPCClientConn) error {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	if client_ := pool.NewClient("gopi.Version", conn); client_ == nil {
		return gopi.ErrAppError
	} else if client, ok := client_.(*version.Client); ok == false {
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
	} else if client, ok := client_.(*hw.Client); ok == false {
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

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/helloworld:client", "rpc/version:client")

	// Set flags
	config.AppFlags.FlagString("name", "", "Your name")
	config.AppFlags.FlagString("addr", "localhost:8080", "Gateway address")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
