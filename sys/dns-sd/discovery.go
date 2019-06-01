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
	"strings"
	"sync"
	"time"

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
	Interface net.Interface
	Domain    string
	Flags     gopi.RPCFlag
	Util      rpc.Util
}

type discovery struct {
	sync.Mutex
	event.Tasks
	Config
	Listener

	errors    chan error
	services  chan rpc.ServiceRecord
	questions chan Question
	log       gopi.Logger
	util      rpc.Util
}

////////////////////////////////////////////////////////////////////////////////
// CONFIGURATION

const (
	DELTA_PROBE = 120 * time.Second
	DEFAULT_TTL = 15 * time.Minute
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Discovery) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<rpc.discovery.dns-sd.Open>{ interface=%v domain='%v' flags=%v }", config.Interface, config.Domain, config.Flags)

	this := new(discovery)
	this.errors = make(chan error)
	this.services = make(chan rpc.ServiceRecord)
	this.questions = make(chan Question)

	if err := this.Config.Init(config, this, this.errors); err != nil {
		logger.Debug2("Config.Init returned nil")
		return nil, err
	} else if err := this.Listener.Init(config, this.errors, this.services, this.questions); err != nil {
		logger.Debug2("Listener.Init returned nil")
		return nil, err
	} else if config.Util == nil {
		logger.Debug2("config.Util == nil")
		return nil, gopi.ErrBadParameter
	} else {
		this.log = logger
		this.util = config.Util
	}

	// Start task to catch errors, receive services,
	// expire records and send probe requests
	this.Tasks.Start(this.BackgroundTask, this.ProbeTask)

	return this, nil
}

func (this *discovery) Close() error {
	this.log.Debug("<rpc.discovery.dns-sd.Close>{ config=%v listener=%v }", this.Config, this.Listener)

	// Remove registered services
	for _, service := range this.Config.GetServices("", rpc.DISCOVERY_TYPE_DB) {
		fmt.Println("TODO, de-register", service)
	}

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

	// Close channels
	close(this.errors)
	close(this.services)
	close(this.questions)

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER AND BROWSE

// Register a service record
func (this *discovery) Register(service gopi.RPCServiceRecord) error {
	this.log.Debug2("<rpc.discovery.dns-sd.Register>{ service=%v }", service)
	if service == nil || service.Name() == "" || service.Service() == "" {
		return gopi.ErrBadParameter
	}

	// Generate service name including subtype
	record := this.util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB)
	if err := record.SetName(service.Name()); err != nil {
		return err
	}
	if err := record.SetService(service.Service(), service.Subtype()); err != nil {
		return err
	}
	if err := record.SetHostPort(fmt.Sprintf("%v:%v", service.Host(), service.Port())); err != nil {
		return err
	}
	if err := record.AppendTXT(service.Text()...); err != nil {
		return err
	}
	if err := record.AppendIP(service.IP4()...); err != nil {
		return err
	}
	if err := record.AppendIP(service.IP6()...); err != nil {
		return err
	}

	// Configure in the registry
	if err := this.Config.Register_(record); err != nil {
		return err
	}

	// Success
	return nil
}

// Remove a service record
func (this *discovery) Remove(service gopi.RPCServiceRecord) error {
	this.log.Debug2("<rpc.discovery.dns-sd.Unregister>{ service=%v }", service)
	if service == nil || service.Name() == "" || service.Service() == "" {
		return gopi.ErrBadParameter
	}

	// Generate service name including subtype
	record := this.util.NewServiceRecord(rpc.DISCOVERY_TYPE_DB)
	if err := record.SetName(service.Name()); err != nil {
		return err
	}
	if err := record.SetService(service.Service(), service.Subtype()); err != nil {
		return err
	}

	// Remove from the registry
	if err := this.Config.Remove_(record); err != nil {
		return err
	} else {
		return nil
	}
}

