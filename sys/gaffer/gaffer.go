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

	// Appflags
	AppFlags *gopi.Flags
}

type gaffer struct {
	log gopi.Logger
	evt chan rpc.GafferEvent

	config
	Instances
	event.Publisher
	event.Tasks
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

var (
	// Group and service names must start with an alpha character to
	// distinguish against a number
	reServiceGroupName = regexp.MustCompile("^[A-Za-z][A-Za-z0-9\\-\\_\\.]*$")

	// Use similar pattern for executables
	reExecutableName = regexp.MustCompile("^[A-Za-z][A-Za-z0-9\\-\\_\\.\\/]*$")
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open a new gaffer instance
func (config Gaffer) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<gaffer.Open>{ path=%v binroot=%v }", strconv.Quote(config.Path), strconv.Quote(config.DefaultBin()))

	this := new(gaffer)
	this.log = logger
	this.evt = make(chan rpc.GafferEvent)

	if err := this.config.Init(config, logger); err != nil {
		logger.Debug2("Config.Init returned nil")
		return nil, err
	}
	if err := this.Instances.Init(config, logger); err != nil {
		logger.Debug2("Instances.Init returned nil")
		return nil, err
	}

	// Start background task which starts and stops instances
	this.Tasks.Start(this.InstanceTask, this.LoggingTask)

	// Success
	return this, nil
}

// Close an opened gaffer instance
func (this *gaffer) Close() error {
	this.log.Debug("<gaffer.Close>{ }")

	// Unsubscribe subscribers
	this.Publisher.Close()

	// Stop background tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Close channels
	close(this.evt)

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
			if reExecutableName.MatchString(path) {
				executables = append(executables, path)
			} else {
				this.log.Warn("Ignoring path: %v", strconv.Quote(path))
			}
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
func (this *gaffer) AddServiceForPath(executable, service string) (rpc.GafferService, error) {
	this.log.Debug2("<gaffer>AddServiceForPath{ executable=%v service=%v }", strconv.Quote(executable), strconv.Quote(service))

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
		if service == "" {
			service = executable
		}
		if name, err := this.config.GenerateNameFromExecutable(service); err != nil {
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
	} else if err := this.config.SetServiceName(service_, new); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *gaffer) SetGroupNameForName(group string, new string) error {
	this.log.Debug2("<gaffer>SetGroupNameForName{ group=%v new=%v }", strconv.Quote(group), strconv.Quote(new))
	return gopi.ErrNotImplemented
}

func (this *gaffer) SetServiceModeForName(service string, mode rpc.GafferServiceMode) error {
	this.log.Debug2("<gaffer>SetServiceModeForName{ service=%v mode=%v }", strconv.Quote(service), mode)
	if service == "" || mode == rpc.GAFFER_MODE_NONE {
		return gopi.ErrBadParameter
	} else if service_ := this.GetServiceByName(service); service_ == nil {
		return gopi.ErrNotFound
	} else if err := this.config.SetServiceMode(service_, mode); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *gaffer) SetServiceRunTimeForName(service string, run_time time.Duration) error {
	this.log.Debug2("<gaffer>SetServiceRunTimeForName{ service=%v run_time=%v }", strconv.Quote(service), run_time)
	if service == "" || run_time < 0 {
		return gopi.ErrBadParameter
	} else if service_ := this.GetServiceByName(service); service_ == nil {
		return gopi.ErrNotFound
	} else if err := this.config.SetServiceRunTime(service_, run_time); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *gaffer) SetServiceIdleTimeForName(service string, idle_time time.Duration) error {
	this.log.Debug2("<gaffer>SetServiceIdleTimeForName{ service=%v idle_time=%v }", strconv.Quote(service), idle_time)
	if service == "" || idle_time < 0 {
		return gopi.ErrBadParameter
	} else if service_ := this.GetServiceByName(service); service_ == nil {
		return gopi.ErrNotFound
	} else if err := this.config.SetServiceIdleTime(service_, idle_time); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *gaffer) SetServiceInstanceCountForName(service string, count uint) error {
	this.log.Debug2("<gaffer>SetServiceInstanceCountForName{ service=%v count=%v }", strconv.Quote(service), count)
	if service == "" {
		return gopi.ErrBadParameter
	} else if service_ := this.GetServiceByName(service); service_ == nil {
		return gopi.ErrNotFound
	} else if err := this.config.SetServiceInstanceCount(service_, count); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *gaffer) SetServiceGroupsForName(service string, groups []string) error {
	this.log.Debug2("<gaffer>SetServiceGroupsForName{ service=%v groups=%v }", strconv.Quote(service), groups)

	if service == "" {
		return gopi.ErrBadParameter
	} else if service_ := this.GetServiceByName(service); service_ == nil {
		return gopi.ErrNotFound
	} else if groups_ := this.GetGroupsForNames(groups); len(groups_) != len(groups) {
		return gopi.ErrNotFound
	} else if err := this.config.SetServiceGroups(service_, groups); err != nil {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// INSTANCES

func (this *gaffer) GenerateInstanceId() uint32 {
	this.log.Debug2("<gaffer>GenerateInstanceId{}")
	return this.Instances.GetUnusedIdentifier()
}

func (this *gaffer) GetInstanceForId(id uint32) rpc.GafferServiceInstance {
	if instance := this.Instances.GetInstanceForId(id); instance != nil {
		return instance
	} else {
		return nil
	}
}

func (this *gaffer) StartInstanceForServiceName(service string, id uint32) (rpc.GafferServiceInstance, error) {
	this.log.Debug2("<gaffer>StartInstanceForServiceName{ service=%v id=%v }", strconv.Quote(service), id)
	if service == "" || id == 0 {
		return nil, gopi.ErrBadParameter
	} else if service_ := this.config.GetServiceByName(service); service_ == nil {
		return nil, gopi.ErrNotFound
	} else if groups := this.config.GetGroupsByName(service_.Groups_); groups == nil {
		return nil, gopi.ErrBadParameter
	} else if root, err := this.Root(); err != nil {
		return nil, err
	} else if instance, err := this.Instances.NewInstance(id, service_, groups, root); err != nil {
		return nil, err
	} else {
		this.EmitInstance(rpc.GAFFER_EVENT_INSTANCE_ADD, instance)
		return instance, nil
	}
}

func (this *gaffer) StopInstanceForId(id uint32) error {
	this.log.Debug2("<gaffer>StopInstanceForId{ id=%v }", id)
	if id == 0 {
		return gopi.ErrBadParameter
	}

	if instance := this.Instances.GetInstanceForId(id); instance == nil {
		return gopi.ErrNotFound
	} else if instance.IsRunning() == false {
		this.log.Warn("StopInstanceForId: Instance %v is not running", id)
		return gopi.ErrOutOfOrder
	} else if err := this.Instances.Stop(instance); err != nil {
		return err
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// TUPLES

func (this *gaffer) SetServiceFlagsForName(service string, tuples rpc.Tuples) error {
	this.log.Debug2("<gaffer>SetServiceFlagsForName{ service=%v tuples=%v }", strconv.Quote(service), tuples)
	if service == "" {
		return gopi.ErrBadParameter
	}
	if service_ := this.config.GetServiceByName(service); service_ == nil {
		return gopi.ErrNotFound
	} else if err := this.config.SetServiceFlags(service_, tuples); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *gaffer) SetGroupFlagsForName(group string, tuples rpc.Tuples) error {
	this.log.Debug2("<gaffer>SetGroupFlagsForName{ group=%v tuples=%v }", strconv.Quote(group), tuples)
	if group == "" {
		return gopi.ErrBadParameter
	}
	if group_ := this.config.GetGroupsByName([]string{group}); len(group_) == 0 {
		return gopi.ErrNotFound
	} else if err := this.config.SetGroupFlags(group_[0], tuples); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *gaffer) SetGroupEnvForName(group string, tuples rpc.Tuples) error {
	this.log.Debug2("<gaffer>SetGroupEnvForName{ group=%v tuples=%v }", strconv.Quote(group), tuples)
	if group == "" {
		return gopi.ErrBadParameter
	}
	if group_ := this.config.GetGroupsByName([]string{group}); len(group_) == 0 {
		return gopi.ErrNotFound
	} else if err := this.config.SetGroupEnv(group_[0], tuples); err != nil {
		return err
	} else {
		return nil
	}
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

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *gaffer) InstanceTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	events := this.Subscribe()
	timer := time.NewTicker(1 * time.Second)
FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt == nil {
				// Do nothing
			} else if evt_, ok := evt.(rpc.GafferEvent); ok == false {
				this.log.Warn("InstanceTask: Unhandled event: %v", evt)
			} else if err := this.InstanceTaskHandler(evt_); err != nil {
				this.log.Error("InstanceTask: %v: %v", evt, err)
			}
		case <-timer.C:
			if err := this.InstanceStartHandler(); err != nil {
				this.log.Error("InstanceStartHandler: %v", err)
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	timer.Stop()
	this.Unsubscribe(events)

	// Success
	return nil
}

func (this *gaffer) InstanceStartHandler() error {
	for _, service := range this.config.Services {
		if service.Mode() != rpc.GAFFER_MODE_AUTO {
			continue
		} else {
			fmt.Println("TODO: Check instances for %v", service)
		}
	}

	// Success
	return nil
}

func (this *gaffer) InstanceTaskHandler(evt rpc.GafferEvent) error {
	switch evt.Type() {
	case rpc.GAFFER_EVENT_INSTANCE_ADD:
		if instance := evt.Instance(); instance == nil {
			return gopi.ErrBadParameter
		} else if instance_, ok := instance.(*ServiceInstance); ok == false {
			return gopi.ErrBadParameter
		} else if err := this.Instances.Start(instance_, this.evt); err != nil {
			return err
		}
	}
	// Return success
	return nil
}

func (this *gaffer) LoggingTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
FOR_LOOP:
	for {
		select {
		case evt := <-this.evt:
			this.Emit(evt)
		case <-stop:
			break FOR_LOOP
		}
	}

	// Success
	return nil
}
