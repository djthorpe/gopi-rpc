/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
	dns "github.com/miekg/dns"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Discovery struct {
	Path      string
	Interface *net.Interface
	Domain    string
	Flags     gopi.RPCFlag
}

type discovery struct {
	sync.Mutex
	event.Tasks
	Config
	Listener

	errors   chan error
	services chan *rpc.ServiceRecord
	log      gopi.Logger
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Discovery) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<rpc.discovery.Open>{ interface=%v domain='%v' }", config.Interface, config.Domain)

	this := new(discovery)
	this.errors = make(chan error)
	this.services = make(chan *rpc.ServiceRecord)

	if err := this.Config.Init(this, config.Path, this.errors); err != nil {
		return nil, err
	} else if err := this.Listener.Init(config, this.errors, this.services); err != nil {
		return nil, err
	} else {
		this.log = logger
	}

	// Start task to catch errors, receive services and expire records
	this.Tasks.Start(this.BackgroundTask)

	return this, nil
}

func (this *discovery) Close() error {
	this.log.Debug("<rpc.discovery.Close>{ config=%v listener=%v }", this.Config, this.Listener)

	// Unsubscribe
	this.Publisher.Close()

	// Release resources, etc
	if err := this.Listener.Destroy(); err != nil {
		return err
	}
	if err := this.Config.Destroy(); err != nil {
		return err
	}
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER AND BROWSE

// Register a service record
func (this *discovery) Register(service gopi.RPCServiceRecord) error {
	this.log.Debug2("<rpc.discovery.Register>{ service=%v }", service)
	if service == nil {
		return gopi.ErrBadParameter
	}
	return gopi.ErrNotImplemented
}

// Lookup service instances from service name
func (this *discovery) Lookup(ctx context.Context, service string) ([]gopi.RPCServiceRecord, error) {
	this.log.Debug2("<rpc.discovery.Lookup>{ service=%v }", strconv.Quote(service))

	// The message should be to lookup service by name
	msg := new(dns.Msg)
	msg.SetQuestion(service+"."+this.domain, dns.TypePTR)
	msg.RecursionDesired = false

	// Wait for services
	services := make(map[string]gopi.RPCServiceRecord, 0)
	stop := make(chan struct{})
	go func() {
		evts := this.Subscribe()
	FOR_LOOP:
		for {
			select {
			case evt := <-evts:
				if evt_, ok := evt.(gopi.RPCEvent); ok {
					s := evt_.ServiceRecord()
					if s != nil && s.Service() == service {
						key := s.Name()
						switch evt_.Type() {
						case gopi.RPC_EVENT_SERVICE_ADDED, gopi.RPC_EVENT_SERVICE_UPDATED:
							services[key] = evt_.ServiceRecord()
						case gopi.RPC_EVENT_SERVICE_REMOVED, gopi.RPC_EVENT_SERVICE_EXPIRED:
							delete(services, key)
						}
					}
				}
			case <-stop:
				break FOR_LOOP
			}
		}
		this.Unsubscribe(evts)
		close(stop)
	}()

	// Perform the query and wait for cancellation
	err := this.Query(msg, ctx)

	// Retrieve the service records
	records := make([]gopi.RPCServiceRecord, 0, len(services))
	for _, record := range services {
		records = append(records, record)
	}

	// Return error or nil on success
	if err == nil || err == context.Canceled || err == context.DeadlineExceeded {
		return records, nil
	} else {
		return nil, err
	}
}

// Enumerate Services
func (this *discovery) EnumerateServices(ctx context.Context) ([]string, error) {
	this.log.Debug2("<rpc.discovery.EnumerateServices>{ }")

	// The message should be to enumerate services
	msg := new(dns.Msg)
	msg.SetQuestion(MDNS_SERVICE_QUERY+"."+this.domain, dns.TypePTR)
	msg.RecursionDesired = false

	// Wait for services
	services := make(map[string]bool, 0)
	stop := make(chan struct{})
	go func() {
		evts := this.Subscribe()
	FOR_LOOP:
		for {
			select {
			case evt := <-evts:
				if evt_, ok := evt.(gopi.RPCEvent); ok && evt_.Type() == gopi.RPC_EVENT_SERVICE_NAME {
					name := evt_.ServiceRecord().Name()
					services[name] = true
				}
			case <-stop:
				break FOR_LOOP
			}
		}
		this.Unsubscribe(evts)
		close(stop)
	}()

	// Perform the query and wait for cancellation
	err := this.Query(msg, ctx)

	// Stop collecting names
	stop <- gopi.DONE
	<-stop

	// Retrieve the service names
	keys := make([]string, 0, len(services))
	for key := range services {
		keys = append(keys, key)
	}

	// Return error or nil on success
	if err == nil || err == context.Canceled || err == context.DeadlineExceeded {
		return keys, nil
	} else {
		return nil, err
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *discovery) String() string {
	return fmt.Sprintf("<rpc.discovery>{ config=%v listener=%v }", this.Config, this.Listener)
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *discovery) BackgroundTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	this.log.Debug("BackgroundTask started")
	start <- gopi.DONE

	// TODO
FOR_LOOP:
	for {
		select {
		case err := <-this.errors:
			this.log.Warn("Discover error: %v", err)
		case service := <-this.services:
			if service.TTL() == 0 {
				this.Config.Remove(service)
			} else {
				this.Config.Register(service)
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	// Success
	this.log.Debug("BackgroundTask completed")
	return nil
}
