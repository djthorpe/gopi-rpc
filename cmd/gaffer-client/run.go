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
	Name        *regexp.Regexp
	Description string
	Callback    func([]string, rpc.GafferClient, rpc.DiscoveryClient) error
}

////////////////////////////////////////////////////////////////////////////////

var (
	reListGroups  = regexp.MustCompile("^@$")
	reListRecords = regexp.MustCompile("^_$")
	reListExecs   = regexp.MustCompile("^/$")
	reGroup       = regexp.MustCompile("^@([A-Za-z][A-Za-z0-9\\.\\-_]*)$")
	reExecutable  = regexp.MustCompile("^/([A-Za-z][A-Za-z0-9\\/\\.\\-_]*)$")
	reService     = regexp.MustCompile("^([A-Za-z][A-Za-z0-9\\.\\-_]*)$")
	reInstance    = regexp.MustCompile("^[1-9][0-9]*$")
	reRecord      = regexp.MustCompile("^_[A-Za-z][A-Za-z0-9\\.\\-_]*$")
)

var (
	cmd = []*Command{
		// First command is the default one
		&Command{nil, "List all instances", ListAllInstances},
		&Command{reListExecs, "List all executables", ListAllExecutables},
		&Command{reListGroups, "List all groups", ListAllGroups},
		&Command{reListRecords, "List all service records", ListAllServiceRecords},
		&Command{reGroup, "@<group> (add|rm|set|env|flags)", GroupCommands},
		&Command{reRecord, "_<service-type>._tcp", RecordCommands},
		&Command{reService, "<service> (rm|set|flags) (groups=@<group-list>)", ServiceCommands},
		&Command{reExecutable, "/<executable> add (groups=@<group-list>)", AddService},
	}
)

////////////////////////////////////////////////////////////////////////////////

func GetCommandForName(name string) *Command {
	if name == "" {
		return cmd[0]
	} else {
		for _, command := range cmd {
			if command.Name != nil && command.Name.MatchString(name) {
				return command
			}
		}
	}
	return nil
}

func Run(app *gopi.AppInstance, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {

	args := app.AppFlags.Args()
	if len(args) == 0 {
		if err := cmd[0].Callback(args, gaffer, discovery); err == gopi.ErrBadParameter {
			return fmt.Errorf("Usage: %v %v", app.AppFlags.Name(), cmd[0].Description)
		} else {
			return err
		}
	} else {
		if command := GetCommandForName(args[0]); command == nil {
			return gopi.ErrHelp
		} else if err := command.Callback(args, gaffer, discovery); err == gopi.ErrBadParameter {
			return fmt.Errorf("Usage: %v %v", app.AppFlags.Name(), command.Description)
		} else {
			return err
		}
	}
}
