/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
	"fmt"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////

func ServiceCommands(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	// Obtain the group name
	service := reService.FindStringSubmatch(args[0])
	if len(service) != 2 {
		return gopi.ErrBadParameter
	}

	fmt.Println(service[1])
	return gopi.ErrNotImplemented

	// Success
	return nil
}
