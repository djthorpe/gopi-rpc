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
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Services
	discovery "github.com/djthorpe/gopi-rpc/rpc/grpc/discovery"
	version "github.com/djthorpe/gopi-rpc/rpc/grpc/version"
)

////////////////////////////////////////////////////////////////////////////////

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

func DiscoveryClient(app *gopi.AppInstance) (*discovery.Client, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := Conn(app); err != nil {
		return nil, err
	} else if client := pool.NewClient("gopi.Discovery", conn); client == nil {
		return nil, gopi.ErrNotFound
	} else if client_, ok := client.(*discovery.Client); client_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client_, nil
	}
}

func VersionClient(app *gopi.AppInstance) (*version.Client, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := Conn(app); err != nil {
		return nil, err
	} else if client := pool.NewClient("gopi.Version", conn); client == nil {
		return nil, gopi.ErrNotFound
	} else if client_, ok := client.(*version.Client); client_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client_, nil
	}
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	if client, err := DiscoveryClient(app); err != nil {
		return err
	} else if err := client.Ping(); err != nil {
		return err
	} else {
		fmt.Println("Conn=", client)
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/discovery:client", "rpc/version:client")

	// Set flags
	config.AppFlags.FlagString("addr", "localhost:8080", "Gateway address")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
