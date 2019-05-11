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
	"regexp"
	"strconv"
	"strings"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	event "github.com/djthorpe/gopi/util/event"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Gaffer struct {
	// Config configuration
	Path        string
	BinRoot     string
	BinOverride bool

	// Instances configuration
	MaxInstances uint32
	DeltaCleanup time.Duration
}

type gaffer struct {
	log gopi.Logger

	config
	Instances
	event.Publisher
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

var (
	reServiceGroupName = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9\\-\\_\\.]*$")
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open a new gaffer instance
func (config Gaffer) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<gaffer.Open>{ path=%v binroot=%v }", strconv.Quote(config.Path), strconv.Quote(config.DefaultBin()))

	this := new(gaffer)
	this.log = logger

	if err := this.config.Init(config, logger); err != nil {
		logger.Debug2("Config.Init returned nil")
		return nil, err
	}
	if err := this.Instances.Init(config, logger); err != nil {
		logger.Debug2("Instances.Init returned nil")
		return nil, err
	}

	// Success
	return this, nil
}

// Close an opened gaffer instance
func (this *gaffer) Close() error {
	this.log.Debug("<gaffer.Close>{ }")

	// Unsubscribe subscribers
	this.Publisher.Close()

	// Release resources, etc
	if err := this.Instances.Destroy(); err != nil {
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

// Return gaffer instance information
func (this *gaffer) String() string {
	return fmt.Sprintf("<gaffer>{ %v %v }", this.config.String(), this.Instances.String())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// GetExecutables returns a list of executables which can be transformed into services,
// the list will be relative to the current binary root path determined either
// by the configuration file or overrideen by the configuration parameter
func (this *gaffer) GetExecutables(recursive bool) []string {
	this.log.Debug2("<gaffer>GetExecutables{ recursive=%v }", recursive)

	executables := make([]string, 0, 100)

	// Discover executables under a path
	if root, err := this.config.Root(); err != nil {
		this.log.Error("Executables: %v", err)
		return nil
	} else if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if path == root {
			return nil
		} else if info.IsDir() && recursive == false {
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

// AddServiceForPath returns a new default service based on the executable
// or returns an error if the executable was not found or invalid. On success,
// the service is added to the configuration
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
		} else if reServiceGroupName.MatchString(name) == false {
			this.log.Warn("AddServiceForPath: %v is not a valid service name", strconv.Quote(name))
			return nil, gopi.ErrBadParameter
		} else if service := NewService(name, executable); service == nil {
			return nil, gopi.ErrBadParameter
		} else if err := this.config.AddService(service); err != nil {
			return nil, err
		} else {
			this.EmitService(rpc.GAFFER_EVENT_SERVICE_ADD, service)
			return service, nil
		}
	}
}

// AddGroupForName returns a new empty group file, or returns an error if the
// name of the group already exists or the name was invalid
func (this *gaffer) AddGroupForName(name string) (rpc.GafferServiceGroup, error) {
	this.log.Debug2("<gaffer>AddGroupForName{ name=%v }", strconv.Quote(name))

	// Check incoming parameters
	if name == "" {
		return nil, gopi.ErrBadParameter
	}
	if reServiceGroupName.MatchString(name) == false {
		this.log.Warn("AddGroupForName: %v is not a valid service group name", strconv.Quote(name))
		return nil, gopi.ErrBadParameter
	}

	// Create a new group
	if group := NewGroup(name); group == nil {
		return nil, gopi.ErrBadParameter
	} else if err := this.config.AddGroup(group); err != nil {
		return nil, err
	} else {
		this.EmitGroup(rpc.GAFFER_EVENT_GROUP_ADD, group)
		return group, nil
	}
}

// GetServiceForName returns a service based on a name
func (this *gaffer) GetServiceForName(service string) rpc.GafferService {
	this.log.Debug2("<gaffer>GetServiceForName{ service=%v }", strconv.Quote(service))
	if service_ := this.config.GetServiceByName(service); service_ == nil {
		return nil
	} else {
		return service_
	}
}

// GetGroupsForNames returns an array of groups based on name of the group
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
	if group == "" {
		return gopi.ErrBadParameter
	} else if groups := this.config.GetGroupsByName([]string{group}); len(groups) != 1 {
		return gopi.ErrNotFound
	} else if services := this.config.ServicesForGroupByName(group); len(services) != 0 {
		services_ := make([]string, len(services))
		for i, service := range services {
			services_[i] = strconv.Quote(service.Name())
		}
		return fmt.Errorf("Group %v is in use by services %v", strconv.Quote(group), strings.Join(services_, ","))
	} else if group_ := groups[0]; group_ == nil {
		return gopi.ErrNotFound
	} else if err := this.config.RemoveGroup(groups[0]); err != nil {
		return err
	} else {
		this.EmitGroup(rpc.GAFFER_EVENT_GROUP_REMOVE, group_)
		return nil
	}
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
		this.EmitService(rpc.GAFFER_EVENT_SERVICE_REMOVE, service_)
		return nil
	}
}

