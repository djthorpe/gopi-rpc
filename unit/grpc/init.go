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
		Type:     gopi.UNIT_RPC_SERVER,
		Requires: []string{"bus"},
		Config: func(app gopi.App) error {
			app.Flags().FlagUint("rpc.port", 0, "Server Port")
			app.Flags().FlagString("rpc.sslcert", "", "SSL Certificate Path")
			app.Flags().FlagString("rpc.sslkey", "", "SSL Key Path")
			return nil
		},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Server{
				Bus:            app.Bus(),
				Port:           app.Flags().GetUint("rpc.port", gopi.FLAG_NS_DEFAULT),
				SSLCertificate: app.Flags().GetString("rpc.sslcert", gopi.FLAG_NS_DEFAULT),
				SSLKey:         app.Flags().GetString("rpc.sslkey", gopi.FLAG_NS_DEFAULT),
			}, app.Log().Clone("grpc/server"))
		},
	})
}
