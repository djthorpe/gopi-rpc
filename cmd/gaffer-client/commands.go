/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"os"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	tablewriter "github.com/olekukonko/tablewriter"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi/sys/logger"

	// Services
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/gaffer"
)

////////////////////////////////////////////////////////////////////////////////

type Command struct {
	Name        string
	Description string
	Callback    func([]string, rpc.GafferClient) error
}

////////////////////////////////////////////////////////////////////////////////

var (
	commands = []*Command{
		&Command{"service", "List available services", ListServices},
		&Command{"executables", "List available executables", ListExecutables},
	}
)

////////////////////////////////////////////////////////////////////////////////

func GetCommand(args []string) (*Command, error) {
	if len(args) == 0 {
		return commands[0], nil
	} else {
		for _, command := range commands {
			if command.Name == args[0] {
				return command, nil
			}
		}
	}
	return nil, gopi.ErrNotFound
}

func ListServices(args []string, client rpc.GafferClient) error {
	if len(args) > 1 {
		return gopi.ErrBadParameter
	} else if list, err := client.ListServices(); err != nil {
		return err
	} else {
		output := tablewriter.NewWriter(os.Stdout)
		output.SetHeader([]string{"SERVICE"})
		for _, cell := range list {
			output.Append([]string{
				cell.Name(),
			})
		}
		output.Render()
	}
	return nil
}

func ListExecutables(args []string, client rpc.GafferClient) error {
	if len(args) > 1 {
		return gopi.ErrBadParameter
	} else if list, err := client.ListExecutables(); err != nil {
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
	return nil
}

func Run(app *gopi.AppInstance, client rpc.GafferClient) error {
	args := app.AppFlags.Args()
	if cmd, err := GetCommand(args); err != nil {
		return err
	} else if err := cmd.Callback(args, client); err != nil {
		return err
	}

	// Success
	return nil
}
