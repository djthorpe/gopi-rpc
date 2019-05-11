/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"fmt"
	"regexp"
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
	Flags_ *Tuples `json:"flags"`

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
	Flags_ *Tuples `json:"flags"`

	// Environment parameters for the instance
	Env_ *Tuples `json:"env"`
}

type ServiceInstance struct {
	// Id is unique identifier for the service
	Id_ uint32 `json:"id"`

	// Service
	Service_ *Service `json:"service"`

	// Flags for the command line
	Flags_ *Tuples `json:"flags"`

	// Environment parameters for the instance
	Env_ *Tuples `json:"env"`

	// Private members
	process *Process
}

type Tuples struct {
	Tuples_ map[string]string `json:"tuples"`
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

var (
	reTupleKey = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9\\-\\_\\.]*$")
)

////////////////////////////////////////////////////////////////////////////////
// SERVICE IMPLEMENTATION

func NewService(name, executable string) *Service {
	this := new(Service)
	this.Name_ = name
	this.Path_ = executable
	this.Groups_ = make([]string, 0)
	this.Flags_ = NewTuples()
	this.Mode_ = rpc.GAFFER_MODE_MANUAL
	this.InstanceCount_ = 1
	this.RunTime_ = 0
	this.IdleTime_ = 0
	return this
}

func CopyService(service *Service) *Service {
	this := NewService(service.Name_, service.Path_)
	if service.Mode_ != rpc.GAFFER_MODE_NONE {
		this.Mode_ = service.Mode_
	}
	if service.Groups_ != nil {
		copy(this.Groups_, service.Groups_)
	}
	if service.Flags_ != nil {
		this.Flags_ = service.Flags_
	}
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
	var groups []string
	copy(groups, this.Groups_)
	return groups
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

func (this *Service) Flags() []string {
	return this.Flags_.Strings()
}

func (this *Service) SetFlags(map[string]string) error {
	// TODO
	return gopi.ErrNotImplemented
}

func (this *Service) String() string {
	return fmt.Sprintf("<gaffer.Service>{ name=%v groups=%v flags=%v mode=%v path=%v run_time=%v idle_time=%v instance_count=%v  }", strconv.Quote(this.Name_), this.Groups(), this.Flags(), this.Mode_, strconv.Quote(this.Path_), this.RunTime_, this.IdleTime_, this.InstanceCount_)
}

////////////////////////////////////////////////////////////////////////////////
// GROUP IMPLEMENTATION

func NewGroup(name string) *ServiceGroup {
	this := new(ServiceGroup)
	this.Name_ = name
	this.Flags_ = NewTuples()
	this.Env_ = NewTuples()
	return this
}

func CopyGroup(group *ServiceGroup) *ServiceGroup {
	this := NewGroup(group.Name_)
	if group.Flags_ != nil {
		this.Flags_ = group.Flags_
	}
	if group.Env_ != nil {
		this.Env_ = group.Env_
	}
	return this
}

func (this *ServiceGroup) Name() string {
	return this.Name_
}

func (this *ServiceGroup) SetFlags(map[string]string) error {
	// TODO
	return gopi.ErrNotImplemented
}

func (this *ServiceGroup) SetEnv(map[string]string) error {
	// TODO
	return gopi.ErrNotImplemented
}

func (this *ServiceGroup) Flags() []string {
	return this.Flags_.Strings()
}

func (this *ServiceGroup) Env() []string {
	return this.Env_.Strings()
}

func (this *ServiceGroup) String() string {
	return fmt.Sprintf("<gaffer.ServiceGroup>{ name=%v flags=%v env=%v }", strconv.Quote(this.Name_), this.Flags_.Strings(), this.Env_.Strings())
}

////////////////////////////////////////////////////////////////////////////////
// INSTANCE IMPLEMENTATION

func NewInstance(id uint32, service *Service, groups []*ServiceGroup) (*ServiceInstance, error) {
	// Check parameters
	if id == 0 || service == nil {
		return nil, gopi.ErrBadParameter
	}

	// Create the instance
	this := new(ServiceInstance)
	this.Service_ = service
	this.Id_ = id
	this.Env_ = NewTuples()
	this.Flags_ = NewTuples()

	// Copy flags from service
	if service.Flags_ != nil {
		for k, v := range service.Flags_.Tuples_ {
			this.Flags_.Set(k, v)
		}
	}

	// Generate the environment & flags from groups
	for _, group := range groups {
		if group.Flags_ != nil {
			for k, v := range group.Flags_.Tuples_ {
				if _, exists := this.Flags_.Tuples_[k]; exists == false {
					this.Flags_.Set(k, v)
				}
			}
		}
		if group.Env_ != nil {
			for k, v := range group.Env_.Tuples_ {
				if _, exists := this.Env_.Tuples_[k]; exists == false {
					this.Env_.Set(k, v)
				}
			}
		}
	}

	// Make the process
	if process, err := NewProcess(this); err != nil {
		return nil, err
	} else {
		this.process = process
	}

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
	return this.Service_.Path()
}

func (this *ServiceInstance) Flags() []string {
	return this.Flags_.Strings()
}

func (this *ServiceInstance) Env() []string {
	return this.Env_.Strings()
}

func (this *ServiceInstance) RunTime() time.Duration {
	return this.Service_.RunTime()
}

func (this *ServiceInstance) IdleTime() time.Duration {
	return this.Service_.IdleTime()
}

func (this *ServiceInstance) String() string {
	return fmt.Sprintf("<gaffer.ServiceInstance>{ id=%v service=%v flags=%v env=%v process=%v }", this.Id_, strconv.Quote(this.Service_.Name()), this.Flags(), this.Env(), this.process)
}

////////////////////////////////////////////////////////////////////////////////
// TUPLES IMPLEMENTATION

func NewTuples() *Tuples {
	this := new(Tuples)
	this.Tuples_ = make(map[string]string)
	return this
}

func (this *Tuples) Set(key, value string) error {
	if reTupleKey.MatchString(key) == false {
		return gopi.ErrBadParameter
	} else {
		this.Tuples_[key] = value
		return nil
	}
}

func (this *Tuples) Strings() []string {
	str := make([]string, 0, len(this.Tuples_))
	for k, v := range this.Tuples_ {
		str = append(str, fmt.Sprintf("%v=%v", k, strconv.Quote(v)))
	}
	return str
}
