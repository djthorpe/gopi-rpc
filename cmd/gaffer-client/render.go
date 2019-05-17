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

func RenderFlags(flags rpc.Tuples) string {
	flags_ := flags.Flags()
	if len(flags_) == 0 {
		return "-"
	} else {
		flags__ := ""
		for i, flag := range flags_ {
			if i > 0 {
				flags__ += "\n"
			}
			flags__ += flag
		}
		return flags__
	}
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

func RenderInstanceStatus(instance rpc.GafferServiceInstance) string {
	if instance.Start().IsZero() && instance.Stop().IsZero() {
		return "Starting"
	} else if instance.Stop().IsZero() == false {
		// Stopped
		return fmt.Sprintf("Exit code %v", instance.ExitCode())
	} else if instance.Start().IsZero() == false {
		dur := time.Now().Sub(instance.Start()).Truncate(time.Minute)
		return fmt.Sprintf("Running %dm", uint(dur.Minutes()))
	}

	// Unhandled status
	return "??"
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
			"TODO",
			"TODO",
		})
	}
	output.Render()
}

func RenderInstances(fh io.Writer, instances []rpc.GafferServiceInstance) {
	output := tablewriter.NewWriter(fh)
	output.SetHeader([]string{"INSTANCE", "SERVICE", "FLAGS", "ENV", "STATUS"})
	for _, instance := range instances {
		output.Append([]string{
			fmt.Sprint(instance.Id()),
			fmt.Sprint(instance.Service().Name()),
			"TODO",
			"TODO",
			RenderInstanceStatus(instance),
		})
	}
	output.Render()
}
