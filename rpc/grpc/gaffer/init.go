/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register service/helloworld:grpc
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/gaffer:service",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server", "gaffer"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Service{
				Server: app.ModuleInstance("rpc/server").(gopi.RPCServer),
				Gaffer: app.ModuleInstance("gaffer").(rpc.Gaffer),
			}, app.Logger)
		},
	})
}
