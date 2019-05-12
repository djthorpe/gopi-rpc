/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc"
	tablewriter "github.com/olekukonko/tablewriter"
)

////////////////////////////////////////////////////////////////////////////////

func RenderGroupList(groups []string) string {
	groups_ := ""
	for i, group := range groups {
		if i > 0 {
			groups_ += ","
		}
		groups_ += "@" + group
	}
	return groups_
}

func RenderFlags(flags []string) string {
	flags_ := ""
	for i, flag := range flags {
		if i > 0 {
			flags_ += "\n"
		}
		flags_ += flag
	}
	return flags_
}

func RenderMode(service rpc.GafferService) string {
	if service.InstanceCount() == 0 {
		return "disabled"
	}
	if mode := fmt.Sprint(service.Mode()); strings.HasPrefix(mode, "GAFFER_MODE_") {
		return strings.ToLower(strings.TrimPrefix(mode, "GAFFER_MODE_"))
	} else {
		return mode
	}
}

func RenderDuration(duration time.Duration) string {
	if duration == 0 {
		return "-"
	}
	return fmt.Sprint(duration.Truncate(time.Second))
}

func RenderServices(fh io.Writer, services []rpc.GafferService) {
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
}

func RenderGroups(fh io.Writer, groups []rpc.GafferServiceGroup) {
	output := tablewriter.NewWriter(fh)
	output.SetHeader([]string{"GROUP", "FLAGS", "ENV"})
	for _, group := range groups {
		output.Append([]string{
			"@" + group.Name(),
			RenderFlags(group.Flags()),
			RenderFlags(group.Env()),
		})
	}
	output.Render()
}

func RenderInstances(fh io.Writer, instances []rpc.GafferServiceInstance) {
	output := tablewriter.NewWriter(fh)
	output.SetHeader([]string{"INSTANCE", "SERVICE", "FLAGS", "ENV"})
	for _, instance := range instances {
		output.Append([]string{
			fmt.Sprint(instance.Id()),
			fmt.Sprint(instance.Service().Name()),
			RenderFlags(instance.Flags()),
			RenderFlags(instance.Env()),
		})
	}
	output.Render()
}
