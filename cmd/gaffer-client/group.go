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

func GroupCommands(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	// Obtain the group name
	group := reGroup.FindStringSubmatch(args[0])
	if len(group) != 2 {
		return gopi.ErrBadParameter
	}

	// Parse arguments
	switch len(args) {
	case 1:
		// Return services for group
		if services, err := gaffer.ListServicesForGroup(group[1]); err != nil {
			return err
		} else if len(services) == 0 {
			return fmt.Errorf("No services")
		} else {
			return OutputServices(os.Stdout, services)
		}
	case 2:
		switch args[1] {
		case "add":
			if group_, err := gaffer.AddGroupForName(group[1]); err != nil {
				return err
			} else {
				return OutputGroups(os.Stdout, []rpc.GafferServiceGroup{group_})
			}
		case "rm":
			if err := gaffer.RemoveGroupForName(group[1]); err != nil {
				return err
			} else {
				return ListAllGroups(args, gaffer, discovery)
			}
		case "start":
			return gopi.ErrNotImplemented
		case "stop":
			return gopi.ErrNotImplemented
		default:
			return gopi.ErrHelp
		}
	default:
		return gopi.ErrHelp
	}

	// Success
	return nil
}
