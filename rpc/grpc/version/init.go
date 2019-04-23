/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package version

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register service/helloworld:grpc
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/version:service",
		Type:     gopi.MODULE_TYPE_SERVICE,
		Requires: []string{"rpc/server"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Service{
				Server: app.ModuleInstance("rpc/server").(gopi.RPCServer),
				Flags:  app.AppFlags,
			}, app.Logger)
		},
	})
}
