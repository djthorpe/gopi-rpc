/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type CommandFunc func(gopi.App, *Runnable) (bool, error)

type Command struct {
	name, description string
	function          CommandFunc
}

type Runnable struct {
	cmd  *Command
	stub rpc.GafferClientStub
	out  io.Writer
	args []string
}

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

const (
	RPC_TIMEOUT = 5 * time.Second
)

var (
	command = []*Command{
		&Command{"services", "List all services", ListServices},
		&Command{"enable", "Enable service by Name or Id", EnableService},
		&Command{"disable", "Disable service by Name or Id", DisableService},
		&Command{"rename", "Rename service by Name or Id", RenameService},
		&Command{"args", "Set service arguments", SetServiceArguments},
		&Command{"start", "Start a service instamce by Name or Id", StartService},
	}
)

////////////////////////////////////////////////////////////////////////////////
// NEW RUNNABLE

func NewRunnable(stub rpc.GafferClientStub, args []string) (*Runnable, error) {
	if len(args) == 0 {
		// Return default runnable
		return &Runnable{command[0], stub, os.Stdout, nil}, nil
	}

	// Find command
	for _, cmd := range command {
		if cmd.name == args[0] {
			// Return runnable
			return &Runnable{cmd, stub, os.Stdout, args[1:]}, nil
		}
	}

	// Return 'not found'
	return nil, fmt.Errorf("Unknown command %v", strconv.Quote(args[0]))
}

func (this *Runnable) Run(app gopi.App) (bool, error) {
	return this.cmd.function(app, this)
}

func (this *Runnable) LookupService(arg string) (rpc.GafferService, error) {
	// Create running context
	ctx, cancel := context.WithTimeout(context.Background(), RPC_TIMEOUT)
	defer cancel()

	// Call for services
	services := make([]rpc.GafferService, 0)
	if services_, err := this.stub.Services(ctx); err != nil {
		return nil, err
	} else if sid, err := strconv.ParseUint(arg, 10, 32); err == nil {
		// Lookup by ID
		for _, service := range services_ {
			if service.Sid() == uint32(sid) {
				services = append(services, service)
			}
		}
	} else {
		// Lookup by name
		for _, service := range services_ {
			if service.Name() == arg {
				services = append(services, service)
			}
		}
	}

	// Check for ambiguity
	if len(services) == 0 {
		return nil, gopi.ErrNotFound
	} else if len(services) > 1 {
		return nil, gopi.ErrDuplicateItem
	} else {
		return services[0], nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// LIST SERVICES

func ListServices(app gopi.App, cmd *Runnable) (bool, error) {
	// Check arguments
	if len(cmd.args) != 0 {
		return false, gopi.ErrBadParameter
	}

	// Create running context
	ctx, cancel := context.WithTimeout(context.Background(), RPC_TIMEOUT)
	defer cancel()

	// Call for services
	if services, err := cmd.stub.Services(ctx); err != nil {
		return false, err
	} else if len(services) == 0 {
		fmt.Fprint(cmd.out, "No services returned")
	} else {
		PrintServiceTable(app, cmd, services)
	}

	// Return success
	return false, nil
}

////////////////////////////////////////////////////////////////////////////////
// ENABLE AND DISABLE SERVICES

func EnableService(app gopi.App, cmd *Runnable) (bool, error) {
	// Check arguments
	if len(cmd.args) != 1 {
		return false, gopi.ErrBadParameter
	}

	// Create running context
	ctx, cancel := context.WithTimeout(context.Background(), RPC_TIMEOUT)
	defer cancel()

	// Return service
	if service, err := cmd.LookupService(cmd.args[0]); err != nil {
		return false, err
	} else if service, err := cmd.stub.Update(ctx, cmd.stub.Mutable(service).SetEnabled(true)); err != nil {
		return false, err
	} else {
		PrintServiceTable(app, cmd, []rpc.GafferService{service})
	}

	// Return success
	return false, nil
}

func DisableService(app gopi.App, cmd *Runnable) (bool, error) {
	// Check arguments
	if len(cmd.args) != 1 {
		return false, gopi.ErrBadParameter
	}

	// Create running context
	ctx, cancel := context.WithTimeout(context.Background(), RPC_TIMEOUT)
	defer cancel()

	// Return service
	if service, err := cmd.LookupService(cmd.args[0]); err != nil {
		return false, err
	} else if service, err := cmd.stub.Update(ctx, cmd.stub.Mutable(service).SetEnabled(false)); err != nil {
		return false, err
	} else {
		PrintServiceTable(app, cmd, []rpc.GafferService{service})
	}

	// Return success
	return false, nil
}

func RenameService(app gopi.App, cmd *Runnable) (bool, error) {
	// Check arguments
	if len(cmd.args) != 2 {
		return false, gopi.ErrBadParameter
	}

	// Create running context
	ctx, cancel := context.WithTimeout(context.Background(), RPC_TIMEOUT)
	defer cancel()

	// Return service
	if service, err := cmd.LookupService(cmd.args[0]); err != nil {
		return false, err
	} else if service, err := cmd.stub.Update(ctx, cmd.stub.Mutable(service).SetName(cmd.args[1])); err != nil {
		return false, err
	} else {
		PrintServiceTable(app, cmd, []rpc.GafferService{service})
	}

	// Return success
	return false, nil
}

func StartService(app gopi.App, cmd *Runnable) (bool, error) {
	// Check arguments
	if len(cmd.args) != 1 {
		return false, gopi.ErrBadParameter
	}

	// Create running context
	ctx, cancel := context.WithTimeout(context.Background(), RPC_TIMEOUT)
	defer cancel()

	// Return service
	if service, err := cmd.LookupService(cmd.args[0]); err != nil {
		return false, err
	} else if services, err := cmd.stub.Start(ctx, service.Sid()); err != nil {
		return false, err
	} else {
		PrintServiceTable(app, cmd, []rpc.GafferService{services})
	}

	// Return success
	return false, nil
}

func SetServiceArguments(app gopi.App, cmd *Runnable) (bool, error) {
	// Check arguments
	if len(cmd.args) < 1 {
		return false, gopi.ErrBadParameter
	}

	// Create running context
	ctx, cancel := context.WithTimeout(context.Background(), RPC_TIMEOUT)
	defer cancel()

	// Return service
	if service, err := cmd.LookupService(cmd.args[0]); err != nil {
		return false, err
	} else if services, err := cmd.stub.Update(ctx, cmd.stub.Mutable(service).SetArgs(cmd.args[1:])); err != nil {
		return false, err
	} else {
		PrintServiceTable(app, cmd, []rpc.GafferService{services})
	}

	// Return success
	return false, nil
}
