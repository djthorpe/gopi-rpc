/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2017
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package mdns

import (
	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name: "rpc/discovery",
		Type: gopi.MODULE_TYPE_MDNS,
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			return gopi.Open(Listener{}, app.Logger)
		},
	})
}
