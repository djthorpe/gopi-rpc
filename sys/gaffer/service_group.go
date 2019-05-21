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
	"strconv"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	// Name is unique name for the service
	Name_ string `json:"name"`

	// Path is the path to the executable
	Path_ string `json:"path"`

	// Groups is a list of groups this service belongs to
	Groups_ []string `json:"groups"`

	// Flags for the command line
	Flags_ rpc.Tuples `json:"flags"`

	// Env for the command line (not used)
	Env_ rpc.Tuples `json:"-"`

	// Mode is whether the instances are started automatically
	Mode_ rpc.GafferServiceMode `json:"mode"`

	// InstanceCount determines the maximum number of running instances
	InstanceCount_ uint `json:"instance_count"`

	// RunTime determines the time the instance should run for before
	// being terminated, or zero otherwise
	RunTime_ time.Duration `json:"run_time"`

	// IdleTime determines the time the instance should be stopped before
	// it can be restarted, when in auto mode, or zero otherwise
	IdleTime_ time.Duration `json:"idle_time"`
}

type ServiceGroup struct {
	// Name is unique name for the service group
	Name_ string `json:"name"`

	// Flags for the command line
	Flags_ rpc.Tuples `json:"flags"`

	// Environment parameters for the instance
	Env_ rpc.Tuples `json:"env"`
}

type ServiceInstance struct {
	// Id is unique identifier for the service
	Id_ uint32 `json:"id"`

	// Path to executable
	Path_ string `json:"path"`

	// Service
	Service_ *Service `json:"service"`

	// Flags for the command line
	Flags_ rpc.Tuples `json:"flags"`

	// Environment parameters for the instance
	Env_ rpc.Tuples `json:"env"`

	// Start timestamp
	Start_ time.Time `json:"start_ts"`

	// Stop timestamp
	Stop_ time.Time `json:"stop_ts"`

	// Private members
	process *Process
	stdout  chan []byte
	stderr  chan []byte
	stop    chan error
}

////////////////////////////////////////////////////////////////////////////////
// SERVICE IMPLEMENTATION

func NewService(name, executable string) *Service {
	// Check parameters
	if name == "" || executable == "" {
		return nil
	}

	// Set members mostly as defaults
	this := new(Service)
	this.Name_ = name
	this.Path_ = executable
	this.Groups_ = make([]string, 0)
	this.Mode_ = rpc.GAFFER_MODE_MANUAL
	this.InstanceCount_ = 1
	this.RunTime_ = 0
	this.IdleTime_ = 0

	// Return success
	return this
}

func CopyService(service *Service) *Service {
	this := NewService(service.Name_, service.Path_)
	if service.Mode_ != rpc.GAFFER_MODE_NONE {
		this.Mode_ = service.Mode_
	}
	this.Groups_ = make([]string, len(service.Groups_))
	for i, group := range service.Groups_ {
		this.Groups_[i] = group
	}
	this.Flags_ = service.Flags_.Copy()
	this.Env_ = service.Env_.Copy()
	this.InstanceCount_ = service.InstanceCount_
	this.RunTime_ = service.RunTime_
	this.IdleTime_ = service.IdleTime_
	return this
}

func (this *Service) Name() string {
	return this.Name_
}

func (this *Service) Path() string {
	return this.Path_
}

func (this *Service) Groups() []string {
	return this.Groups_
}

func (this *Service) Mode() rpc.GafferServiceMode {
	return this.Mode_
}

func (this *Service) RunTime() time.Duration {
	return this.RunTime_
}

func (this *Service) IdleTime() time.Duration {
	return this.IdleTime_
}

func (this *Service) InstanceCount() uint {
	return this.InstanceCount_
}

func (this *Service) IsMemberOfGroup(group string) bool {
	for _, group_ := range this.Groups_ {
		if group_ == group {
			return true
		}
	}
	return false
}

func (this *Service) Flags() rpc.Tuples {
	return this.Flags_
}

func (this *Service) String() string {
	return fmt.Sprintf("<gaffer.Service>{ name=%v groups=%v flags=%v mode=%v path=%v run_time=%v idle_time=%v instance_count=%v  }", strconv.Quote(this.Name_), this.Groups(), this.Flags(), this.Mode_, strconv.Quote(this.Path_), this.RunTime_, this.IdleTime_, this.InstanceCount_)
}

////////////////////////////////////////////////////////////////////////////////
// GROUP IMPLEMENTATION

func NewGroup(name string) *ServiceGroup {
	// Check parameters
	if name == "" {
		return nil
	}

	// Set members
	this := new(ServiceGroup)
	this.Name_ = name

	// Return success
	return this
}

