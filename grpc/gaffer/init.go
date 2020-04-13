/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
)

func init() {

	// Kernel
	gopi.UnitRegister(gopi.UnitConfig{
		Name:     KernelService{}.Name(),
		Type:     gopi.UNIT_RPC_SERVICE,
		Requires: []string{"server", "gaffer/kernel"},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(KernelService{
				Server: app.UnitInstance("server").(gopi.RPCServer),
				Kernel: app.UnitInstance("gaffer/kernel").(rpc.GafferKernel),
			}, app.Log().Clone(KernelService{}.Name()))
		},
	})

	// Service
	gopi.UnitRegister(gopi.UnitConfig{
		Name:     GafferService{}.Name(),
		Type:     gopi.UNIT_RPC_SERVICE,
		Requires: []string{"server", "gaffer/service"},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(GafferService{
				Server: app.UnitInstance("server").(gopi.RPCServer),
				Gaffer: app.UnitInstance("gaffer/service").(rpc.Gaffer),
			}, app.Log().Clone(GafferService{}.Name()))
		},
	})

	// Client
	gopi.UnitRegister(gopi.UnitConfig{
		Name: KernelClient{}.Name(),
		Type: gopi.UNIT_RPC_CLIENT,
		Stub: func(conn gopi.RPCClientConn) (gopi.RPCClientStub, error) {
			if unit, err := gopi.New(KernelClient{Conn: conn}, nil); err != nil {
				return nil, err
			} else {
				return unit.(gopi.RPCClientStub), nil
			}
		},
	})
}
