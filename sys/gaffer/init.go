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
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name:     "gaffer",
		Type:     gopi.MODULE_TYPE_OTHER,
		Requires: []string{"rpc/util"},
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("gaffer.path", "", "Gaffer Database File")
			config.AppFlags.FlagString("gaffer.root", "", "Gaffer Binary Root")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			path, _ := app.AppFlags.GetString("gaffer.path")
			binroot, binoverride := app.AppFlags.GetString("gaffer.root")
			return gopi.Open(Gaffer{
				Path:        path,
				BinRoot:     binroot,
				BinOverride: binoverride,
				Util:        app.ModuleInstance("rpc/util").(rpc.Util),
			}, app.Logger)
		},
	})
}
