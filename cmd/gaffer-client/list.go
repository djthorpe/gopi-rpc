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
	rpc "github.com/djthorpe/gopi-rpc"
	tablewriter "github.com/olekukonko/tablewriter"
)

////////////////////////////////////////////////////////////////////////////////

func ListAllExecutables(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	if list, err := gaffer.ListExecutables(); err != nil {
		return err
	} else {
		output := tablewriter.NewWriter(os.Stdout)
		for _, cell := range list {
			output.Append([]string{
				"/" + cell,
			})
		}
		output.Render()
	}

	// Return success
	return nil
}

func ListAllGroups(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	if groups, err := gaffer.ListGroups(); err != nil {
		return err
	} else if len(groups) == 0 {

	} else {
		return OutputGroups(os.Stdout, groups)
	}

	// Return success
	return nil
}

func ListAllServiceRecords(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	if list, err := discovery.Enumerate(rpc.DISCOVERY_TYPE_DB, time.Second); err != nil {
		return err
	} else {
		output := tablewriter.NewWriter(os.Stdout)
		for _, cell := range list {
			output.Append([]string{
				cell,
			})
		}
		output.Render()
	}

	// Return success
	return nil
}

func ListAllInstances(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	if instances, err := gaffer.ListInstances(); err != nil {
		return err
	} else if len(instances) == 0 {
		return ListAllServices(args, gaffer, discovery)
	} else {
		return OutputInstances(os.Stdout, instances)
	}
}

func ListAllServices(args []string, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	if services, err := gaffer.ListServices(); err != nil {
		return err
	} else if len(services) == 0 {
		return fmt.Errorf("No services")
	} else {
		return OutputServices(os.Stdout, services)
	}
}
