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
		Name:     Service{}.Name(),
		Type:     gopi.UNIT_RPC_SERVICE,
		Requires: []string{"server"},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Service{
				Server: app.UnitInstance("server").(gopi.RPCServer),
			}, app.Log().Clone(Service{}.Name()))
		},
	})
	gopi.UnitRegister(gopi.UnitConfig{
		Name: Client{}.Name(),
		Type: gopi.UNIT_RPC_CLIENT,
		Stub: func(conn gopi.RPCClientConn) (gopi.RPCClientStub, error) {
			if unit, err := gopi.New(Client{Conn: conn}, nil); err != nil {
				return nil, err
			} else {
				return unit.(gopi.RPCClientStub), nil
			}
		},
	})
}
