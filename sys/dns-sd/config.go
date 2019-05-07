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
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	sync.Mutex
	event.Tasks
	event.Publisher

	// Public members
	Services []rpc.ServiceRecord `json:"services"`

	// Private members
	errors   chan<- error
	modified bool
	path     string
	source   gopi.Driver
	util     rpc.Util
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	FILENAME_DEFAULT = "discovery.json"
	WRITE_DELTA      = 5 * time.Second
	EXPIRE_DELTA     = 60 * time.Second
)

////////////////////////////////////////////////////////////////////////////////
// INIT / DEINIT

func (this *Config) Init(config Discovery, source gopi.Driver, errors chan<- error) error {
	this.Services = make([]rpc.ServiceRecord, 0)
	this.modified = false

	if config.Util == nil {
		return gopi.ErrBadParameter
	} else {
		this.util = config.Util
	}

	// Allow nil for source and errors
	this.source = source
	this.errors = errors

	// Read or create file
	if config.Path != "" {
		if err := this.CreatePath(config.Path); err != nil {
			return fmt.Errorf("Error: %v: %v", config.Path, err)
		}
	}

	this.Tasks.Start(this.WriteConfigTask, this.ExpireTask)

	return nil
}

func (this *Config) Destroy() error {

	// Unsubscribe listeners
	this.Publisher.Close()

	// Stop all tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// TODO: write if modified

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this Config) String() string {
	params := ""
	if this.path != "" {
		params += fmt.Sprintf("path=%v ", strconv.Quote(this.path))
	}
	return fmt.Sprintf("<Config>{ %vnum_services=%v }", params, len(this.Services))
}

////////////////////////////////////////////////////////////////////////////////
// READ AND WRITE CONFIG

// Create path if it doesn't exist
func (this *Config) CreatePath(path string) error {
	// Append home directory if relative path
	if filepath.IsAbs(path) == false {
		if homedir, err := os.UserHomeDir(); err != nil {
			return err
		} else {
			path = filepath.Join(homedir, path)
		}
	}

	// Set path
	this.path = path

	// Append filename
	if stat, err := os.Stat(this.path); err == nil && stat.IsDir() {
		// append default filename
		this.path = filepath.Join(this.path, FILENAME_DEFAULT)
	}

	// Read file
	if stat, err := os.Stat(this.path); err == nil && stat.Mode().IsRegular() {
		if err := this.Read(this.path); err != nil {
			return err
		} else {
			return nil
		}
	} else if os.IsNotExist(err) {
		// Create file
		if fh, err := os.Create(this.path); err != nil {
			return err
		} else if err := fh.Close(); err != nil {
			return err
		}
	}

	// Success
	return nil
}

// Set the modified flag to true
func (this *Config) SetModified() {
	this.Lock()
	defer this.Unlock()
	this.modified = true
}

// Write the configuration file to disk
func (this *Config) Write(path string, indent bool) error {
	this.Lock()
	defer this.Unlock()
	if fh, err := os.Create(path); err != nil {
		return err
	} else {
		defer fh.Close()
		if err := this.util.Writer(fh, this.Services, indent); err != nil {
			return err
		} else {
			this.modified = false
		}
	}

	// Success
	return nil
}

// Read the configuration from a file
func (this *Config) Read(path string) error {
	this.Lock()
	defer this.Unlock()
	if fh, err := os.Open(path); err != nil {
		return err
	} else {
		defer fh.Close()
		if records, err := this.util.Reader(fh); err != nil {
			return err
		} else {
			this.Services = this.UnexpiredServices(records)
			this.modified = false
		}
	}

	// Success
	return nil
}

// UnexpiredServices returns an array of unexpired services
func (this *Config) UnexpiredServices(records []rpc.ServiceRecord) []rpc.ServiceRecord {
	services := make([]rpc.ServiceRecord, 0, len(this.Services))
	for _, service := range records {
		if service.Expired() == false {
			services = append(services, service)
		}
	}
	return services
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *Config) WriteConfigTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	ticker := time.NewTicker(WRITE_DELTA)
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			if this.modified {
				if this.path == "" {
					// Do nothing
				} else if err := this.Write(this.path, true); err != nil && this.errors != nil {
					this.errors <- err
				}
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	// Stop the ticker
	ticker.Stop()

	// Success
	return nil
}

func (this *Config) ExpireTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	ticker := time.NewTimer(time.Second)
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			for _, service := range this.Services {
				if service.Expired() {
					this.Emit(this.util.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_EXPIRED, service))
				}
			}
			if services := this.UnexpiredServices(this.Services); len(services) != len(this.Services) {
				this.Services = services
				this.SetModified()
			}
			ticker.Reset(EXPIRE_DELTA)
		case <-stop:
			break FOR_LOOP
		}
	}

	// Stop the ticker
	ticker.Stop()

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RETURN SERVICES

// EnumerateServices returns the names of discovered service types
// and filters to those which have been registered locally (rather
// than through DNS service discovery) if flag is set
func (this *Config) EnumerateServices(source rpc.DiscoveryType) []string {
	// Get all the records
	type_map := make(map[string]bool, 10)
	for _, record := range this.GetServices("", source) {
		type_map[record.Service()] = true
	}
	records := make([]string, 0, len(type_map))
	for key := range type_map {
		records = append(records, key)
	}
	return records
}

func (this *Config) GetServices(service string, source rpc.DiscoveryType) []rpc.ServiceRecord {
	this.Lock()
	defer this.Unlock()

	records := make([]rpc.ServiceRecord, 0)
	for _, record := range this.Services {
		if source != rpc.DISCOVERY_TYPE_NONE && source != record.Source() {
			continue
		}
		if service != "" && service != record.Service() {
			continue
		}
		if record.Expired() {
			continue
		}
		records = append(records, record)
	}
	return records
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER & REMOVE SERVICES

func (this *Config) Register_(service rpc.ServiceRecord) error {
	if service == nil || service.Key() == "" {
		return gopi.ErrBadParameter
	}
	if service.Service() == rpc.DISCOVERY_SERVICE_QUERY {
		this.Emit(this.util.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_NAME, service))
	} else if index := this.IndexForService(service); index == -1 {
		this.Lock()
		this.Services = append(this.Services, service)
		this.Unlock()
		this.SetModified()
		this.Emit(this.util.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_ADDED, service))
	} else {
		this.Lock()
		this.Services[index] = service
		this.Unlock()
		this.SetModified()
		this.Emit(this.util.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_UPDATED, service))
	}

	// Success
	return nil
}

func (this *Config) Remove_(service rpc.ServiceRecord) error {
	if service == nil || service.Key() == "" {
		return gopi.ErrBadParameter
	}
	if index := this.IndexForService(service); index == -1 {
		return gopi.ErrNotModified
	} else if err := this.RemoveAtIndex(index); err != nil {
		return err
	} else {
		this.Emit(this.util.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_REMOVED, service))
	}

	// Success
	return nil
}

func (this *Config) RemoveAtIndex(index int) error {
	if index < 0 || index >= len(this.Services) {
		return gopi.ErrBadParameter
	}

	// Remove the record
	this.Services = append(this.Services[:index], this.Services[index+1:]...)
	this.SetModified()
	return nil
}

func (this *Config) IndexForService(service rpc.ServiceRecord) int {
	if service == nil {
		return -1
	}
	for i, s := range this.Services {
		if service.Key() == s.Key() {
			return i
		}
	}
	// Return 'not found'
	return -1
}
