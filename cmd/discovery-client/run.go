/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	tablewriter "github.com/olekukonko/tablewriter"
)

////////////////////////////////////////////////////////////////////////////////

func (this *Runner) Run(stub rpc.DiscoveryClient) error {
	this.log.Info("Connected to %v", stub.Conn().Addr())

	// By default, enumerate the services
	this.RegisterCommand(&Command{
		re:    regexp.MustCompile("^_$"),
		scope: COMMAND_SCOPE_ROOT,
		cb:    this.EnumerateServices,
	})
	// Watch for discovery events
	this.RegisterCommand(&Command{
		re:    regexp.MustCompile("^watch$"),
		scope: COMMAND_SCOPE_ROOT,
		cb:    this.Watch,
	})
	// Lookup service
	this.RegisterCommand(&Command{
		re:    regexp.MustCompile("^_([A-Za-z]\\w+)\\._tcp$"),
		scope: COMMAND_SCOPE_ROOT,
		cb:    this.ServiceCommands,
	})
	this.RegisterCommand(&Command{
		re:    regexp.MustCompile("^([A-Za-z]\\w+)$"),
		scope: COMMAND_SCOPE_ROOT,
		cb:    this.ServiceCommands,
	})

	args := this.Args()
	if len(args) == 0 {
		// Default command is enumerate services
		return this.EnumerateServices(stub, nil)
	} else if command, matches := this.GetCommand(args[0], COMMAND_SCOPE_ROOT); command == nil {
		return gopi.ErrHelp
	} else {
		return command.cb(stub, matches[1:])
	}
}

func (this *Runner) Watch(stub rpc.DiscoveryClient, _ []string) error {
	fmt.Println("WATCH")
	return nil
}

func (this *Runner) EnumerateServices(stub rpc.DiscoveryClient, _ []string) error {
	timeout, _ := this.app.AppFlags.GetDuration("timeout")
	if services, err := stub.Enumerate(this.DiscoveryType(), this.TimeoutOrDefault(timeout)); err != nil {
		return err
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Service"})
		for _, service := range services {
			table.Append([]string{service})
		}
		table.Render()
	}

	// Success
	return nil
}

func (this *Runner) ServiceCommands(stub rpc.DiscoveryClient, service []string) error {
	if len(service) == 1 {
		return this.ServiceLookup(stub, service)
	} else {
		return gopi.ErrHelp
	}
}

func (this *Runner) ServiceLookup(stub rpc.DiscoveryClient, service []string) error {
	timeout, _ := this.app.AppFlags.GetDuration("timeout")
	if len(service) != 1 {
		return gopi.ErrBadParameter
	} else if services, err := stub.Lookup(fmt.Sprintf("_%v._tcp", service[0]), this.DiscoveryType(), this.TimeoutOrDefault(timeout)); err != nil {
		return err
	} else if len(services) == 0 {
		return fmt.Errorf("No service records found for %v", strconv.Quote(service[0]))
	} else {
		PrintServices(os.Stdout, services)
	}

	// Success
	return nil
}

