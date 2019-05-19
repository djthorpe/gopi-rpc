/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
	"fmt"
	"io"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	tablewriter "github.com/olekukonko/tablewriter"
)

////////////////////////////////////////////////////////////////////////////////

func OutputServices(fh io.Writer, services []rpc.GafferService) error {
	output := tablewriter.NewWriter(fh)
	output.SetHeader([]string{"SERVICE", "GROUPS", "FLAGS", "MODE", "RUN TIME", "IDLE TIME"})
	for _, service := range services {
		output.Append([]string{
			service.Name(),
			RenderGroupList(service.Groups()),
			RenderFlags(service.Flags()),
			RenderMode(service),
			RenderDuration(service.RunTime()),
			RenderDuration(service.IdleTime()),
		})
	}
	output.Render()
	return nil
}

func OutputGroups(fh io.Writer, groups []rpc.GafferServiceGroup) error {
	output := tablewriter.NewWriter(fh)
	output.SetHeader([]string{"GROUP", "FLAGS", "ENV"})
	for _, group := range groups {
		output.Append([]string{
			"@" + group.Name(),
			RenderFlags(group.Flags()),
			RenderEnv(group.Env()),
		})
	}
	output.Render()
	return nil

}

func OutputInstances(fh io.Writer, instances []rpc.GafferServiceInstance) error {
	output := tablewriter.NewWriter(fh)
	output.SetHeader([]string{"INSTANCE", "SERVICE", "FLAGS", "ENV", "STATUS"})
	for _, instance := range instances {
		output.Append([]string{
			fmt.Sprint(instance.Id()),
			fmt.Sprint(instance.Service().Name()),
			RenderFlags(instance.Flags()),
			RenderEnv(instance.Env()),
			RenderInstanceStatus(instance),
		})
	}
	output.Render()
	return nil
}

func OutputRecords(fh io.Writer, records []gopi.RPCServiceRecord) error {
	output := tablewriter.NewWriter(fh)
	output.SetHeader([]string{"SERVICE", "NAME", "HOST", "ADDR", "TXT"})
	for _, record := range records {
		output.Append([]string{
			RenderService(record),
			record.Name(),
			RenderHost(record),
			RenderIP(record),
			RenderTxt(record),
		})
	}
	output.Render()
	return nil
}
