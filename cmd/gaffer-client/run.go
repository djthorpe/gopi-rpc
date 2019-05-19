/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
	"regexp"

	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////

type Command struct {
	Name        string
	Description string
	Callback    func([]string, rpc.GafferClient, rpc.DiscoveryClient) error
}

////////////////////////////////////////////////////////////////////////////////

var (
	reGroup      = regexp.MustCompile("^@[A-Za-z][A-Za-z0-9\\.\\-_]*$")
	reExecutable = regexp.MustCompile("^/[A-Za-z][A-Za-z0-9\\/\\.\\-_]*$")
	reService    = regexp.MustCompile("^[A-Za-z][A-Za-z0-9\\.\\-_]*$")
	reInstance   = regexp.MustCompile("^[1-9][0-9]*$")
	reRecord     = regexp.MustCompile("^_[A-Za-z][A-Za-z0-9\\.\\-_]*$")
)

var (
	cmd = []*Command{
		// First command is the default one
		&Command{"services", "List services and instances", CommandServices},
		&Command{"groups", "List groups", CommandGroups},
		&Command{"records", "List service records", CommandRecords},
		&Command{"execs", "List executables", CommandExecs},
	}
)

////////////////////////////////////////////////////////////////////////////////

func GetCommandForName(name string) *Command {
	for _, command := range cmd {
		if command.Name == name {
			return command
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func Run(app *gopi.AppInstance, gaffer rpc.GafferClient, discovery rpc.DiscoveryClient) error {
	args := app.AppFlags.Args()
	if len(args) == 0 {
		RunCommand(cmd[0], app, gaffer, discovery)
	} else {
		return gopi.ErrNotImplemented
	}
}
