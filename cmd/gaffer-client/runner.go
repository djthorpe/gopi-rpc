/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////

type Runner struct {
	gaffer    rpc.GafferClient
	discovery rpc.DiscoveryClient
	flags     *gopi.Flags
	commands  map[Scope][]*Cmd
	cancels   []context.CancelFunc
	wait      bool

	event.Merger
	event.Tasks
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
	SCOPE_INSTANCE
	SCOPE_GROUP
	SCOPE_RECORD
	SCOPE_SERVICE_PARAM
)

var (
	reGroupKey = regexp.MustCompile("^@([A-Za-z][A-Za-z0-9\\-\\_]*)$")
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
				regexp.MustCompile("^([1-9][0-9]*)$"),
				"<instance> (stop)",
				"List or remove a service instance",
				this.InstanceCommands,
			},
			&Cmd{
				regexp.MustCompile("^_([a-zA-Z][a-zA-Z0-9\\-]*)$"),
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
				"<service> (start|rm|flags|set|disable|auto|manual)",
				"List service instances, remove or start a service",
				this.ServiceCommands,
			},
			&Cmd{
				regexp.MustCompile("^@([a-zA-Z][a-zA-Z0-9\\-\\_\\.]*)$"),
				"(<flag>...) <@group> (add|rm|flags|env)",
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
			&Cmd{
				regexp.MustCompile("^flags$"),
				"<service> flags (<key>|<key>=<value>...)",
				"Set service flags",
				this.SetServiceFlags,
			},
			&Cmd{
				regexp.MustCompile("^set$"),
				"<service> set (name=<value>|groups=@<group>,@<group>,...)",
				"Set service parameters",
				this.SetServiceParams,
			},
			&Cmd{
				regexp.MustCompile("^disable$"),
				"<service> disable",
				"Disable",
				this.DisableService,
			},
			&Cmd{
				regexp.MustCompile("^(manual|auto)$"),
				"<service> (manual|auto) (instance_count=<uint>|run=<duration>|idle=<duration>)",
				"Enable a service for manual or auto startup",
				this.EnableService,
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
			&Cmd{
				regexp.MustCompile("^flags$"),
				"<service> flags (<key>|<key>=<value>...)",
				"Set group flags",
				this.SetGroupFlags,
			},
			&Cmd{
				regexp.MustCompile("^env$"),
				"<service> env (<key>|<key>=<value>...)",
				"Set group environment",
				this.SetGroupEnv,
			},
		},
		SCOPE_INSTANCE: []*Cmd{
			&Cmd{
				regexp.MustCompile("^stop$"),
				"<instance> stop",
				"Stop an instance",
				this.StopInstance,
			},
		},
		SCOPE_RECORD: []*Cmd{},
		SCOPE_SERVICE_PARAM: []*Cmd{
			&Cmd{
				regexp.MustCompile("^name=([a-zA-Z][a-zA-Z0-9\\-\\_\\.]*)$"),
				"<service> set name=<service>",
				"Rename service",
				this.SetServiceName,
			},
			&Cmd{
				regexp.MustCompile("^groups=(@[a-zA-Z][a-zA-Z0-9\\-\\_\\.\\@\\,]*)$"),
				"<service> set groups=@<group>,@<group>,...",
				"Set service groups",
				this.SetServiceGroups,
			},
		},
	}

	this.cancels = make([]context.CancelFunc, 0)
	this.Tasks.Start(this.EventTask)

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

func (this *Runner) Close() error {
	// Cancel streams
	for _, cancel := range this.cancels {
		cancel()
	}

	// End tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Close merger
	this.Merger.Close()

	// Success
	return nil
}

func (this *Runner) Wait() bool {
	return this.wait
}

func (this *Runner) ParseGroups(groups string) ([]string, error) {
	if groups_ := strings.Split(groups, ","); len(groups_) == 0 {
		return nil, gopi.ErrBadParameter
	} else {
		g := make([]string, 0, len(groups_))
		for _, group := range groups_ {
			if matched := reGroupKey.FindStringSubmatch(group); len(matched) != 2 {
				return nil, fmt.Errorf("Invalid group: %v", strconv.Quote(group))
			} else {
				g = append(g, matched[1])
			}
		}
		return g, nil
	}
}

func (this *Runner) AddGaffer(gaffer rpc.GafferClient) {
	ctx, cancel := context.WithCancel(context.Background())
	this.cancels = append(this.cancels, cancel)

	// Add gaffer to the merger
	this.Merger.Merge(gaffer)

	go func() {
		if err := gaffer.StreamEvents(ctx); err != nil && err != context.Canceled {
			fmt.Println(err)
		}
	}()

}

func (this *Runner) EventTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	events := this.Merger.Subscribe()
FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt_, ok := evt.(rpc.GafferEvent); ok {
				this.PrintEvent(evt_)
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	// Unsubscribe, return success
	this.Merger.Unsubscribe(events)
	return nil
}

func (this *Runner) PrintEvent(evt rpc.GafferEvent) {
	switch evt.Type() {
	case rpc.GAFFER_EVENT_NONE:
		// Ping event
		return
	case rpc.GAFFER_EVENT_INSTANCE_RUN:
		// Ignore run messages
		return
	case rpc.GAFFER_EVENT_LOG_STDOUT:
		fmt.Print(string(evt.Data()))
	case rpc.GAFFER_EVENT_LOG_STDERR:
		fmt.Print(string(evt.Data()))
	case rpc.GAFFER_EVENT_INSTANCE_STOP_OK:
		fmt.Printf("%v[%v] Stopped OK\n", evt.Service().Name(), evt.Instance().Id())
	case rpc.GAFFER_EVENT_INSTANCE_STOP_ERROR:
		fmt.Printf("%v[%v] Stopped with exit code %v\n", evt.Service().Name(), evt.Instance().Id(), evt.Instance().ExitCode())
	case rpc.GAFFER_EVENT_INSTANCE_STOP_KILLED:
		fmt.Printf("%v[%v] Killed\n", evt.Service().Name(), evt.Instance().Id())
		/*
			case GAFFER_EVENT_SERVICE_ADD:
			case GAFFER_EVENT_SERVICE_CHANGE:
			case GAFFER_EVENT_SERVICE_REMOVE:
			case GAFFER_EVENT_GROUP_ADD:
			case GAFFER_EVENT_GROUP_CHANGE:
			case GAFFER_EVENT_GROUP_REMOVE:
			case GAFFER_EVENT_INSTANCE_ADD:
			case GAFFER_EVENT_INSTANCE_START:
		*/
	default:
		fmt.Println("Unhandled:", evt.Type())
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
