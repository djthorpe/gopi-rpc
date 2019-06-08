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
	"regexp"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////

type Runner struct {
	gaffer    rpc.GafferClient
	discovery rpc.DiscoveryClient
	flags     *gopi.Flags
	commands  map[Scope][]*Cmd
}

type Cmd struct {
	re                  *regexp.Regexp
	syntax, description string
	f                   Func
}

type Scope uint

type Func func(cmd *Cmd, args []string) error

////////////////////////////////////////////////////////////////////////////////

const (
	SCOPE_ROOT Scope = iota
	SCOPE_SERVICE
	SCOPE_GROUP
	SCOPE_RECORD
)

////////////////////////////////////////////////////////////////////////////////

func NewRunner() *Runner {
	this := new(Runner)
	this.commands = map[Scope][]*Cmd{
		SCOPE_ROOT: []*Cmd{
			&Cmd{
				regexp.MustCompile("^$"),
				"",
				"List all services",
				this.ListAllServices,
			},
			&Cmd{
				regexp.MustCompile("^/$"),
				"/",
				"List all executables",
				this.ListAllExecutables,
			},
			&Cmd{
				regexp.MustCompile("^@$"),
				"@",
				"List all groups",
				this.ListAllGroups,
			},
			&Cmd{
				regexp.MustCompile("^_$"),
				"_",
				"Enumerate service records",
				this.ListAllServiceRecords,
			},
			&Cmd{
				regexp.MustCompile("^_([a-zA-Z][a-zA-Z0-9\\-\\_\\.]*)$"),
				"_<type>",
				"Lookup service records",
				this.LookupServiceRecords,
			},
			&Cmd{
				regexp.MustCompile("^/([a-zA-Z][a-zA-Z0-9\\-\\_\\.]*)$"),
				"/<exec> add",
				"Add service",
				this.AddService,
			},
			&Cmd{
				regexp.MustCompile("^([a-zA-Z][a-zA-Z0-9\\-\\_\\.]*)$"),
				"<service> (start|rm)",
				"List service instances, remove or start a service",
				this.ServiceCommands,
			},
			&Cmd{
				regexp.MustCompile("^@([a-zA-Z][a-zA-Z0-9\\-\\_\\.]*)$"),
				"(<flag>...) <@group> (add|rm)",
				"Add or remove a group",
				this.GroupCommands,
			},
		},
		SCOPE_SERVICE: []*Cmd{
			&Cmd{
				regexp.MustCompile("^rm$"),
				"<service> rm",
				"Remove a service",
				this.RemoveService,
			},
			&Cmd{
				regexp.MustCompile("^start$"),
				"<service> start",
				"Start a service",
				this.StartService,
			},
		},
		SCOPE_GROUP: []*Cmd{
			&Cmd{
				regexp.MustCompile("^add$"),
				"<@group> add",
				"Add a group",
				this.AddGroup,
			},
			&Cmd{
				regexp.MustCompile("^rm$"),
				"<@group> rm",
				"Remove a group",
				this.RemoveGroup,
			},
		},
		SCOPE_RECORD: []*Cmd{},
	}
	return this
}

////////////////////////////////////////////////////////////////////////////////

func (this *Runner) Usage(flags *gopi.Flags) {
	fh := os.Stdout

	fmt.Fprintf(fh, "%v: Microservice Scheduler\nhttps://github.com/djthorpe/gopi-rpc/\n\n", flags.Name())
	fmt.Fprintf(fh, "Syntax:\n\n")
	fmt.Fprintf(fh, "  %v (<flag>...) <service|_type|@group|instance> (<command>) (<argument>...)\n\n", flags.Name())
	fmt.Fprintf(fh, "Commands:\n\n")

	// Root commands
	for _, command := range this.commands[SCOPE_ROOT] {
		fmt.Fprintf(fh, "  %v %v\n", flags.Name(), command.syntax)
		fmt.Fprintf(fh, "        %v\n", command.description)
		fmt.Fprintf(fh, "\n")
	}

	// Flags
	fmt.Fprintf(fh, "Flags:\n\n")
	flags.PrintDefaults()
}

func (this *Runner) SyntaxError(cmd *Cmd) error {
	return fmt.Errorf("Syntax: %v %v", this.flags.Name(), cmd.syntax)
}

func (this *Runner) CommandForScope(scope Scope, match string) (*Cmd, []string) {
	if cmds, exists := this.commands[scope]; exists {
		for _, cmd := range cmds {
			if params := cmd.re.FindStringSubmatch(match); len(params) > 0 {
				return cmd, params[1:]
			}
		}
	}
	// Failure
	return nil, nil
}

func (this *Runner) Run(gaffer rpc.GafferClient, discovery rpc.DiscoveryClient, flags *gopi.Flags) error {
	// Set running instance variables
	this.flags = flags
	this.gaffer = gaffer
	this.discovery = discovery

	args := this.flags.Args()
	cmd, params := this.CommandForScope(SCOPE_ROOT, "")
	if len(args) >= 1 {
		cmd, params = this.CommandForScope(SCOPE_ROOT, args[0])
		params = append(params, args[1:]...)
	}
	if cmd == nil {
		return gopi.ErrHelp
	} else {
		return cmd.f(cmd, params)
	}
}

/*

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

*/
