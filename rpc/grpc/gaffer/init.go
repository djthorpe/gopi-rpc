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
	// Service "rpc/gaffer:service"
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

	// Client Stub "rpc/gaffer:client"
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/gaffer:client",
		Type:     gopi.MODULE_TYPE_CLIENT,
		Requires: []string{"rpc/clientpool"},
		Run: func(app *gopi.AppInstance, _ gopi.Driver) error {
			if clientpool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool); clientpool == nil {
				return gopi.ErrAppError
			} else {
				clientpool.RegisterClient("gopi.Gaffer", NewGafferClient)
				return nil
			}
		},
	})
}
