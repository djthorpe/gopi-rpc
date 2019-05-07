/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	"context"
	"fmt"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type GoogleCast struct {
	Discovery gopi.RPCServiceDiscovery
}

type googlecast struct {
	log       gopi.Logger
	discovery gopi.RPCServiceDiscovery
	devices   map[string]*castdevice
	event.Publisher
	event.Tasks
}

////////////////////////////////////////////////////////////////////////////////
// COMSTANTS

const (
	SERVICE_TYPE_GOOGLECAST = "_googlecast._tcp"
	DELTA_LOOKUP_TIME       = 60 * time.Second
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config GoogleCast) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<googlecast.Open>{ discovery=%v }", config.Discovery)

	this := new(googlecast)
	this.log = logger
	this.discovery = config.Discovery
	this.devices = make(map[string]*castdevice)

	if this.discovery == nil {
		return nil, gopi.ErrBadParameter
	}

	// Run background tasks
	this.Tasks.Start(this.Watch, this.Lookup)

	// Success
	return this, nil
}

func (this *googlecast) Close() error {
	this.log.Debug("<googlecast.Close>{ }")

	// Stop background tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Unsubscribe
	this.Publisher.Close()

	// Release resources
	this.devices = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// INTERFACE IMPLEMENTATION

func (this *googlecast) Devices() []rpc.GoogleCastDevice {
	devices := make([]rpc.GoogleCastDevice, 0, len(this.devices))
	for _, device := range this.devices {
		devices = append(devices, device)
	}
	return devices
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *googlecast) String() string {
	return fmt.Sprintf("<googlecast>{ devices=%v }", this.Devices())
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *googlecast) Lookup(start chan<- event.Signal, stop <-chan event.Signal) error {
	this.log.Debug("<googlecast.Lookup> Started")
	start <- gopi.DONE

	// Periodically lookup Googlecast devices
	timer := time.NewTimer(100 * time.Millisecond)
FOR_LOOP:
	for {
		select {
		case <-timer.C:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			this.discovery.Lookup(ctx, SERVICE_TYPE_GOOGLECAST)
			cancel()
			timer.Reset(DELTA_LOOKUP_TIME)
		case <-stop:
			break FOR_LOOP
		}
	}
	this.log.Debug("<googlecast.Lookup> Stopped")
	return nil
}

func (this *googlecast) Watch(start chan<- event.Signal, stop <-chan event.Signal) error {
	this.log.Debug("<googlecast.Watch> Started")
	start <- gopi.DONE

	events := this.discovery.Subscribe()
FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt_, ok := evt.(gopi.RPCEvent); ok {
				this.WatchEvent(evt_)
			}
		case <-stop:
			break FOR_LOOP
		}
	}
	this.discovery.Unsubscribe(events)
	this.log.Debug("<googlecast.Watch> Stopped")
	return nil
}

func (this *googlecast) WatchEvent(evt gopi.RPCEvent) error {
	if service := evt.ServiceRecord(); service == nil || service.Service() != SERVICE_TYPE_GOOGLECAST {
		return nil
	} else if cast := NewCastDevice(service); cast == nil {
		return nil
	} else if evt.Type() == gopi.RPC_EVENT_SERVICE_EXPIRED {
		this.log.Debug2("<googlecast.Watch> Expired: %v", cast)
		this.Emit(NewCastEvent(this, gopi.RPC_EVENT_SERVICE_EXPIRED, cast))
		delete(this.devices, cast.Id())
	} else if evt.Type() == gopi.RPC_EVENT_SERVICE_ADDED || evt.Type() == gopi.RPC_EVENT_SERVICE_UPDATED {
		name := cast.Id()
		if other, exists := this.devices[name]; exists == false {
			this.log.Debug2("<googlecast.Watch> Added: %v", cast)
			this.Emit(NewCastEvent(this, gopi.RPC_EVENT_SERVICE_ADDED, cast))
		} else if cast.EqualsState(other) == false {
			this.log.Debug2("<googlecast.Watch> Updated: %v", cast)
			this.Emit(NewCastEvent(this, gopi.RPC_EVENT_SERVICE_UPDATED, cast))
		}
		this.devices[name] = cast
	}
	return nil
}
