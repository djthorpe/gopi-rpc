/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

/*
import (
	"encoding/json"
	"os"
	"regexp"
	"strconv"
	"strings"

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
		&Command{"flags", "Set flags for instance or group", FlagsServiceGroup},
		&Command{"groups", "Set groups for a service", SetServiceGroups},
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
		if service_, err := client.AddServiceForPath(service_group, []string{}); err != nil {
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

func FlagsServiceGroup(args []string, client rpc.GafferClient) error {
	if len(args) < 2 {
		return gopi.ErrBadParameter
	}
	if service_group := args[1]; reGroupName.MatchString(service_group) {
		group := strings.TrimPrefix(service_group, "@")
		if tuples, err := DecodeTuples(args[2:]); err != nil {
			return err
		} else if group, err := client.SetFlagsForGroup(group, tuples); err != nil {
			return err
		} else {
			RenderGroups(os.Stdout, []rpc.GafferServiceGroup{group})
		}
	} else if reServiceName.MatchString(service_group) {
		if tuples, err := DecodeTuples(args[2:]); err != nil {
			return err
		} else if service, err := client.SetFlagsForService(service_group, tuples); err != nil {
			return err
		} else {
			RenderServices(os.Stdout, []rpc.GafferService{service})
		}
	} else {
		return gopi.ErrBadParameter
	}

	// Return success
	return nil
}

func SetServiceGroups(args []string, client rpc.GafferClient) error {
	if len(args) < 3 {
		return gopi.ErrBadParameter
	}
	if service_group := args[1]; reServiceName.MatchString(service_group) {
		groups := make([]string, len(args)-2)
		for i, group := range args[2:] {
			if reGroupName.MatchString(group) == false {
				return gopi.ErrBadParameter
			} else {
				groups[i] = strings.TrimPrefix(group, "@")
			}
		}
		if service, err := client.SetServiceGroups(service_group, groups); err != nil {
			return err
		} else {
			RenderServices(os.Stdout, []rpc.GafferService{service})
		}
	} else {
		return gopi.ErrBadParameter
	}
	// Return success
	return nil
}

func DecodeTuples(args []string) (rpc.Tuples, error) {
	var tuples rpc.Tuples
	// Remove initial minus sign
	for i, arg := range args {
		args[i] = strings.TrimPrefix(arg, "-")
	}
	// JSON encode and then decode
	if b, err := json.Marshal(args); err != nil {
		return tuples, err
	} else if err := tuples.UnmarshalJSON(b); err != nil {
		return tuples, err
	}
	// Return tuples
	return tuples, nil
}

func Run2(app *gopi.AppInstance, client rpc.GafferClient) error {
	args := app.AppFlags.Args()
	if cmd, err := GetCommand(args); err != nil {
		return err
	} else if err := cmd.Callback(args, client); err != nil {
		return err
	}

	// Success
	return nil
}
*/
