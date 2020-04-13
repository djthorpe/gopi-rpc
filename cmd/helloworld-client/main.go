/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
	app "github.com/djthorpe/gopi/v2/app"

	// Units
	_ "github.com/djthorpe/gopi-rpc/v2/grpc/helloworld"
	_ "github.com/djthorpe/gopi-rpc/v2/unit/grpc"
	_ "github.com/djthorpe/gopi/v2/unit/logger"
	_ "github.com/djthorpe/gopi/v2/unit/mdns"
)

////////////////////////////////////////////////////////////////////////////////
// CONNECT

func ParseAddr(endpoint string) (net.IP, uint16, error) {
	if host, port, err := net.SplitHostPort(endpoint); err != nil {
		return nil, 0, err
	} else if port_, err := strconv.ParseUint(port, 10, 32); err != nil {
		return nil, 0, err
	} else if addr := net.ParseIP(host); addr != nil {
		return addr, uint16(port_), nil
	} else if addrs, err := net.LookupHost(host); err != nil {
		return nil, 0, err
	} else if len(addrs) == 0 {
		return nil, 0, gopi.ErrNotFound
	} else if addr := net.ParseIP(addrs[0]); addr != nil {
		return addr, uint16(port_), nil
	} else {
		return nil, 0, gopi.ErrBadParameter
	}
}

func ConnectStub(app gopi.App, host string) (rpc.HelloworldStub, error) {
	clientpool := app.UnitInstance("clientpool").(gopi.RPCClientPool)

	if addr, port, err := ParseAddr(host); err != nil {
		return nil, gopi.ErrBadParameter.WithPrefix("-addr")
	} else if conn, err := clientpool.ConnectAddr(addr, port); err != nil {
		return nil, err
	} else if stub, ok := clientpool.CreateStub("gopi.Helloworld", conn).(rpc.HelloworldStub); ok == false {
		return nil, gopi.ErrInternalAppError
	} else {
		return stub, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// MAIN

func Main(app gopi.App, args []string) error {
	if addr := app.Flags().GetString("addr", gopi.FLAG_NS_DEFAULT); addr == "" {
		return fmt.Errorf("%w: Missing -addr flag", gopi.ErrBadParameter)
	} else if client, err := ConnectStub(app, addr); err != nil {
		return err
	} else if err := client.Ping(context.Background()); err != nil {
		return err
	} else {
		fmt.Println("Hello clent=", client)
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP

func main() {
	if app, err := app.NewCommandLineTool(Main, nil, "clientpool"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		app.Flags().FlagString("addr", "", "Service address or name")

		// Run and exit
		os.Exit(app.Run())
	}
}
