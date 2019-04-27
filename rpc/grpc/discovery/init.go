/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Service
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/discovery:service",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server", "rpc/discovery"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Service{
				Server:    app.ModuleInstance("rpc/server").(gopi.RPCServer),
				Discovery: app.ModuleInstance("rpc/discovery").(gopi.RPCServiceDiscovery),
			}, app.Logger)
		},
	})

	// Client
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/discovery:client",
		Type:     gopi.MODULE_TYPE_CLIENT,
		Requires: []string{"rpc/clientpool"},
		Run: func(app *gopi.AppInstance, _ gopi.Driver) error {
			if clientpool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool); clientpool == nil {
				return gopi.ErrAppError
			} else {
				clientpool.RegisterClient("gopi.Discovery", NewDiscoveryClient)
				return nil
			}
		},
	})
}