// GetServices returns all services
func (this *gaffer) GetServices() []rpc.GafferService {
	services := make([]rpc.GafferService, len(this.config.Services))
	for i, service := range this.config.Services {
		services[i] = service
	}
	return services
}

// GetGroups returns all groups
func (this *gaffer) GetGroups() []rpc.GafferServiceGroup {
	groups := make([]rpc.GafferServiceGroup, len(this.config.ServiceGroups))
	for i, group := range this.config.ServiceGroups {
		groups[i] = group
	}
	return groups
}

////////////////////////////////////////////////////////////////////////////////
// EDIT METHODS

func (this *gaffer) SetServiceNameForName(service string, new string) error {
	this.log.Debug2("<gaffer>SetServiceNameForName{ service=%v new=%v }", strconv.Quote(service), strconv.Quote(new))
	if service == "" || new == "" {
		return gopi.ErrBadParameter
	} else if service == new {
		return gopi.ErrNotModified
	} else if reServiceGroupName.MatchString(new) == false {
		this.log.Warn("SetServiceNameForName: %v is not a valid service name", strconv.Quote(new))
		return gopi.ErrBadParameter
	} else if service_ := this.config.GetServiceByName(service); service_ == nil {
		return gopi.ErrNotFound
	} else if new_ := this.config.GetServiceByName(new); new_ != nil {
		return fmt.Errorf("SetServiceNameForName: %v Exists", strconv.Quote(new))
	} else {
		return gopi.ErrNotImplemented
	}
}

func (this *gaffer) SetGroupNameForName(group string, new string) error {
	return gopi.ErrNotImplemented
}

func (this *gaffer) SetServiceModeForName(service string, mode rpc.GafferServiceMode) error {
	return gopi.ErrNotImplemented
}

func (this *gaffer) SetServiceInstanceCountForName(service string, count uint) error {
	return gopi.ErrNotImplemented
}

func (this *gaffer) AddServiceGroupForName(service string, group string, position uint) error {
	return gopi.ErrNotImplemented
}

func (this *gaffer) RemoveServiceGroupForName(service string, group string) error {
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// INSTANCES

func (this *gaffer) GenerateInstanceId() uint32 {
	this.log.Debug2("<gaffer>GenerateInstanceId{}")
	return this.Instances.GetUnusedIdentifier()
}

func (this *gaffer) StartInstanceForServiceName(service string, id uint32) (rpc.GafferServiceInstance, error) {
	this.log.Debug2("<gaffer>StartInstanceForServiceName{ service=%v id=%v }", strconv.Quote(service), id)
	if service == "" || id == 0 {
		return nil, gopi.ErrBadParameter
	} else if service_ := this.config.GetServiceByName(service); service_ == nil {
		return nil, gopi.ErrNotFound
	} else if groups := this.config.GetGroupsByName(service_.Groups_); groups == nil {
		return nil, gopi.ErrBadParameter
	} else if instance, err := this.Instances.NewInstance(id, service_, groups); err != nil {
		return nil, err
	} else {
		return instance, nil
	}
}

func (this *gaffer) StopInstanceForId(id uint32) error {
	return gopi.ErrNotImplemented
}

////////////////////////////////////////////////////////////////////////////////
// EMIT

func (this *gaffer) EmitService(t rpc.GafferEventType, service rpc.GafferService) {
	this.Emit(NewEventWithService(this, t, service))
}

func (this *gaffer) EmitGroup(t rpc.GafferEventType, group rpc.GafferServiceGroup) {
	this.Emit(NewEventWithGroup(this, t, group))
}

func (this *gaffer) EmitInstance(t rpc.GafferEventType, instance rpc.GafferServiceInstance) {
	this.Emit(NewEventWithInstance(this, t, instance))
}
