/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2018
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package main

import (
	"context"
	"os"
	"regexp"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"

	// Modules
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/discovery"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/googlecast"
	_ "github.com/djthorpe/gopi-rpc/rpc/grpc/version"
	_ "github.com/djthorpe/gopi-rpc/sys/dns-sd"
	_ "github.com/djthorpe/gopi-rpc/sys/grpc"
	_ "github.com/djthorpe/gopi-rpc/sys/rpcutil"
	_ "github.com/djthorpe/gopi/sys/logger"
)

////////////////////////////////////////////////////////////////////////////////

type Runner struct {
	app      *gopi.AppInstance
	log      gopi.Logger
	services map[string]gopi.RPCServiceRecord
	commands []*Command
	stubs    []rpc.DiscoveryClient
	cancels  []context.CancelFunc
	errors   chan error

	event.Tasks
	event.Merger
	sync.WaitGroup
}

type Command struct {
	re    *regexp.Regexp
	scope CommandScope
	cb    func(stub rpc.DiscoveryClient, args []string) error
}

type CommandScope uint

////////////////////////////////////////////////////////////////////////////////

const (
	DISCOVERY_TIMEOUT = 400 * time.Millisecond
)

const (
	COMMAND_SCOPE_ROOT CommandScope = iota
)

////////////////////////////////////////////////////////////////////////////////

func NewRunner(app *gopi.AppInstance) *Runner {
	this := new(Runner)
	this.app = app
	this.log = app.Logger
	this.services = make(map[string]gopi.RPCServiceRecord)
	this.commands = make([]*Command, 0)
	this.stubs = make([]rpc.DiscoveryClient, 0)
	this.cancels = make([]context.CancelFunc, 0)
	this.errors = make(chan error)

	// Task to receive messages
	this.Tasks.Start(this.EventTask)

	return this
}

func (this *Runner) Close() error {
	// Call cancels
	for _, cancel := range this.cancels {
		cancel()
	}

	// Wait until all streams are completed
	this.WaitGroup.Wait()

	// Stop tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Release resources
	close(this.errors)
	this.services = nil
	this.commands = nil
	this.cancels = nil
	this.stubs = nil

	// return success
	return nil
}

// Add commands
func (this *Runner) RegisterCommand(cmd *Command) {
	this.commands = append(this.commands, cmd)
}

// Return command-line args
func (this *Runner) Args() []string {
	return this.app.AppFlags.Args()
}

// Match an argument and return
func (this *Runner) GetCommand(name string, scope CommandScope) (*Command, []string) {
	for _, cmd := range this.commands {
		if scope != cmd.scope {
			continue
		}
		if matches := cmd.re.FindStringSubmatch(name); matches != nil {
			return cmd, matches
		}
	}
	return nil, nil
}

// Pool returns the client pool or nil
func (this *Runner) Pool() gopi.RPCClientPool {
	if pool, ok := this.app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool); ok == false || pool == nil {
		return nil
	} else {
		return pool
	}
}

// Return discovery type
func (this *Runner) DiscoveryType() rpc.DiscoveryType {
	if dns, _ := this.app.AppFlags.GetBool("dns"); dns {
		return rpc.DISCOVERY_TYPE_DNS
	} else {
		return rpc.DISCOVERY_TYPE_DB
	}
}

// AddStub adds a stub to watch
func (this *Runner) AddStub(stub rpc.DiscoveryClient, service string) error {
	this.stubs = append(this.stubs, stub)

	// Create background task to stream messages
	ctx, cancel := context.WithCancel(context.Background())
	this.cancels = append(this.cancels, cancel)
	go func() {
		this.WaitGroup.Add(1)
		this.Merger.Merge(stub)
		if err := stub.StreamEvents(ctx, service); err != nil && err != context.Canceled {
			this.errors <- err
		}
		this.Merger.Unmerge(stub)
		this.WaitGroup.Done()
	}()

	// return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func Main(app *gopi.AppInstance, services []gopi.RPCServiceRecord, done chan<- struct{}) error {
	runner := NewRunner(app)
	defer runner.Close()

	if stub, err := runner.Pool().NewClientEx("gopi.Discovery", services, gopi.RPC_FLAG_NONE); err != nil {
		return err
	} else if err := runner.Run(stub.(rpc.DiscoveryClient)); err != nil {
		return err
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	// Create the configuration
	config := gopi.NewAppConfig("rpc/discovery:client", "rpc/version:client", "googlecast:client", "discovery")

	// Set subtype as "discovery"
	//config.AppFlags.SetParam(gopi.PARAM_SERVICE_SUBTYPE, "discovery")

	// Set flags
	config.AppFlags.FlagBool("dns", false, "Use DNS lookup rather than cache")

	// Run the command line tool - look for services for 400ms
	os.Exit(rpc.Client(config, DISCOVERY_TIMEOUT, Main))
}
