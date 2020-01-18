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
	"os"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi/v2"
	app "github.com/djthorpe/gopi/v2/app"

	// Units
	_ "github.com/djthorpe/gopi-rpc/v2/grpc/helloworld"
	_ "github.com/djthorpe/gopi-rpc/v2/unit/grpc"
	_ "github.com/djthorpe/gopi/v2/unit/logger"
	_ "github.com/djthorpe/gopi/v2/unit/mdns"
)

////////////////////////////////////////////////////////////////////////////////
// MAIN

func Main(app gopi.App, args []string) error {

	clientpool := app.UnitInstance("clientpool").(gopi.RPCClientPool)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	addr := app.Flags().GetString("addr", gopi.FLAG_NS_DEFAULT)
	defer cancel()
	if records, err := clientpool.Lookup(ctx, addr, 0); err != nil {
		return err
	} else if len(records) == 0 {
		return gopi.ErrNotFound.WithPrefix(addr)
	} else if conn, err := clientpool.Connect(records[0], 0); err != nil {
		return err
	} else {
		fmt.Println(conn)
	}

	fmt.Println("Press CTRL+C to exit")
	app.WaitForSignal(context.Background(), os.Interrupt)

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BOOTSTRAP

func main() {
	if app, err := app.NewCommandLineTool(Main, nil, "clientpool", "discovery"); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		app.Flags().FlagString("addr", "", "Service address or name")

		// Run and exit
		os.Exit(app.Run())
	}
}
