/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
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
	"strings"
	"time"

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
		&Command{"group", "List available groups", ListGroups},
		&Command{"instance", "List running instances", ListInstances},
		&Command{"executables", "List available executables", ListExecutables},
		&Command{"add", "Add new service or group", AddServiceGroup},
		&Command{"rm", "Remove a service or group", RemoveServiceGroup},
		&Command{"start", "Start a service or group", StartServiceGroup},
		&Command{"stop", "Stop an instance, service or group", StopServiceGroup},
		&Command{"flags", "Set flags for a service or group", SetFlags},
	}

	reInstanceId  = regexp.MustCompile("^[0-9]+$")
	reServiceName = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9\\.\\-\\_]*$")
	reGroupName   = regexp.MustCompile("^\\@[a-zA-Z][a-zA-Z0-9\\.\\-\\_]*$")
	reTupleKey    = reServiceName
	reTuplePair   = regexp.MustCompile("^([a-zA-Z][a-zA-Z0-9\\.\\-\\_]*)=(.*)$")
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
		RenderServices(os.Stdout, list)
	}
	return nil
}

func ListGroups(args []string, client rpc.GafferClient) error {
	if len(args) > 1 {
		return gopi.ErrBadParameter
	} else if list, err := client.ListGroups(); err != nil {
		return err
	} else {
		RenderGroups(os.Stdout, list)
	}
	return nil
}

func ListInstances(args []string, client rpc.GafferClient) error {
	if len(args) > 1 {
		return gopi.ErrBadParameter
	} else if list, err := client.ListInstances(); err != nil {
		return err
	} else {
		RenderInstances(os.Stdout, list)
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

func AddServiceGroup(args []string, client rpc.GafferClient) error {
	if len(args) != 2 {
		return gopi.ErrBadParameter
	}
	if service_group := args[1]; reGroupName.MatchString(service_group) {
		group := strings.TrimPrefix(service_group, "@")
		if group_, err := client.AddGroupForName(group); err != nil {
			return err
		} else {
			RenderGroups(os.Stdout, []rpc.GafferServiceGroup{group_})
		}
	} else {
		if service_, err := client.AddServiceForPath(service_group); err != nil {
			return err
		} else {
			RenderServices(os.Stdout, []rpc.GafferService{service_})
		}
	}

	// Success
	return nil
}

func RemoveServiceGroup(args []string, client rpc.GafferClient) error {
	if len(args) != 2 {
		return gopi.ErrBadParameter
	}
	if service_group := args[1]; reGroupName.MatchString(service_group) {
		group := strings.TrimPrefix(service_group, "@")
		if err := client.RemoveGroupForName(group); err != nil {
			return err
		} else {
			return ListGroups([]string{"group"}, client)
		}
	} else if reServiceName.MatchString(service_group) {
		if err := client.RemoveServiceForName(service_group); err != nil {
			return err
		} else {
			return ListServices([]string{"service"}, client)
		}
	} else {
		return gopi.ErrBadParameter
	}
}

func StartServiceGroup(args []string, client rpc.GafferClient) error {
	if len(args) != 2 {
		return gopi.ErrBadParameter
	}
	if service_group := args[1]; reGroupName.MatchString(service_group) {
		// We don't yet allow for starting service groups
		return gopi.ErrNotImplemented
	} else if reServiceName.MatchString(service_group) {
		if id, err := client.GetInstanceId(); err != nil {
			return err
		} else if instance, err := client.StartInstance(service_group, id); err != nil {
			return err
		} else {
			RenderInstances(os.Stdout, []rpc.GafferServiceInstance{instance})
		}
	} else {
		return gopi.ErrBadParameter
	}

	// Success
	return nil
}

func StopServiceGroup(args []string, client rpc.GafferClient) error {
	if len(args) != 2 {
		return gopi.ErrBadParameter
	}
	if service_group := args[1]; reInstanceId.MatchString(service_group) {
		if instance, err := strconv.ParseUint(service_group, 10, 32); err != nil {
			return err
		} else if instance, err := client.StopInstance(uint32(instance)); err != nil {
			return err
		} else {
			RenderInstances(os.Stdout, []rpc.GafferServiceInstance{instance})
		}
	} else {
		// We don't yet allow for stopping by service or group
		return gopi.ErrNotImplemented
	}

	// Success
	return nil
}

func SetFlags(args []string, client rpc.GafferClient) error {
	if len(args) <= 2 {
		return gopi.ErrBadParameter
	}
	if service_group := args[1]; reServiceName.MatchString(service_group) {
		if tuples, err := Tuples(client, args[2:], true); err != nil {
			return err
		} else if service, err := client.SetFlagsForService(service_group, tuples); err != nil {
			return err
		} else {
			RenderServices(os.Stdout, []rpc.GafferService{service})
		}
	} else if reGroupName.MatchString(service_group) {
		group := strings.TrimPrefix(service_group, "@")
		if tuples, err := Tuples(client, args[2:], false); err != nil {
			return err
		} else if group, err := client.SetFlagsForGroup(group, tuples); err != nil {
			return err
		} else {
			RenderGroups(os.Stdout, []rpc.GafferServiceGroup{group})
		}
	} else {
		return gopi.ErrBadParameter
	}

	// Success
	return nil
}

func Tuples(client rpc.GafferClient, args []string, flag bool) (rpc.GafferTuples, error) {
	if len(args) == 0 {
		// With zero arguments, return bad parameter
		return nil, gopi.ErrBadParameter
	} else if len(args) == 1 && args[0] == "-" {
		// Return empty tuples array
		return client.NewTuples(), nil
	} else {
		tuples := client.NewTuples()
		for _, tuple := range args {
			if flag && strings.HasPrefix(tuple, "-") {
				// Remove initial -
				tuple = strings.TrimPrefix(tuple, "-")
			}
			if reTupleKey.MatchString(tuple) {
				tuples.AddString(tuple, "true")
			} else if match := reTuplePair.FindStringSubmatch(tuple); match != nil && len(match) == 3 {
				if value, err := TupleConv(match[2]); err != nil {
					return nil, err
				} else {
					tuples.AddString(match[1], value)
				}
			} else {
				return nil, fmt.Errorf("Invalid argument: %v", strconv.Quote(tuple))
			}
		}
		return tuples, nil
	}
}

func TupleConv(value string) (string, error) {
	fmt.Println(value)

	if value_int, err := strconv.ParseInt(value, 10, 64); err == nil {
		return fmt.Sprint(value_int), nil
	}
	if value_uint, err := strconv.ParseUint(value, 10, 64); err == nil {
		return fmt.Sprint(value_uint), nil
	}
	if value_bool, err := strconv.ParseBool(value); err == nil {
		return fmt.Sprint(value_bool), nil
	}
	if value_duration, err := time.ParseDuration(value); err == nil {
		return fmt.Sprint(value_duration), nil
	}
	if value_string, err := strconv.Unquote(value); err == nil {
		return value_string, nil
	}
	return "", fmt.Errorf("Syntax error: %v", strconv.Quote(value))
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
