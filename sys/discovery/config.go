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
	"io"
	"os"
	"sync"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	sync.Mutex

	// Public members
	Services []*rpc.ServiceRecord `json:"services"`

	// Private members
	modified bool
}

////////////////////////////////////////////////////////////////////////////////
// INIT / DEINIT

func (this *Config) Init() error {
	this.Services = make([]*rpc.ServiceRecord, 0)
	this.modified = false
	return nil
}

func (this *Config) Destroy() error {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// READ AND WRITE CONFIG

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
		if err := this.Reader(fh); err != nil {
			return err
		}
	}

	// Success
	return nil
}

// Reader reads the configuration from an io.Reader object
func (this *Config) Reader(fh io.Reader) error {
	this.Lock()
	defer this.Unlock()

	dec := json.NewDecoder(fh)
	config := new(Config)
	if err := dec.Decode(&config); err != nil {
		return err
	}

	// Copy over the config
	this.Services = config.Services

	// Clear modified flag
	this.modified = false

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// REGISTER & REMOVE SERVICES

func (this *Config) Register(service *rpc.ServiceRecord) error {
	if service == nil || service.TTL() == 0 || service.Key() == "" {
		return gopi.ErrBadParameter
	}
	if index := this.IndexForService(service); index == -1 {
		this.Services = append(this.Services, service)
		this.modified = true
	} else {
		this.Services[index] = service
		this.modified = true
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
	} else {
		// Remove the record
		this.Services = append(this.Services[:index], this.Services[index+1:]...)
		this.modified = true
	}

	// Success
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
