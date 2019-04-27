/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	Services []*rpc.ServiceRecord `json:"services"`

	// Private members
	errors   chan<- error
	modified bool
	path     string
	source   gopi.Driver
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

func (this *Config) Init(source gopi.Driver, path string, errors chan<- error) error {
	this.Services = make([]*rpc.ServiceRecord, 0)
	this.modified = false

	// Check incoming parameters
	if errors == nil || source == nil {
		return gopi.ErrBadParameter
	}

	this.source = source
	this.errors = errors

	// Read or create file
	if path != "" {
		if err := this.CreatePath(path); err != nil {
			return err
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
	return fmt.Sprintf("<Config>{ %v num_services=%v }", strings.TrimSpace(params), len(this.Services))
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

	// Read file
	if stat, err := os.Stat(this.path); err == nil && stat.Mode().IsRegular() {
		if err := this.Read(this.path); err != nil {
			return err
		} else {
			return nil
		}
	} else if err == nil && stat.IsDir() {
		// append default filename
		this.path = filepath.Join(this.path, FILENAME_DEFAULT)
	}

	// Create file
	if fh, err := os.Create(this.path); err != nil {
		return err
	} else if err := fh.Close(); err != nil {
		return err
	} else {
		return nil
	}
}

// Set the modified flag to true
func (this *Config) SetModified() {
	this.Lock()
	defer this.Unlock()
	this.modified = true
}

// Write the configuration file to disk
func (this *Config) Write(path string, indent bool) error {
	if fh, err := os.Create(path); err != nil {
		return err
	} else {
		defer fh.Close()
		return this.Writer(fh, indent)
	}
}

// Writer writes the configuration to a io.Writer object
func (this *Config) Writer(fh io.Writer, indent bool) error {
	this.Lock()
	defer this.Unlock()

	enc := json.NewEncoder(fh)
	if indent {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(this); err != nil {
		return err
	}

	// Clear modified flag
	this.modified = false

	// Success
	return nil
}

// Read the configuration from a file
func (this *Config) Read(path string) error {
	// Read configuration
	if fh, err := os.Open(path); err != nil {
		return err
	} else {
		defer fh.Close()
		if config, err := this.Reader(fh); err != nil {
			return err
		} else {
			this.Lock()
			defer this.Unlock()
			this.Services = config.ExpireServices()
			this.modified = false
		}
	}

	// Success
	return nil
}

// Reader reads the configuration from an io.Reader object
func (this *Config) Reader(fh io.Reader) (*Config, error) {
	this.Lock()
	defer this.Unlock()

	dec := json.NewDecoder(fh)
	config := new(Config)
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}

	// Success
	return config, nil
}

// ExpireServices returns an array of unexpired services
func (this *Config) ExpireServices() []*rpc.ServiceRecord {
	services := make([]*rpc.ServiceRecord, 0, len(this.Services))
	for _, service := range this.Services {
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
				} else if err := this.Write(this.path, true); err != nil {
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
					this.Emit(rpc.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_EXPIRED, service))
				}
			}
			if services := this.ExpireServices(); len(services) != len(this.Services) {
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
// REGISTER & REMOVE SERVICES

func (this *Config) Register(service *rpc.ServiceRecord) error {
	if service == nil || service.TTL() == 0 || service.Key() == "" {
		return gopi.ErrBadParameter
	}
	if service.Service() == MDNS_SERVICE_QUERY {
		this.Emit(rpc.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_NAME, service))
	} else if index := this.IndexForService(service); index == -1 {
		this.Services = append(this.Services, service)
		this.SetModified()
		this.Emit(rpc.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_ADDED, service))
	} else {
		this.Services[index] = service
		this.SetModified()
		this.Emit(rpc.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_UPDATED, service))
	}

	// Success
	return nil
}

func (this *Config) Remove(service *rpc.ServiceRecord) error {
	if service == nil || service.Key() == "" {
		return gopi.ErrBadParameter
	}
	if index := this.IndexForService(service); index == -1 {
		return gopi.ErrNotModified
	} else if err := this.RemoveAtIndex(index); err != nil {
		return err
	} else {
		this.Emit(rpc.NewEvent(this.source, gopi.RPC_EVENT_SERVICE_REMOVED, service))
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

func (this *Config) IndexForService(service *rpc.ServiceRecord) int {
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
