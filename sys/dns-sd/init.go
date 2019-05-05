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
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/rpcutil"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register InputManager
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/discovery",
		Requires: []string{"rpc/util"},
		Type:     gopi.MODULE_TYPE_DISCOVERY,
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagString("sd.iface", "", "Service Discovery Interface")
			config.AppFlags.FlagString("sd.domain", "local.", "Service Discovery Domain")
			config.AppFlags.FlagBool("sd.ip4", true, "Bind to IPv4 addresses")
			config.AppFlags.FlagBool("sd.ip6", true, "Bind to IPv6 addresses")
			config.AppFlags.FlagString("sd.cache", "", "Service cache file")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			domain, _ := app.AppFlags.GetString("sd.domain")
			name, _ := app.AppFlags.GetString("sd.iface")
			ip4, _ := app.AppFlags.GetBool("sd.ip4")
			ip6, _ := app.AppFlags.GetBool("sd.ip6")
			path, _ := app.AppFlags.GetString("sd.cache")
			if config, err := getDiscoveryConfig(domain, name, ip4, ip6); err != nil {
				return nil, err
			} else {
				config.Path = path
				config.Util = app.ModuleInstance("rpc/util").(rpc.Util)
				return gopi.Open(config, app.Logger)
			}
		},
	})
}

func getDiscoveryConfig(domain, net_iface_name string, ip4, ip6 bool) (Discovery, error) {
	config := Discovery{Domain: domain}
	if ip4 {
		config.Flags |= gopi.RPC_FLAG_INET_V4
	}
	if ip6 {
		config.Flags |= gopi.RPC_FLAG_INET_V6
	}
	if net_iface_name == "" {
		return config, nil
	}
	if ifaces, err := net.Interfaces(); err != nil {
		return config, err
	} else {
		iface_names := ""
		for _, iface := range ifaces {
			if iface.Name == net_iface_name {
				iface2 := iface
				config.Interface = &iface2
			}
			iface_names += "'" + iface.Name + "',"
		}
		if config.Interface == nil {
			return config, fmt.Errorf("Invalid -sd.iface flag (values: %v)", strings.Trim(iface_names, ","))
		}
		return config, nil
	}
}
