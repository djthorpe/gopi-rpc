/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"fmt"
	"io"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	tablewriter "github.com/olekukonko/tablewriter"
)

////////////////////////////////////////////////////////////////////////////////

func PrintEvent(fh io.Writer, event gopi.RPCEvent) {

	table := tablewriter.NewWriter(fh)
	service := event.ServiceRecord()
	t := strings.ToLower(strings.TrimPrefix(fmt.Sprint(event.Type()), "RPC_EVENT_"))

	if service == nil {
		table.Append([]string{
			t,
		})
		table.Render()
	} else {
		table.Append([]string{
			t,
			RenderService(service),
			service.Name(),
			RenderHost(service),
			RenderIP(service),
			RenderTxt(service),
		})
		table.Render()

	}
}

func PrintServices(fh io.Writer, services []gopi.RPCServiceRecord) {
	table := tablewriter.NewWriter(fh)
	table.SetHeader([]string{"Service", "Name", "Host", "Addr", "TXT"})
	for _, service := range services {
		table.Append([]string{
			RenderService(service),
			service.Name(),
			RenderHost(service),
			RenderIP(service),
			RenderTxt(service),
		})
	}
	table.Render()
}

func RenderHost(service gopi.RPCServiceRecord) string {
	if service.Port() == 0 {
		return service.Host()
	} else {
		return fmt.Sprintf("%v:%v", service.Host(), service.Port())
	}
}

func RenderService(service gopi.RPCServiceRecord) string {
	if service.Subtype() == "" {
		return service.Service()
	} else {
		return fmt.Sprintf("%v, %v", service.Subtype(), service.Service())
	}
}

func RenderIP(service gopi.RPCServiceRecord) string {
	ips := make([]string, 0)
	for _, ip := range service.IP4() {
		ips = append(ips, fmt.Sprint(ip))
	}
	for _, ip := range service.IP6() {
		ips = append(ips, fmt.Sprint(ip))
	}
	return strings.Join(ips, "\n")
}

func RenderTxt(service gopi.RPCServiceRecord) string {
	return strings.Join(service.Text(), "\n")
}
