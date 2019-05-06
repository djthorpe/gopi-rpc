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
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name:     "googlecast",
		Type:     gopi.MODULE_TYPE_OTHER,
		Requires: []string{"discovery"},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(GoogleCast{
				Discovery: app.ModuleInstance("discovery").(gopi.RPCServiceDiscovery),
			}, app.Logger)
		},
	})
}
