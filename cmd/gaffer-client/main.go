/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
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

const (
	DISCOVERY_TIMEOUT = 250 * time.Millisecond
)

var (
	runner *Runner
)

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, records []gopi.RPCServiceRecord, done chan<- struct{}) error {
	if gaffer, err := app.ClientPool.NewClientEx("gopi.Gaffer", records, gopi.RPC_FLAG_NONE); err != nil {
		return err
	} else if discovery, err := app.ClientPool.NewClientEx("gopi.Discovery", records, gopi.RPC_FLAG_NONE); err != nil {
		return err
	} else {
		return runner.Run(gaffer.(rpc.GafferClient), discovery.(rpc.DiscoveryClient), app.AppFlags)
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/gaffer:client", "rpc/discovery:client", "discovery")

	// Create the runner
	runner = NewRunner()

	// Set usage
	config.AppFlags.SetUsageFunc(runner.Usage)

	// Set flags
	config.AppFlags.FlagBool("dns", false, "Use DNS for service discovery")

	// Run the command line tool
	os.Exit(rpc.Client(config, DISCOVERY_TIMEOUT, Main))
}
