/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	gopi "github.com/djthorpe/gopi/v2"
)

func init() {
	gopi.UnitRegister(gopi.UnitConfig{
		Type:     gopi.UNIT_RPC_SERVER,
		Name:     Server{}.Name(),
		Requires: []string{"bus"},
		Config: func(app gopi.App) error {
			app.Flags().FlagString("rpc.fifo", "", "Server Fifo file")
			app.Flags().FlagString("rpc.fifogid", "", "Server Fifo group")
			app.Flags().FlagUint("rpc.port", 0, "Server Port")
			app.Flags().FlagString("rpc.sslcert", "", "SSL Certificate Path")
			app.Flags().FlagString("rpc.sslkey", "", "SSL Key Path")
			return nil
		},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Server{
				Bus:            app.Bus(),
				File:           app.Flags().GetString("rpc.fifo", gopi.FLAG_NS_DEFAULT),
				FileGroup:      app.Flags().GetString("rpc.fifogid", gopi.FLAG_NS_DEFAULT),
				Port:           app.Flags().GetUint("rpc.port", gopi.FLAG_NS_DEFAULT),
				SSLCertificate: app.Flags().GetString("rpc.sslcert", gopi.FLAG_NS_DEFAULT),
				SSLKey:         app.Flags().GetString("rpc.sslkey", gopi.FLAG_NS_DEFAULT),
			}, app.Log().Clone(Server{}.Name()))
		},
	})
	gopi.UnitRegister(gopi.UnitConfig{
		Type: gopi.UNIT_RPC_CLIENTPOOL,
		Name: ClientPool{}.Name(),
		Config: func(app gopi.App) error {
			app.Flags().FlagDuration("rpc.timeout", 0, "Connection timeout")
			app.Flags().FlagBool("rpc.insecure", false, "Allow plaintext connections")
			app.Flags().FlagBool("rpc.skipverify", true, "Skip SSL certificate verification")
			return nil
		},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(ClientPool{}, app.Log().Clone(ClientPool{}.Name()))
		},
		Run: func(app gopi.App, instance gopi.Unit) error {
			if discovery, ok := app.UnitInstance("discovery").(gopi.RPCServiceDiscovery); discovery != nil && ok {
				instance.(*clientpool).discovery = discovery
			}
			//for _, module := range gopi.UnitsByType(gopi.UNIT_RPC_CLIENT) {
			//	// Register
			//}
			return nil
		},
	})
}
