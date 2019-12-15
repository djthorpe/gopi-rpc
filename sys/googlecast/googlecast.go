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
	"math/rand"
	"net"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	"github.com/djthorpe/gopi/util/errors"
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
	channels  map[*castconn]*castdevice
	event.Publisher
	event.Tasks
	sync.WaitGroup
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
	this.channels = make(map[*castconn]*castdevice)

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

	errs := errors.CompoundError{}

	// Close channels
	for channel := range this.channels {
		errs.Add(this.Disconnect(channel))
	}

	// Waitgroup
	this.WaitGroup.Wait()

	// Stop background tasks
	if err := this.Tasks.Close(); err != nil {
		errs.Add(err)
	}

	// Unsubscribe
	this.Publisher.Close()

	// Release resources
	this.channels = nil
	this.devices = nil

	// Return any errors caught
	return errs.ErrorOrSelf()
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

func (this *googlecast) Connect(device rpc.GoogleCastDevice, flag gopi.RPCFlag, timeout time.Duration) (rpc.GoogleCastChannel, error) {
	this.log.Debug("<googlecast.Connect>{ device=%v flag=%v timeout=%v }", device, flag, timeout)

	if device_, ok := device.(*castdevice); device_ == nil || ok == false {
		return nil, gopi.ErrBadParameter
	} else if ip, err := getAddr(device_, flag); err != nil {
		return nil, err
	} else if channel, err := gopi.Open(CastConn{
		Addr:    ip.String(),
		Port:    uint16(device_.port),
		Timeout: timeout,
	}, this.log); err != nil {
		return nil, fmt.Errorf("Connect: %w", err)
	} else if channel_, ok := channel.(*castconn); ok == false {
		return nil, gopi.ErrAppError
	} else {
		// Watch channel for messages
		evt := channel_.Subscribe()
		// Connect for messages
		if err := channel_.Connect(); err != nil {
			channel.Close()
			return nil, err
		}
		go this.watchEvents(evt)

		// Add channel
		this.channels[channel.(*castconn)] = device_

		// Return channel
		return channel_, nil
	}
}

func (this *googlecast) Disconnect(channel rpc.GoogleCastChannel) error {
	this.log.Debug("<googlecast.Disconnect>{ channel=%v }", channel)

	if channel_, ok := channel.(*castconn); channel_ == nil || ok == false {
		return gopi.ErrBadParameter
	} else if err := channel_.Disconnect(); err != nil {
		return err
	} else {
		// Remove channel from list of channels
		for chan_ := range this.channels {
			if chan_ == channel_ {
				delete(this.channels, chan_)
				return channel_.Close()
			}
		}
		// Channel not found, return error
		return gopi.ErrBadParameter
	}
}

func (this *googlecast) watchEvents(evts <-chan gopi.Event) {
	this.WaitGroup.Add(1)
FOR_LOOP:
	for {
		select {
		case evt := <-evts:
			if evt == nil {
				continue
			}

			evt_ := evt.(rpc.GoogleChannelEvent)

			// re-emit event
			this.Emit(NewChannelEvent(this, evt_.Type(), evt_.Channel()))

			// End loop if disconnect
			if evt_.Type() == rpc.GOOGLE_CAST_EVENT_DISCONNECT {
				//emit and break
				break FOR_LOOP
			}
		}
	}
	this.WaitGroup.Done()
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

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func getAddr(device *castdevice, flag gopi.RPCFlag) (net.IP, error) {
	switch flag & (gopi.RPC_FLAG_INET_V4 | gopi.RPC_FLAG_INET_V6) {
	case gopi.RPC_FLAG_INET_V4:
		if len(device.ip4) == 0 {
			return nil, gopi.ErrNotFound
		} else if flag&gopi.RPC_FLAG_SERVICE_ANY != 0 {
			// Return any
			index := rand.Intn(len(device.ip4) - 1)
			return device.ip4[index], nil
		} else {
			// Return first
			return device.ip4[0], nil
		}
	case gopi.RPC_FLAG_INET_V6:
		if len(device.ip6) == 0 {
			return nil, gopi.ErrNotFound
		} else if flag&gopi.RPC_FLAG_SERVICE_ANY != 0 {
			// Return any
			index := rand.Intn(len(device.ip6) - 1)
			return device.ip6[index], nil
		} else {
			// Return first
			return device.ip6[0], nil
		}
	case (gopi.RPC_FLAG_INET_V6 | gopi.RPC_FLAG_INET_V4):
		if addr, err := getAddr(device, gopi.RPC_FLAG_INET_V4); err == nil {
			return addr, nil
		} else if addr, err := getAddr(device, gopi.RPC_FLAG_INET_V6); err == nil {
			return addr, nil
		} else {
			return nil, gopi.ErrNotFound
		}
	default:
		return nil, gopi.ErrBadParameter
	}
}
