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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////

func ServiceCommands(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	// Obtain the service name
	service := reService.FindStringSubmatch(args[0])
	if len(service) != 2 {
		return gopi.ErrBadParameter
	}

	fmt.Println(service[1])
	return gopi.ErrNotImplemented

	// Success
	return nil
}

func AddService(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	// Obtain the executable name
	exec := reExecutable.FindStringSubmatch(args[0])
	if len(exec) != 2 {
		return gopi.ErrBadParameter
	}
	if len(args) < 2 {
		return gopi.ErrBadParameter
	}
	if args[1] != "add" {
		return gopi.ErrBadParameter
	}
	if service, err := gaffer.AddServiceForPath(exec[1], []string{}); err != nil {
		return err
	} else {
		return OutputServices(os.Stdout, []rpc.GafferService{service})
	}
}
