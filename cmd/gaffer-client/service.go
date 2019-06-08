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
	tablewriter "github.com/olekukonko/tablewriter"
)

func (this *Runner) ListAllServices(cmd *Cmd, args []string) error {
	// No arguments allowed
	if len(args) > 0 {
		return this.SyntaxError(cmd)
	}
	// List services
	if services, err := this.gaffer.ListServices(); err != nil {
		return err
	} else if len(services) == 0 {
		return fmt.Errorf("No services")
	} else {
		return OutputServices(os.Stdout, services)
	}
}

func (this *Runner) ListAllGroups(cmd *Cmd, args []string) error {
	// No arguments allowed
	if len(args) > 0 {
		return this.SyntaxError(cmd)
	}
	// List groups
	if groups, err := this.gaffer.ListGroups(); err != nil {
		return err
	} else if len(groups) == 0 {
		return fmt.Errorf("No groups")
	} else {
		return OutputGroups(os.Stdout, groups)
	}
}

func (this *Runner) ListAllExecutables(cmd *Cmd, args []string) error {
	// No arguments allowed
	if len(args) > 0 {
		return this.SyntaxError(cmd)
	}
	// List executables
	if list, err := this.gaffer.ListExecutables(); err != nil {
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

func (this *Runner) ListAllServiceRecords(cmd *Cmd, args []string) error {
	// No arguments allowed
	if len(args) > 0 {
		return this.SyntaxError(cmd)
	}
	discovery_type := rpc.DISCOVERY_TYPE_DB
	if dns, _ := this.flags.GetBool("dns"); dns {
		discovery_type = rpc.DISCOVERY_TYPE_DNS
	}
	// List service records
	if list, err := this.discovery.Enumerate(discovery_type, DISCOVERY_TIMEOUT); err != nil {
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

	// Success
	return nil
}

func (this *Runner) LookupServiceRecords(cmd *Cmd, args []string) error {
	return gopi.ErrNotImplemented
}

func (this *Runner) AddService(cmd *Cmd, args []string) error {
	// First argument is the executable name, and second should be add
	if len(args) < 2 || args[1] != "add" {
		return this.SyntaxError(cmd)
	}
	// TODO: Remaining flags
	if len(args) != 2 {
		return this.SyntaxError(cmd)
	}
	if service, err := this.gaffer.AddServiceForPath(args[0], []string{}); err != nil {
		return err
	} else {
		return OutputServices(os.Stdout, []rpc.GafferService{service})
	}
	// Return success
	return nil
}

func (this *Runner) ServiceCommands(cmd *Cmd, args []string) error {
	// Check for arguments
	if len(args) < 1 {
		return runner.SyntaxError(cmd)
	}
	// With no arguments, list services
	if len(args) == 1 {
		// TODO: List instances for service
		if list, err := this.gaffer.ListInstances(); err != nil {
			return err
		} else {
			OutputInstances(os.Stdout, list)
		}
	} else if cmd_, _ := this.CommandForScope(SCOPE_SERVICE, args[1]); cmd_ == nil {
		return runner.SyntaxError(cmd)
	} else {
		return cmd_.f(cmd_, args)
	}

	// Return success
	return nil
}

func (this *Runner) GroupCommands(cmd *Cmd, args []string) error {
	// Check for arguments
	if len(args) < 2 {
		return runner.SyntaxError(cmd)
	}
	if cmd_, _ := this.CommandForScope(SCOPE_GROUP, args[1]); cmd_ == nil {
		return runner.SyntaxError(cmd)
	} else {
		return cmd_.f(cmd_, args)
	}

	// Return success
	return nil
}

func (this *Runner) AddGroup(cmd *Cmd, args []string) error {
	if group, err := this.gaffer.AddGroupForName(args[0]); err != nil {
		return err
	} else {
		OutputGroups(os.Stdout, []rpc.GafferServiceGroup{
			group,
		})
	}

	// Return success
	return nil
}

func (this *Runner) RemoveGroup(cmd *Cmd, args []string) error {
	if err := this.gaffer.RemoveGroupForName(args[0]); err != nil {
		return err
	} else {
		return this.ListAllGroups(cmd, []string{})
	}

	// Return success
	return nil
}

func (this *Runner) RemoveService(cmd *Cmd, args []string) error {
	if len(args) == 0 {
		return this.SyntaxError(cmd)
	} else if err := this.gaffer.RemoveServiceForName(args[0]); err != nil {
		return err
	} else {
		// List services
		return this.ListAllServices(cmd, []string{})
	}
}

func (this *Runner) StartService(cmd *Cmd, args []string) error {
	// Generate an id
	if id, err := this.gaffer.GetInstanceId(); err != nil {
		return err
	} else if instance, err := this.gaffer.StartInstance(args[0], id); err != nil {
		return err
	} else {
		OutputInstances(os.Stdout, []rpc.GafferServiceInstance{
			instance,
		})
	}

	// Success
	return nil
}