// Lookup service instances from service name
func (this *discovery) Lookup(ctx context.Context, service string) ([]gopi.RPCServiceRecord, error) {
	this.log.Debug2("<rpc.discovery.dns-sd.Lookup>{ service=%v }", strconv.Quote(service))

	// The message should be to lookup service by name
	msg := new(dns.Msg)
	msg.SetQuestion(service+"."+this.domain+".", dns.TypePTR)
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
				// TODO: break loop when enough records have been collected
			case <-stop:
				break FOR_LOOP
			}
		}
		this.Unsubscribe(evts)
	}()

	// Perform the query and wait for cancellation, then stop background go routine
	err := this.QueryAll(msg, ctx)
	close(stop)

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
	this.log.Debug2("<rpc.discovery.dns-sd.EnumerateServices>{ }")

	// The message should be to enumerate services
	msg := new(dns.Msg)
	msg.SetQuestion(rpc.DISCOVERY_SERVICE_QUERY+"."+this.domain+".", dns.TypePTR)
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
	}()

	// Perform the query and wait for cancellation
	err := this.QueryAll(msg, ctx)

	// Stop collecting names
	stop <- gopi.DONE
	close(stop)

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

// ServiceInstances returns all cached servicerecords for a particular service name
func (this *discovery) ServiceInstances(service string) []gopi.RPCServiceRecord {
	this.log.Debug2("<rpc.discovery.dns-sd.ServiceRecords>{ service=%v }", strconv.Quote(service))
	records := make([]gopi.RPCServiceRecord, 0, len(this.Config.Services))
	for _, record := range this.Config.Services {
		if service == "" || (record.Service() == service && strings.TrimSpace(record.Service()) != "") {
			records = append(records, record)
		}
	}
	return records
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *discovery) String() string {
	return fmt.Sprintf("<rpc.discovery>{ config=%v listener=%v }", this.Config, this.Listener)
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *discovery) BackgroundTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	this.log.Debug2("<rpc.discovery.dns-sd.BackgroundTask> started")
	start <- gopi.DONE

FOR_LOOP:
	for {
		select {
		case err := <-this.errors:
			this.log.Warn("rpc.discovery: %v", err)
		case service := <-this.services:
			if service.TTL() == 0 {
				this.Config.Remove_(service)
			} else {
				this.Config.Register_(service)
			}
		case question := <-this.questions:
			if question.Query == rpc.DISCOVERY_SERVICE_QUERY {
				if services := this.Config.EnumerateServices(rpc.DISCOVERY_TYPE_DB); len(services) > 0 {
					if err := this.Listener.AnswerEnumMulticast(services, question, DEFAULT_TTL); err != nil {
						this.log.Warn("rpc.discovery: %v", err)
					}
				}
			} else {
				if instances := this.Config.GetServices(question.Query, rpc.DISCOVERY_TYPE_DB); len(instances) > 0 {
					if err := this.Listener.AnswerInstanceMulticast(instances, question, DEFAULT_TTL); err != nil {
						this.log.Warn("rpc.discovery: %v", err)
					}
				}
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	// Success
	this.log.Debug2("<rpc.discovery.dns-sd.BackgroundTask> completed")
	return nil
}

func (this *discovery) ProbeTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	this.log.Debug2("<rpc.discovery.dns-sd.ProbeTask> started")
	start <- gopi.DONE

	timer := time.NewTimer(10 * time.Second)
	services := []string{}
FOR_LOOP:
	for {
		select {
		case <-timer.C:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			if len(services) == 0 {
				if services_, err := this.EnumerateServices(ctx); err != nil {
					this.errors <- err
				} else {
					services = services_
				}
			} else {
				this.log.Debug2("<rpc.discovery.dns-sd.ProbeTask> Probing for %v", services[0])
				if _, err := this.Lookup(ctx, services[0]); err != nil {
					this.errors <- err
				}
				services = services[1:len(services)]
			}
			cancel()
			timer.Reset(DELTA_PROBE)
		case <-stop:
			break FOR_LOOP
		}
	}

	// Success
	timer.Stop()
	this.log.Debug2("<rpc.discovery.dns-sd.ProbeTask> completed")
	return nil
}
