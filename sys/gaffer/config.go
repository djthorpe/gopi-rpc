/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type config_ struct {
	// Public Members
	BinRoot       string          `json:"root"`
	Services      []*Service      `json:"services"`
	ServiceGroups []*ServiceGroup `json:"groups"`
}

type config struct {
	// Database
	config_

	// Private Members
	log      gopi.Logger
	path     string
	root     string
	modified bool

	sync.Mutex
	event.Tasks
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	FILENAME_DEFAULT = "gaffer.json"
	WRITE_DELTA      = 5 * time.Second
)

////////////////////////////////////////////////////////////////////////////////
// INIT / DESTROY

func (config Gaffer) DefaultBin() string {
	if config.BinRoot == "" {
		return os.Getenv("GOBIN")
	} else if stat, err := os.Stat(config.BinRoot); err != nil {
		return ""
	} else if stat.IsDir() == false {
		return ""
	} else {
		return config.BinRoot
	}
}

func (this *config) Init(config Gaffer, logger gopi.Logger) error {
	logger.Debug("<gaffer.config.Init>{ config=%+v }", config)

	this.log = logger
	this.Services = make([]*Service, 0)
	this.ServiceGroups = make([]*ServiceGroup, 0)

	// Read or create file
	if config.Path != "" {
		if err := this.ReadPath(config.Path); err != nil {
			return fmt.Errorf("ReadPath: %v: %v", config.Path, err)
		}
	}

	// Override the root if invalid or override is set
	if config.BinOverride || this.config_.BinRoot == "" {
		this.root = config.DefaultBin()
	} else {
		this.root = this.config_.BinRoot
	}
	if _, err := this.Root(); err != nil {
		return err
	} else if this.config_.BinRoot == "" {
		// Write back immediately if the this.config_.BinRoot is set
		this.config_.BinRoot = this.root
		this.SetModified()
	}

	// Start process to write occasionally to disk
	this.Tasks.Start(this.WriteConfigTask)

	// Success
	return nil
}

func (this *config) Destroy() error {
	this.log.Debug("<gaffer.config.Destroy>{ path=%v root=%v }", strconv.Quote(this.path), strconv.Quote(this.root))

	// Stop all tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *config) String() string {
	return fmt.Sprintf("<config>{ path=%v root=%v num_services=%v }", strconv.Quote(this.path), strconv.Quote(this.root), len(this.Services))
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *config) Root() (string, error) {
	if this.root == "" {
		return "", fmt.Errorf("Missing -gaffer.root path")
	} else if stat, err := os.Stat(this.root); os.IsNotExist(err) || stat.IsDir() == false {
		return "", fmt.Errorf("Invalid -gaffer.root path")
	} else if err := isExecutableFileAtPath(this.root); err != nil {
		return "", fmt.Errorf("Invalid permissions for -gaffer.root path")
	} else {
		return this.root, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// READ AND WRITE CONFIG

// SetModified sets the modified flag to true
func (this *config) SetModified() {
	this.Lock()
	defer this.Unlock()
	this.modified = true
}

// ReadPath creates regular file if it doesn't exist, or else reads from the path
func (this *config) ReadPath(path string) error {
	this.log.Debug2("<gaffer.config>ReadPath{ path=%v }", strconv.Quote(path))

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
		if err := this.ReadPath_(this.path); err != nil {
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
		} else {
			this.SetModified()
			return nil
		}
	} else {
		return err
	}
}

// WritePath writes the configuration file to disk
func (this *config) WritePath(path string, indent bool) error {
	this.log.Debug2("<gaffer.config>WritePath{ path=%v indent=%v }", strconv.Quote(path), indent)
	this.Lock()
	defer this.Unlock()
	if fh, err := os.Create(path); err != nil {
		return err
	} else {
		defer fh.Close()
		if err := this.Writer(fh, this.Services, indent); err != nil {
			return err
		} else {
			this.modified = false
		}
	}

	// Success
	return nil
}

func (this *config) ReadPath_(path string) error {
	this.Lock()
	defer this.Unlock()

	if fh, err := os.Open(path); err != nil {
		return err
	} else {
		defer fh.Close()
		if err := this.Reader(fh); err != nil {
			return err
		} else {
			this.modified = false
		}
	}

	// Success
	return nil
}

// Reader reads the configuration from an io.Reader object
func (this *config) Reader(fh io.Reader) error {
	dec := json.NewDecoder(fh)
	if err := dec.Decode(&this.config_); err != nil {
		return err
	} else {
		return nil
	}
}

// Writer writes an array of service records to a io.Writer object
func (this *config) Writer(fh io.Writer, records []*Service, indent bool) error {
	enc := json.NewEncoder(fh)
	if indent {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(this.config_); err != nil {
		return err
	}
	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SERVICES

func (this *config) AddService(service *Service) error {
	this.log.Debug2("<gaffer.config>AddService{ service=%v }", service)

	if service == nil {
		return gopi.ErrBadParameter
	} else if service_ := this.GetServiceByName(service.Name()); service_ != nil {
		return fmt.Errorf("Duplicate service name: %v", strconv.Quote(service.Name()))
	} else {
		this.Lock()
		defer this.Unlock()
		this.Services = append(this.Services, service)
		this.modified = true
	}

	// Success
	return nil
}

// GetServiceByName returns a service structure from name
func (this *config) GetServiceByName(service string) *Service {
	this.log.Debug2("<gaffer.config>GetServiceByName{ service=%v }", strconv.Quote(service))
	this.Lock()
	defer this.Unlock()

	if service == "" {
		return nil
	}
	for _, service_ := range this.Services {
		if service == service_.Name() {
			return service_
		}
	}
	return nil
}

// GetGroupsByName returns an array of group structures, or nil if any
// of the groups could not be found
func (this *config) GetGroupsByName(groups []string) []*ServiceGroup {
	this.log.Debug2("<gaffer.config>GetGroupsByName{ groups=%v }", groups)
	this.Lock()
	defer this.Unlock()

	if len(groups) == 0 {
		return nil
	}
	for _, service_ := range this.Services {
		if service == service_.Name() {
			return service_
		}
	}
	return nil
}

func (this *config) GenerateNameFromExecutable(executable string) (string, error) {
	this.log.Debug2("<gaffer.config>GenerateNameFromExecutable{ executable=%v }", strconv.Quote(executable))

	// Get the base name and check against
	base := filepath.Base(executable)
	service := base
	for i := 1; i <= 10; i++ {
		if this.GetServiceByName(service) == nil {
			return service, nil
		} else {
			service = fmt.Sprintf("%s-%v", base, i)
		}
	}

	// Return a not found error after 10 attempts
	return "", gopi.ErrNotFound
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *config) WriteConfigTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	ticker := time.NewTimer(100 * time.Millisecond)
FOR_LOOP:
	for {
		select {
		case <-ticker.C:
			if this.modified {
				if this.path == "" {
					// Do nothing
				} else if err := this.WritePath(this.path, true); err != nil {
					this.log.Warn("Write: %v: %v", this.path, err)
				}
			}
			ticker.Reset(WRITE_DELTA)
		case <-stop:
			break FOR_LOOP
		}
	}

	// Stop the ticker
	ticker.Stop()

	// Try and write
	if this.modified {
		if this.path == "" {
			// Do nothing
		} else if err := this.WritePath(this.path, true); err != nil {
			this.log.Warn("Write: %v: %v", this.path, err)
		}
	}

	// Success
	return nil
}
