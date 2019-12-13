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
	rpc "github.com/djthorpe/gopi-rpc"
)

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
	group, err := this.gaffer.AddGroupForName(args[0])
	if err != nil {
		return err
	}
	if len(args) >= 3 {
		return this.SetGroupFlags(cmd, args)
	} else {
		OutputGroups(os.Stdout, []rpc.GafferServiceGroup{
			group,
		})
	}

	// Return success
	return nil
}

func (this *Runner) RemoveGroup(cmd *Cmd, args []string) error {
	if len(args) != 2 {
		return this.SyntaxError(cmd)
	} else if err := this.gaffer.RemoveGroupForName(args[0]); err != nil {
		return err
	} else {
		return this.ListAllGroups(cmd, []string{})
	}

	// Return success
	return nil
}

func (this *Runner) SetGroupFlags(cmd *Cmd, args []string) error {
	if len(args) < 3 {
		return this.SyntaxError(cmd)
	} else if tuples, err := rpc.NewTuples(args[2:]); err != nil {
		return err
	} else if group, err := this.gaffer.SetFlagsForGroup(args[0], tuples); err != nil {
		return err
	} else {
		return OutputGroups(os.Stdout, []rpc.GafferServiceGroup{group})
	}

	// Success
	return nil
}

func (this *Runner) SetGroupEnv(cmd *Cmd, args []string) error {
	if len(args) < 3 {
		return this.SyntaxError(cmd)
	} else if tuples, err := rpc.NewTuples(args[2:]); err != nil {
		return err
	} else if group, err := this.gaffer.SetEnvForGroup(args[0], tuples); err != nil {
		return err
	} else {
		return OutputGroups(os.Stdout, []rpc.GafferServiceGroup{group})
	}

	// Success
	return nil
}
