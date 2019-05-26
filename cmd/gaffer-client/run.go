/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
	"fmt"
	"regexp"

	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////

type Command struct {
	Name        string
	Re          *regexp.Regexp
	Description string
	Callback    func([]string, rpc.GafferClient, rpc.DiscoveryClient) error
}

////////////////////////////////////////////////////////////////////////////////

var (
	reGroup            = regexp.MustCompile("^@([A-Za-z][A-Za-z0-9\\.\\-_]*)$")
	reExecutable       = regexp.MustCompile("^/([A-Za-z][A-Za-z0-9\\/\\.\\-_]*)$")
	reService          = regexp.MustCompile("^([A-Za-z][A-Za-z0-9\\.\\-_]*)$")
	reServiceRemove    = regexp.MustCompile("^([A-Za-z][A-Za-z0-9\\.\\-_]*) rm$")
	reServiceStartStop = regexp.MustCompile("^([A-Za-z][A-Za-z0-9\\.\\-_]*) (start|stop)$")
	reServiceFlags     = regexp.MustCompile("^([A-Za-z][A-Za-z0-9\\.\\-_]*) flags$")
	reInstance         = regexp.MustCompile("^[1-9][0-9]*$")
	reRecord           = regexp.MustCompile("^_[A-Za-z][A-Za-z0-9\\.\\-_]*$")
)

var (
	root_commands = []*Command{
		// First command is the default one
		&Command{"", nil, "List all services & instances", ListAllInstances},
		&Command{"/", nil, "List all executables", ListAllExecutables},
		&Command{"@", nil, "List all groups", ListAllGroups},
		&Command{"_", nil, "List all service records", ListAllServiceRecords},
		&Command{"_<service-type>._tcp", reRecord, "List service records", RecordCommands},
		&Command{"/<executable> add name=<service> groups=@<group-list> mode=(manual|auto)", reExecutable, "Add service", AddService},
		&Command{"<service> rm", reServiceRemove, "Remove Service", ServiceCommands},
		&Command{"<service> (start|stop)", reServiceStartStop, "Start or stop service instances", ServiceCommands},
		&Command{"<service> flags (<key>=<value> | <key>)...", reServiceFlags, "Set service flags", ServiceCommands},
		&Command{"<service> set name=<service> groups=@<group-list>", reServiceFlags, "Set service parameters", ServiceCommands},
		&Command{"<service> disable", reServiceFlags, "Disable service", ServiceCommands},
		&Command{"<service> (manual|auto) instance_count=<uint> run_time=<duration> idle_time=<duration>", reServiceFlags, "Enable service", ServiceCommands},
		&Command{"@<group> add", reGroup, "Add a group", GroupCommands},
		&Command{"@<group> rm", reGroup, "Remove a group", GroupCommands},
		&Command{"@<group> flags (<key>=<value> | <key>)...", reGroup, "Set group flags", GroupCommands},
		&Command{"@<group> env (<key>=<value> | <key>)...", reGroup, "Set group environment", GroupCommands},
		&Command{"@<group> set name=@<group>", reGroup, "Set group parameters", GroupCommands},
	}
)

////////////////////////////////////////////////////////////////////////////////

func GetCommandForArgument(commands []*Command, arg string) *Command {
	if arg == "" {
		return commands[0]
	} else {
		for _, command := range commands {
			if command.Re != nil {
				if command.Re.MatchString(arg) {
					return command
				}
			} else if command.Name == arg {
				return command
			}
		}
	}
	return nil
}

func Run(app *gopi.AppInstance, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	// Get command
	args := app.AppFlags.Args()
	command := GetCommandForArgument(root_commands, "")
	if len(args) > 0 {
		command = GetCommandForArgument(root_commands, args[0])
	}
	if command == nil {
		return gopi.ErrHelp
	}

	// Call command
	if err := command.Callback(args, gaffer, discovery); err == gopi.ErrBadParameter {
		return fmt.Errorf("Usage: %v %v", app.AppFlags.Name(), command.Name)
	} else {
		return err
	}
}
