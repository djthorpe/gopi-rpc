/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"fmt"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
	tablewriter "github.com/olekukonko/tablewriter"
)

////////////////////////////////////////////////////////////////////////////////
// PRINT SERVICE TABLE

func PrintServiceTable(_ gopi.App, cmd *Runnable, services []rpc.GafferService) {
	table := tablewriter.NewWriter(cmd.out)
	table.SetHeader([]string{"ID", "Name", "Path", "Cwd", "User", "Group", "Args", "Enabled"})
	for _, service := range services {
		table.Append([]string{
			fmt.Sprint(service.Sid()),
			fmt.Sprint(service.Name()),
			fmt.Sprint(service.Path()),
			fmt.Sprint(service.Cwd()),
			fmt.Sprint(service.User()),
			fmt.Sprint(service.Group()),
			fmt.Sprint(service.Args()),
			fmt.Sprint(service.Enabled()),
		})
	}
	table.Render()
}
