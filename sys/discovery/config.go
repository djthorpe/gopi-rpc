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
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Config struct {
	sync.Mutex

	// Public members
	Services []gopi.RPCServiceRecord `json:"services"`

	// Private members
	modified bool
}

////////////////////////////////////////////////////////////////////////////////
// INIT / DEINIT

func (this *Config) Init() error {
	this.Services = make([]gopi.RPCServiceRecord, 0)
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
