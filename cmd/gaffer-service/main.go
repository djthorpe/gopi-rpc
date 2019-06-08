/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/gaffer"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Services
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/discovery"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/gaffer"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/version"
)

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/gaffer:service", "rpc/version:service", "rpc/discovery:service")

	// Subtype
	config.AppFlags.SetParam(gopi.PARAM_SERVICE_SUBTYPE, "gaffer")

	// Run the server and register all the services
	os.Exit(rpc.Server(config))
}
