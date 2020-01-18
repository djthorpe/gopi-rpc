/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
)

func init() {
	gopi.UnitRegister(gopi.UnitConfig{
		Type: gopi.UNIT_RPC_SERVER,
		Config: func(app gopi.App) error {
			app.Flags().FlagUint("rpc.port", 0, "Server Port")
			app.Flags().FlagString("rpc.sslcert", "", "SSL Certificate Path")
			app.Flags().FlagString("rpc.sslkey", "", "SSL Key Path")
			return nil
		},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Server{}, app.Log().Clone("grpc/server"))
		},
	})
}
