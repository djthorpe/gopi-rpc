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
	"os"
	"regexp"
	"strconv"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
	tablewriter "github.com/olekukonko/tablewriter"
)

////////////////////////////////////////////////////////////////////////////////

func (this *Runner) Run(stub rpc.DiscoveryClient) error {
	this.log.Info("Connected to %v", stub.Conn().Addr())

	// By default, enumerate the services
	this.RegisterCommand(&Command{
		re:    regexp.MustCompile("^_$"),
		scope: COMMAND_SCOPE_ROOT,
		cb:    this.EnumerateServices,
	})
	// Watch for discovery events
	this.RegisterCommand(&Command{
		re:    regexp.MustCompile("^watch$"),
		scope: COMMAND_SCOPE_ROOT,
		cb:    this.Watch,
	})
	// Lookup service
	this.RegisterCommand(&Command{
		re:    regexp.MustCompile("^_([A-Za-z]\\w+)\\._tcp$"),
		scope: COMMAND_SCOPE_ROOT,
		cb:    this.ServiceCommands,
	})
	this.RegisterCommand(&Command{
		re:    regexp.MustCompile("^([A-Za-z]\\w+)$"),
		scope: COMMAND_SCOPE_ROOT,
		cb:    this.ServiceCommands,
	})

	args := this.Args()
	if len(args) == 0 {
		// Default command is enumerate services
		return this.EnumerateServices(stub, nil)
	} else if command, matches := this.GetCommand(args[0], COMMAND_SCOPE_ROOT); command == nil {
		return gopi.ErrHelp
	} else {
		return command.cb(stub, matches[1:])
	}
}

func (this *Runner) Watch(stub rpc.DiscoveryClient, services []string) error {

	// Watch
	this.AddStub(stub, "")

	// Wait for CTRL+C signal, then stop
	this.log.Info("Press CTRL+C to end watching for events")
	this.app.WaitForSignal()

	// Return success
	return nil
}

func (this *Runner) EnumerateServices(stub rpc.DiscoveryClient, _ []string) error {
	timeout, _ := this.app.AppFlags.GetDuration("timeout")
	if services, err := stub.Enumerate(this.DiscoveryType(), this.TimeoutOrDefault(timeout)); err != nil {
		return err
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Service"})
		for _, service := range services {
			table.Append([]string{service})
		}
		table.Render()
	}

	// Success
	return nil
}

func (this *Runner) ServiceCommands(stub rpc.DiscoveryClient, service []string) error {
	if len(service) == 1 {
		return this.ServiceLookup(stub, service)
	} else {
		return gopi.ErrHelp
	}
}

func (this *Runner) ServiceLookup(stub rpc.DiscoveryClient, service []string) error {
	timeout, _ := this.app.AppFlags.GetDuration("timeout")
	if len(service) != 1 {
		return gopi.ErrBadParameter
	} else if services, err := stub.Lookup(fmt.Sprintf("_%v._tcp", service[0]), this.DiscoveryType(), this.TimeoutOrDefault(timeout)); err != nil {
		return err
	} else if len(services) == 0 {
		return fmt.Errorf("No service records found for %v", strconv.Quote(service[0]))
	} else {
		PrintServices(os.Stdout, services)
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (this *Runner) EventTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	var last gopi.RPCEvent

	start <- gopi.DONE
	events := this.Merger.Subscribe()
FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt_, ok := evt.(gopi.RPCEvent); ok && evt_ != nil && evt_.Type() != gopi.RPC_EVENT_NONE {
				if EventEquals(last, evt_) == false {
					PrintEvent(os.Stdout, evt_)
				}
				last = evt_
			}
		case err := <-this.errors:
			this.log.Error("%v", err)
		case <-stop:
			break FOR_LOOP
		}
	}

	// Unsubscribe, return success
	this.Merger.Unsubscribe(events)
	return nil
}

func EventEquals(this, other gopi.RPCEvent) bool {
	// Check event
	if this == nil && other == nil {
		return true
	} else if this == nil || other == nil {
		return false
	} else if this.Name() != other.Name() {
		return false
	} else if this.Type() != other.Type() {
		return false
	} else if this.ServiceRecord() == nil && other.ServiceRecord() == nil {
		return true
	} else if this.ServiceRecord() == nil || other.ServiceRecord() == nil {
		return false
	}

	// Check service record
	this_r := this.ServiceRecord()
	other_r := other.ServiceRecord()
	if this_r.Name() != other_r.Name() {
		return false
	} else if this_r.Host() != other_r.Host() {
		return false
	} else if this_r.Port() != other_r.Port() {
		return false
	} else if this_r.Service() != other_r.Service() {
		return false
	} else if this_r.Subtype() != other_r.Subtype() {
		return false
	}

	// Check text records
	this_txt := this_r.Text()
	other_text := other_r.Text()
	if len(this_txt) != len(other_text) {
		return false
	}
	for i, t := range this_txt {
		if t != other_text[i] {
			return false
		}
	}

	// Return true
	return true
}

/*
func Register(app *gopi.AppInstance, client rpc.DiscoveryClient, name, service, subtype string, port uint) error {
	util := app.ModuleInstance("rpc/util").(rpc.Util)
	record := util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB)

	if err := record.SetName(name); err != nil {
		return err
	} else if hostname, err := os.Hostname(); err != nil {
		return err
	} else if addrs, err := net.LookupHost(hostname); err != nil {
		return err
	} else if len(addrs) == 0 {
		return fmt.Errorf("No addresses found for host %v", strconv.Quote(hostname))
	} else if err := record.SetHostPort(hostname + ":" + fmt.Sprint(port)); err != nil {
		return err
	} else {
		for _, addr := range addrs {
			if ip := net.ParseIP(addr); ip == nil {
				return fmt.Errorf("Cannot parse %v", strconv.Quote(addr))
			} else if err := record.AppendIP(ip); err != nil {
				return err
			}
		}
	}
	if err := record.SetService(service, subtype); err != nil {
		return err
	}

	fmt.Println(record)

	return client.Register(record)
}
*/
