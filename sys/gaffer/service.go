/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"fmt"
	"strconv"

	// Frameworks
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

	// Mode is whether the instances are started automatically
	Mode_ rpc.GafferServiceMode `json:"mode"`

	// Instances determines the maximum number of running instances
	Instances_ uint `json:"instances"`
}

type ServiceGroup struct {
	// Name is unique name for the service group
	Name_ string `json:"name"`
}

type ServiceInstance struct {
	// Id is unique identifier for the service
	Id_ uint `json:"name"`
}

/*
	// Timeout is the length of time a service should run for
	// before cancelling
	Timeout time.Duration `json:"timeout"`

	// Flags on the command line
	Flags []*Tuple `json:"flags"`

	// Args on the command line
	Args []string `json:"args"`
}

// Tuple defines a key-value pair for flags or environment vars
type Tuple struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
*/

////////////////////////////////////////////////////////////////////////////////
// SERVICE IMPLEMENTATION

func NewService(name, executable string) *Service {
	this := new(Service)
	this.Name_ = name
	this.Path_ = executable
	this.Groups_ = make([]string, 0)
	this.Mode_ = rpc.GAFFER_MODE_MANUAL
	this.Instances_ = 1
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

func (this *Service) Instances() uint {
	return this.Instances_
}

func (this *Service) IsMemberOfGroup(group string) bool {
	for _, group_ := range this.Groups_ {
		if group_ == group {
			return true
		}
	}
	return false
}

func (this *Service) String() string {
	return fmt.Sprintf("<gaffer.Service>{ name=%v }", strconv.Quote(this.Name_))
}

////////////////////////////////////////////////////////////////////////////////
// GROUP IMPLEMENTATION

func NewGroup(name string) *ServiceGroup {
	this := new(ServiceGroup)
	this.Name_ = name
	return this
}

func (this *ServiceGroup) Name() string {
	return this.Name_
}

func (this *ServiceGroup) String() string {
	return fmt.Sprintf("<gaffer.ServiceGroup>{ name=%v }", strconv.Quote(this.Name_))
}