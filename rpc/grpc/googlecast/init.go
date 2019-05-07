/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Service
	gopi.RegisterModule(gopi.Module{
		Name:     "googlecast:service",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server", "googlecast"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Service{
				Server:     app.ModuleInstance("rpc/server").(gopi.RPCServer),
				GoogleCast: app.ModuleInstance("googlecast").(rpc.GoogleCast),
			}, app.Logger)
		},
	})

	// Client
	gopi.RegisterModule(gopi.Module{
		Name:     "googlecast:client",
		Type:     gopi.MODULE_TYPE_CLIENT,
		Requires: []string{"rpc/clientpool"},
		Run: func(app *gopi.AppInstance, _ gopi.Driver) error {
			if clientpool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool); clientpool == nil {
				return gopi.ErrAppError
			} else {
				clientpool.RegisterClient("gopi.GoogleCast", NewGoogleCastClient)
				return nil
			}
		},
	})
}