/*

func RenderHost(service gopi.RPCServiceRecord) string {
	if service.Port() == 0 {
		return service.Host()
	} else {
		return fmt.Sprintf("%v:%v", service.Host(), service.Port())
	}
}

func RenderService(service gopi.RPCServiceRecord) string {
	if service.Subtype() == "" {
		return service.Service()
	} else {
		return fmt.Sprintf("%v, %v", service.Subtype(), service.Service())
	}
}

func RenderIP(service gopi.RPCServiceRecord) string {
	ips := make([]string, 0)
	for _, ip := range service.IP4() {
		ips = append(ips, fmt.Sprint(ip))
	}
	for _, ip := range service.IP6() {
		ips = append(ips, fmt.Sprint(ip))
	}
	return strings.Join(ips, "\n")
}

func RenderTxt(service gopi.RPCServiceRecord) string {
	return strings.Join(service.Text(), "\n")
}

func Conn(app *gopi.AppInstance) (gopi.RPCClientConn, error) {
	// Return a single connection
	addr, _ := app.AppFlags.GetString("addr")
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)

	// If service is nil, then use the address
	if records, err := pool.Lookup(ctx, "", addr, 1); err != nil {
		return nil, err
	} else if len(records) == 0 {
		return nil, gopi.ErrDeadlineExceeded
	} else if conn, err := pool.Connect(records[0], 0); err != nil {
		return nil, err
	} else {
		return conn, nil
	}
}

func DiscoveryClient(app *gopi.AppInstance) (rpc.DiscoveryClient, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := Conn(app); err != nil {
		return nil, err
	} else if client := pool.NewClient("gopi.Discovery", conn); client == nil {
		return nil, gopi.ErrNotFound
	} else if client_, ok := client.(rpc.DiscoveryClient); client_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client_, nil
	}
}

func VersionClient(app *gopi.AppInstance) (rpc.VersionClient, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := Conn(app); err != nil {
		return nil, err
	} else if client := pool.NewClient("gopi.Version", conn); client == nil {
		return nil, gopi.ErrNotFound
	} else if client_, ok := client.(rpc.VersionClient); client_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client_, nil
	}
}

func GoogleCastClient(app *gopi.AppInstance) (rpc.GoogleCastClient, error) {
	pool := app.ModuleInstance("rpc/clientpool").(gopi.RPCClientPool)

	if conn, err := Conn(app); err != nil {
		return nil, err
	} else if client := pool.NewClient("gopi.GoogleCast", conn); client == nil {
		return nil, gopi.ErrNotFound
	} else if client_, ok := client.(rpc.GoogleCastClient); client_ == nil || ok == false {
		return nil, gopi.ErrAppError
	} else {
		return client_, nil
	}
}

func Watch(app *gopi.AppInstance, start chan<- struct{}, stop <-chan struct{}) error {
	start <- gopi.DONE
	service := ""

	if watch, _ := app.AppFlags.GetBool("watch"); watch == false {
		return nil
	}
	if args := app.AppFlags.Args(); len(args) == 1 {
		service = args[0]
	}

	events := make(chan gopi.RPCEvent)
	go func() {
		fmt.Println("START")
	FOR_LOOP:
		for {
			select {
			case evt := <-events:
				fmt.Println(evt)
			case <-stop:
				break FOR_LOOP
			}
		}
		fmt.Println("STOP")
	}()

	if client, err := DiscoveryClient(app); err != nil {
		return err
	} else if err := client.StreamEvents(service, events); err != nil {
		return err
	}

	// Success
	return nil
}

func Register(app *gopi.AppInstance, client rpc.DiscoveryClient, name, service, subtype string, port uint) error {
	util := app.ModuleInstance("rpc/util").(rpc.Util)
	record := util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB)

	if err := record.SetName(name); err != nil {
		return err
	} else if hostname, err := os.Hostname(); err != nil {
		return err
	} else if addrs, err := net.LookupHost(hostname); err != nil {
		return err
	} else if len(addrs) == 0 {
		return fmt.Errorf("No addresses found for host %v", strconv.Quote(hostname))
	} else if err := record.SetHostPort(hostname + ":" + fmt.Sprint(port)); err != nil {
		return err
	} else {
		for _, addr := range addrs {
			if ip := net.ParseIP(addr); ip == nil {
				return fmt.Errorf("Cannot parse %v", strconv.Quote(addr))
			} else if err := record.AppendIP(ip); err != nil {
				return err
			}
		}
	}
	if err := record.SetService(service, subtype); err != nil {
		return err
	}

	fmt.Println(record)

	return client.Register(record)
}

func Main(app *gopi.AppInstance, done chan<- struct{}) error {
	t := rpc.DISCOVERY_TYPE_DB
	r := ""
	service := ""
	if dns, _ := app.AppFlags.GetBool("dns"); dns {
		t = rpc.DISCOVERY_TYPE_DNS
	}
	if register_name, _ := app.AppFlags.GetString("register"); register_name != "" {
		r = register_name
	}
	if args := app.AppFlags.Args(); len(args) == 1 {
		service = args[0]
	} else if len(args) > 1 {
		return gopi.ErrHelp
	}

	if client, err := DiscoveryClient(app); err != nil {
		return err
	} else if err := client.Ping(); err != nil {
		return err
	} else if service == "" {
		if r != "" {
			return fmt.Errorf("Missing service parameter")
		}
		if services, err := client.Enumerate(t, 0); err != nil {
			return err
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Service"})
			for _, service := range services {
				table.Append([]string{service})
			}
			table.Render()
		}
	} else {
		if r != "" {
			// Perform registration
			subtype, _ := app.AppFlags.GetString("subtype")
			port, _ := app.AppFlags.GetUint("port")
			if port == 0 {
				return fmt.Errorf("Missing -port parameter")
			}
			if err := Register(app, client, r, service, subtype, port); err != nil {
				return err
			}
		}
		if instances, err := client.Lookup(service, t, 0); err != nil {
			return err
		} else {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Service", "Name", "Host", "Addr", "TXT"})
			for _, instance := range instances {
				table.Append([]string{
					RenderService(instance),
					instance.Name(),
					RenderHost(instance),
					RenderIP(instance),
					RenderTxt(instance),
				})
			}
			table.Render()
		}
	}

	if service == "_googlecast._tcp" {
		if client, err := GoogleCastClient(app); err != nil {
			return err
		} else if err := client.Ping(); err != nil {
			return err
		} else {
			if devices, err := client.Devices(); err != nil {
				return err
			} else {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Name", "Service", "State", "Model"})
				for _, device := range devices {
					table.Append([]string{
						device.Name(),
						device.Service(),
						fmt.Sprint(device.State()),
						device.Model(),
					})
				}
				table.Render()
			}
		}
	}

	// Success
	return nil
}
*/
