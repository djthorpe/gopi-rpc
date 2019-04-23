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
	"os/user"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// RPC Clients
	hw "github.com/djthorpe/gopi-rpc/rpc/grpc/helloworld"
)

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, done chan<- struct{}) error {

	// Client Pool
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	addr, _ := app.AppFlags.GetString("addr")

	// Lookup any service record for application
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	if records, err := pool.Lookup(ctx, "", addr, 1); err != nil {
		done <- gopi.DONE
		return err
	} else if len(records) == 0 {
		done <- gopi.DONE
		return gopi.ErrDeadlineExceeded
	} else if conn, err := pool.Connect(records[0], 0); err != nil {
		done <- gopi.DONE
		return err
	} else if err := RunHelloworld(app, conn); err != nil {
		done <- gopi.DONE
		return err
	} else if err := pool.Disconnect(conn); err != nil {
		done <- gopi.DONE
		return err
	}

	// Success
	done <- gopi.DONE
	return nil
}

func RunHelloworld(app *gopi.AppInstance, conn gopi.RPCClientConn) error {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	name, _ := app.AppFlags.GetString("name")
	if client_ := pool.NewClient("gopi.Helloworld", conn); client_ == nil {
		return gopi.ErrAppError
	} else if client, ok := client_.(*hw.Client); ok == false {
		return gopi.ErrAppError
	} else if message, err := client.SayHello(name); err != nil {
		return err
	} else {
		fmt.Printf("%v says '%v'\n\n", conn.Name(), message)
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/client/helloworld")

	if cur, err := user.Current(); err == nil {
		config.AppFlags.FlagString("name", cur.Name, "Your name")
	} else {
		config.AppFlags.FlagString("name", "", "Your name")
	}
	config.AppFlags.FlagString("addr", "", "Gateway address")

	// Run the command line tool
	os.Exit(gopi.CommandLineTool(config, Main))
}
