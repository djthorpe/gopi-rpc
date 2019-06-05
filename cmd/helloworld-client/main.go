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
	"fmt"
	"os"
	"strconv"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Services
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/helloworld"
)

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, services []gopi.RPCServiceRecord, done chan<- struct{}) error {

	// Get the name argument
	name, _ := app.AppFlags.GetString("name")

	// Run the RPC for greeter
	if stub, err := app.ClientPool.NewClientEx("gopi.Greeter", services, gopi.RPC_FLAG_SERVICE_ANY); err != nil {
		return err
	} else if stub_ := stub.(rpc.GreeterClient); stub_ != nil {
		if err := stub_.Ping(); err != nil {
			return err
		} else if reply, err := stub_.SayHello(name); err != nil {
			return err
		} else {
			fmt.Println("Service says:", strconv.Quote(reply))
		}
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/helloworld:client", "discovery")

	// Set flags
	config.AppFlags.FlagString("name", "", "Your name")

	// Run the command line tool
	os.Exit(rpc.Client(config, 200*time.Millisecond, Main))
}
