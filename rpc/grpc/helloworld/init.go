/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package helloworld

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
)

func init() {
	gopi.UnitRegister(gopi.UnitConfig{
		Name:     "rpc/helloworld/service",
		Type:     gopi.UNIT_RPC_SERVICE,
		Requires: []string{"server"},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Service{}, app.Log().Clone("rpc/helloworld/service"))
		},
	})
}
