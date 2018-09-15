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
	"fmt"
	"net"
	"strings"

	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name: "rpc/mdns",
		Type: gopi.MODULE_TYPE_DISCOVERY,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("dns-sd.iface", "", "DNS Service Discovery Interface")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			config := Listener{}
			if name, exists := app.AppFlags.GetString("dns-sd.iface"); exists {
				if ifaces, err := net.Interfaces(); err != nil {
					return nil, err
				} else {
					iface_names := ""
					for _, iface := range ifaces {
						if iface.Name == name {
							config.Interface = &iface
						}
						iface_names += "'" + iface.Name + "',"
					}
					if config.Interface == nil {
						return nil, fmt.Errorf("Invalid -dns-sd.iface flag (values: %v)", strings.Trim(iface_names, ","))
					}
				}
			}
			return gopi.Open(config, app.Logger)
		},
	})
}
