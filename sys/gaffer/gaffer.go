/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Gaffer struct {
	Path        string
	BinRoot     string
	BinOverride bool
}

type gaffer struct {
	log gopi.Logger

	config
	instances
}

////////////////////////////////////////////////////////////////////////////////
// COMSTANTS

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Gaffer) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<gaffer.Open>{ path=%v binroot=%v }", strconv.Quote(config.Path), strconv.Quote(config.DefaultBin()))

	this := new(gaffer)
	this.log = logger

	if err := this.config.Init(config, logger); err != nil {
		logger.Debug2("Config.Init returned nil")
		return nil, err
	}
	if err := this.instances.Init(config, logger); err != nil {
		logger.Debug2("Instances.Init returned nil")
		return nil, err
	}

	// Success
	return this, nil
}

func (this *gaffer) Close() error {
	this.log.Debug("<gaffer.Close>{ }")

	// Release resources, etc
	if err := this.instances.Destroy(); err != nil {
		return err
	}
	if err := this.config.Destroy(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *gaffer) String() string {
	return fmt.Sprintf("<gaffer>{ %v %v }", this.config.String(), this.instances.String())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *gaffer) Executables(recursive bool) []string {
	this.log.Debug2("<gaffer>Executables{ recursive=%v }", recursive)

	executables := make([]string, 0, 100)

	// Discover executables under a path
	if root, err := this.config.Root(); err != nil {
		this.log.Error("Executables: %v", err)
		return nil
	} else if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && recursive == false {
			return filepath.SkipDir
		}
		if info.Mode().IsRegular() && isExecutableFileAtPath(path) == nil {
			// Trim prefix
			sep := string(filepath.Separator)
			path := strings.TrimPrefix(strings.TrimPrefix(path, root), sep)
			// Append
			executables = append(executables, path)
		}
		return nil
	}); err != nil {
		this.log.Error("Executables: %v", err)
		return nil
	}

	return executables
}

// Return a new service
func (this *gaffer) AddServiceForPath(executable string) (rpc.GafferService, error) {
	this.log.Debug2("<gaffer>AddServiceForPath{ executable=%v }", strconv.Quote(executable))

	// Check incoming parameters
	if executable == "" {
		return nil, gopi.ErrBadParameter
	}
	if root, err := this.config.Root(); err != nil {
		return nil, gopi.ErrBadParameter
	} else if filepath.IsAbs(executable) {
		return nil, gopi.ErrBadParameter
	} else {
		// Locate executable
		executable = filepath.Clean(executable)
		executable_ := filepath.Join(root, executable)
		if stat, err := os.Stat(executable_); os.IsNotExist(err) {
			return nil, gopi.ErrNotFound
		} else if stat.Mode().IsRegular() == false {
			return nil, gopi.ErrBadParameter
		} else if err := isExecutableFileAtPath(executable_); err != nil {
			return nil, fmt.Errorf("No executable permissions: %v", executable_)
		}

		// Find a name which doesn't clash
		if name, err := this.config.GenerateNameFromExecutable(executable); err != nil {
			return nil, err
		} else if service := NewService(name, executable); service == nil {
			return nil, gopi.ErrBadParameter
		} else if err := this.config.AddService(service); err != nil {
			return nil, err
		} else {
			return service, nil
		}
	}
}

// Return a new group
func (this *gaffer) AddGroupForName(name string) (rpc.GafferServiceGroup, error) {
	this.log.Debug2("<gaffer>AddGroupForName{ name=%v }", strconv.Quote(name))

	// Check incoming parameters
	if name == "" {
		return nil, gopi.ErrBadParameter
	}
	// Create a new group
	if group := NewGroup(name); group == nil {
		return nil, gopi.ErrBadParameter
	} else if err := this.config.AddGroup(group); err != nil {
		return nil, err
	} else {
		return group, nil
	}
}

// Return an existing service
func (this *gaffer) GetServiceForName(service string) rpc.GafferService {
	this.log.Debug2("<gaffer>GetServiceForName{ service=%v }", strconv.Quote(service))
	return this.config.GetServiceByName(service)
}

// Return an array of service groups or nil if any name could not be found
func (this *gaffer) GetGroupsForNames(groups []string) []rpc.GafferServiceGroup {
	this.log.Debug2("<gaffer>GetGroupsForNames{ groups=%v }", groups)
	if groups_ := this.config.GetGroupsByName(groups); groups_ == nil {
		return nil
	} else {
		groups__ := make([]rpc.GafferServiceGroup, len(groups_))
		for i, group := range groups_ {
			groups__[i] = group
		}
		return groups__
	}
}

// Remove a group
func (this *gaffer) RemoveGroupForName(group string) error {
	this.log.Debug2("<gaffer>RemoveGroupForName{ group=%v }", strconv.Quote(group))
	return gopi.ErrNotImplemented

}

// Remove a service
func (this *gaffer) RemoveServiceForName(service string) error {
	this.log.Debug2("<gaffer>RemoveServiceForName{ service=%v }", strconv.Quote(service))
	if service == "" {
		return gopi.ErrBadParameter
	} else if service_ := this.config.GetServiceByName(service); service_ == nil {
		return gopi.ErrNotFound
	} else if err := this.config.RemoveService(service_); err != nil {
		return err
	} else {
		return nil
	}
}

// Return all services and groups
func (this *gaffer) Services() []rpc.GafferService {
	services := make([]rpc.GafferService, len(this.config.Services))
	for i, service := range this.config.Services {
		services[i] = service
	}
	return services
}

func (this *gaffer) Groups() []rpc.GafferServiceGroup {
	groups := make([]rpc.GafferServiceGroup, len(this.config.ServiceGroups))
	for i, group := range this.config.ServiceGroups {
		groups[i] = group
	}
	return groups
}