func CopyGroup(group *ServiceGroup) *ServiceGroup {
	if this := NewGroup(group.Name_); this == nil {
		return nil
	} else {
		this.Flags_ = group.Flags_.Copy()
		this.Env_ = group.Env_.Copy()
		return this
	}
}

func (this *ServiceGroup) Name() string {
	return this.Name_
}

func (this *ServiceGroup) Flags() rpc.Tuples {
	return this.Flags_
}

func (this *ServiceGroup) Env() rpc.Tuples {
	return this.Env_
}

func (this *ServiceGroup) String() string {
	return fmt.Sprintf("<gaffer.ServiceGroup>{ name=%v flags=%v env=%v }", strconv.Quote(this.Name_), this.Flags_, this.Env_)
}

////////////////////////////////////////////////////////////////////////////////
// INSTANCE IMPLEMENTATION

func NewInstance(id uint32, service *Service, groups []*ServiceGroup, path string, expander func(string) string) (*ServiceInstance, error) {
	// Check parameters
	if id == 0 || service == nil || groups == nil {
		return nil, gopi.ErrBadParameter
	}

	// Create the instance
	this := new(ServiceInstance)
	this.Service_ = service
	this.Path_ = path
	this.Id_ = id
	this.Flags_ = service.Flags_.Copy()
	this.Env_ = service.Env_.Copy()

	// Generate the environment & flags from groups, in order from left to right
	for _, group := range groups {
		for _, key := range group.Flags_.Keys() {
			if exists := this.Flags_.ExistsForKey(key); exists == false {
				if err := this.Flags_.SetStringForKey(key, group.Flags_.StringForKey(key)); err != nil {
					return nil, err
				}
			}
		}
		for _, key := range group.Env_.Keys() {
			if exists := this.Env_.ExistsForKey(key); exists == false {
				if err := this.Env_.SetStringForKey(key, group.Env_.StringForKey(key)); err != nil {
					return nil, err
				}
			}
		}
	}

	// Resolve environment parameters
	for _, key := range this.Env_.Keys() {
		value := this.Env_.StringForKey(key)
		this.Env_.SetStringForKey(key, os.Expand(value, func(key_ string) string {
			if this.Env_.ExistsForKey(key_) {
				return this.Env_.StringForKey(key_)
			} else if key_ == "$" {
				return key_
			} else if expander != nil {
				return expander(key_)
			} else {
				return "${" + key_ + "}"
			}
		}))
	}

	// Resolve flag parameters
	for _, key := range this.Flags_.Keys() {
		value := this.Flags_.StringForKey(key)
		this.Flags_.SetStringForKey(key, os.Expand(value, func(key_ string) string {
			if this.Env_.ExistsForKey(key_) {
				return this.Env_.StringForKey(key_)
			} else if key_ == "$" {
				return key_
			} else if expander != nil {
				return expander(key_)
			} else {
				// Not found, replace with ${key}
				return "${" + key_ + "}"
			}
		}))
	}

	// Make the process
	if process, err := NewProcess(this); err != nil {
		return nil, err
	} else {
		this.process = process
	}

	// Make the channels and start receiving data on them
	this.stdout, this.stderr = make(chan []byte), make(chan []byte)
	this.stop = make(chan error)

	// Success
	return this, nil
}

func (this *ServiceInstance) Id() uint32 {
	return this.Id_
}

func (this *ServiceInstance) Service() rpc.GafferService {
	return this.Service_
}

func (this *ServiceInstance) Path() string {
	return this.Path_
}

func (this *ServiceInstance) Flags() rpc.Tuples {
	return this.Flags_
}

func (this *ServiceInstance) Env() rpc.Tuples {
	return this.Env_
}

func (this *ServiceInstance) RunTime() time.Duration {
	return this.Service_.RunTime()
}

func (this *ServiceInstance) IdleTime() time.Duration {
	return this.Service_.IdleTime()
}

func (this *ServiceInstance) Start() time.Time {
	return this.Start_
}

func (this *ServiceInstance) Stop() time.Time {
	return this.Stop_
}

func (this *ServiceInstance) ExitCode() int64 {
	if this.process == nil {
		return 0
	} else {
		return this.process.ExitCode()
	}
}

func (this *ServiceInstance) IsRunning() bool {
	if this.process == nil {
		return false
	} else {
		return this.process.IsRunning()
	}
}

func (this *ServiceInstance) String() string {
	return fmt.Sprintf("<gaffer.ServiceInstance>{ id=%v service=%v flags=%v env=%v exit_code=%v %v }", this.Id_, strconv.Quote(this.Service_.Name()), this.Flags(), this.Env(), this.ExitCode(), this.process)
}
