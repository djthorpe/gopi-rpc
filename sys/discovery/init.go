/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"fmt"
	"net"
	"strings"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name: "rpc/discovery",
		Type: gopi.MODULE_TYPE_DISCOVERY,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("dns-sd.iface", "", "Service Discovery Interface")
			config.AppFlags.FlagString("dns-sd.domain", "local.", "Service Discovery Domain")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			domain, _ := app.AppFlags.GetString("dns-sd.domain")
			name, _ := app.AppFlags.GetString("dns-sd.iface")
			if config, err := GetConfig(domain, name); err != nil {
				return nil, err
			} else {
				return gopi.Open(config, app.Logger)
			}
		},
	})
}

func GetConfig(domain, iface_name string) (Discovery, error) {
	config := Discovery{Domain: domain}
	if iface_name == "" {
		return config, nil
	}
	if ifaces, err := net.Interfaces(); err != nil {
		return config, err
	} else {
		iface_names := ""
		for _, iface := range ifaces {
			if iface.Name == iface_name {
				config.Interface = &iface
			}
			iface_names += "'" + iface.Name + "',"
		}
		if config.Interface == nil {
			return config, fmt.Errorf("Invalid -dns-sd.iface flag (values: %v)", strings.Trim(iface_names, ","))
		}
		return config, nil
	}
}
