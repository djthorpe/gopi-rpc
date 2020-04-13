/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Kernel
	gopi.UnitRegister(gopi.UnitConfig{
		Name: Kernel{}.Name(),
		Config: func(app gopi.App) error {
			app.Flags().FlagString("gaffer.root", "", "Root folder for executables")
			return nil
		},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Kernel{
				Root: app.Flags().GetString("gaffer.root", gopi.FLAG_NS_DEFAULT),
			}, app.Log().Clone(Kernel{}.Name()))
		},
	})

	// Service
	gopi.UnitRegister(gopi.UnitConfig{
		Name:     Service{}.Name(),
		Requires: []string{"clientpool"},
		Config: func(app gopi.App) error {
			app.Flags().FlagString("gaffer.fifo", "", "Unix socket connection to kernel")
			return nil
		},
		New: func(app gopi.App) (gopi.Unit, error) {
			return gopi.New(Service{
				Fifo:       app.Flags().GetString("gaffer.fifo", gopi.FLAG_NS_DEFAULT),
				Clientpool: app.UnitInstance("clientpool").(gopi.RPCClientPool),
			}, app.Log().Clone(Service{}.Name()))
		},
	})
}
