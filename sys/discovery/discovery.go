/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"fmt"
	"net"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	iface "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Discovery struct {
	Interface *net.Interface
	Domain    string
	Flags     iface.RPCFlags
}

type discovery struct {
	sync.Mutex
	event.Tasks
	Config
	Listener

	errors   chan error
	services chan *ServiceRecord
	log      gopi.Logger
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Discovery) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<rpc.discovery.Open>{ interface=%v domain='%v' }", config.Interface, config.Domain)

	this := new(discovery)
	this.errors = make(chan error)
	this.services = make(chan *ServiceRecord)

	if err := this.Config.Init(); err != nil {
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
			this.log.Warn("Error: %v", err)
		case service := <-this.services:
			if service.ttl == 0 {
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

/*

start chan<- event.Signal, stop <-chan event.Signal) error {
	this.log.Debug("START ttl_expire")
	start <- gopi.DONE

	timer := time.NewTicker(500 * time.Millisecond)

FOR_LOOP:
	for {
		select {
		case <-timer.C:
			// look for expiring TTL records in a very non-optimal way
			expired_keys := make([]string, 0, 1)
			for _, entry := range this.entries {
				if time.Now().After(entry.Timestamp.Add(entry.TTL)) {
					expired_keys = append(expired_keys, entry.Key)
				}
			}
			for _, key := range expired_keys {
				fmt.Printf("EXP: %v\n", this.entries[key])
				delete(this.entries, key)
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	this.log.Debug("STOP ttl_expire")
	return nil
}
*/
