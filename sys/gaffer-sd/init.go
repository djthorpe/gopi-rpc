/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/discovery"
	_ "github.com/djthorpe/gopi-rpc/sys/rpcutil"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/discovery:gaffer",
		Type:     gopi.MODULE_TYPE_DISCOVERY,
		Requires: []string{"rpc/discovery:client", "rpc/clientpool"},
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("sd.addr", "", "Discovery Service Address")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			addr, _ := app.AppFlags.GetString("sd.addr")
			return gopi.Open(Gaffer{
				Addr: addr,
				Pool: app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool),
			}, app.Logger)
		},
	})
}
