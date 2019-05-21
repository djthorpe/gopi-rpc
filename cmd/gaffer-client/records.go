/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
	"fmt"
	"os"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////

func RecordCommands(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	if len(args) != 1 {
		return gopi.ErrHelp
	} else if records, err := discovery.Lookup(args[0], rpc.DISCOVERY_TYPE_DB, time.Second); err != nil {
		return err
	} else if len(records) == 0 {
		return fmt.Errorf("No service records")
	} else {
		return OutputRecords(os.Stdout, records)
	}

	// Success
	return nil
}
