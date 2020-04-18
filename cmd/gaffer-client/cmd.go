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
}

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

const (
	RPC_TIMEOUT = 5 * time.Second
)

var (
	command = []*Command{
		&Command{"services", "List all services", ListServices},
	}
)

////////////////////////////////////////////////////////////////////////////////
// NEW RUNNABLE

func NewRunnable(stub rpc.GafferClientStub, args []string) (*Runnable, error) {
	if len(args) == 0 {
		// Return default runnable
		return &Runnable{command[0], stub}, nil
	}

	// Find command
	for _, cmd := range command {
		if cmd.name == args[0] {
			// Return default runnable
			return &Runnable{cmd, stub}, nil
		}
	}

	// Return 'not found'
	return nil, fmt.Errorf("Unknown command %v", strconv.Quote(args[0]))
}

func (this *Runnable) Run(app gopi.App) (bool, error) {
	return this.cmd.function(app, this)
}

////////////////////////////////////////////////////////////////////////////////
// LIST SERVICES

func ListServices(_ gopi.App, cmd *Runnable) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RPC_TIMEOUT)
	defer cancel()
	if services, err := cmd.stub.Services(ctx); err != nil {
		return false, err
	} else {
		fmt.Println("services=", services)
	}

	// Return success
	return false, nil
}
