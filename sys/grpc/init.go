/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package grpc

import (
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	grpc "google.golang.org/grpc"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/rpcutil"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	// Register GRPC rpc/server
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/server",
		Type:     gopi.MODULE_TYPE_OTHER,
		Requires: []string{"rpc/util"},
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagUint("rpc.port", 0, "Server Port")
			config.AppFlags.FlagString("rpc.sslcert", "", "SSL Certificate Path")
			config.AppFlags.FlagString("rpc.sslkey", "", "SSL Key Path")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			port, _ := app.AppFlags.GetUint("rpc.port")
			key, _ := app.AppFlags.GetString("rpc.sslkey")
			cert, _ := app.AppFlags.GetString("rpc.sslcert")
			return gopi.Open(Server{
				Port:           port,
				SSLCertificate: cert,
				SSLKey:         key,
				ServerOption:   []grpc.ServerOption{},
				Util:           app.ModuleInstance("rpc/util").(rpc.Util),
			}, app.Logger)
		},
	})

	// Register GRPC rpc/clientpool module
	gopi.RegisterModule(gopi.Module{
		Name:     "rpc/clientpool",
		Type:     gopi.MODULE_TYPE_OTHER,
		Requires: []string{"rpc/util"},
		Config: func(config *gopi.AppConfig) {
			config.AppFlags.FlagDuration("rpc.timeout", 5*time.Second, "Connection timeout")
			config.AppFlags.FlagBool("rpc.insecure", false, "Allow plaintext connections")
			config.AppFlags.FlagBool("rpc.skipverify", true, "Skip SSL certificate verification")
		},
		New: func(app *gopi.AppInstance) (gopi.Driver, error) {
			insecure, _ := app.AppFlags.GetBool("rpc.insecure")
			skipverify, _ := app.AppFlags.GetBool("rpc.skipverify")
			timeout, _ := app.AppFlags.GetDuration("rpc.timeout")
			config := ClientPool{
				SSL:        (insecure == false),
				SkipVerify: skipverify,
				Timeout:    timeout,
				Util:       app.ModuleInstance("rpc/util").(rpc.RPCUtil),
			}
			return gopi.Open(config, app.Logger)
		},
		Run: func(app *gopi.AppInstance, driver gopi.Driver) error {
			// Hook in the discovery module if it's found
			if discovery := app.ModuleInstance("discovery"); discovery != nil {
				driver.(*clientpool).discovery = discovery.(gopi.RPCServiceDiscovery)
			}
			return nil
		},
	})
}
