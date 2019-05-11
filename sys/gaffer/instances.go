/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Instances struct {
	sync.Mutex

	// Private Members
	log           gopi.Logger
	max_instances uint32
	delta_cleanup time.Duration
	instances     map[uint32]*ServiceInstance
	ids           map[uint32]time.Time
	r             *rand.Rand
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// MAX_INSTANCES is maximum number of instances that can run at once
	MAX_INSTANCES = 0xFFFF

	// DELTA_CLEANUP is time an ID is considered to be stale before re-used
	DELTA_CLEANUP = 60 * time.Second
)

////////////////////////////////////////////////////////////////////////////////
// INIT / DESTROY

func (this *Instances) Init(config Gaffer, logger gopi.Logger) error {
	logger.Debug("<gaffer.instances.Init>{ config=%+v }", config)

	this.log = logger
	this.instances = make(map[uint32]*ServiceInstance)
	this.r = rand.New(rand.NewSource(time.Now().Unix()))
	this.ids = make(map[uint32]time.Time)

	if config.MaxInstances == 0 {
		this.max_instances = MAX_INSTANCES
	} else {
		this.max_instances = config.MaxInstances
	}

	if config.DeltaCleanup == 0 {
		this.delta_cleanup = DELTA_CLEANUP
	} else {
		this.delta_cleanup = config.DeltaCleanup
	}

	// Success
	return nil
}

func (this *Instances) Destroy() error {
	this.log.Debug("<gaffer.instances.Destroy>{ instances=%v }", this.GetInstances())

	// TODO: Stop instances

	// Release resources
	this.ids = nil
	this.instances = nil

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// CREATE INSTANCE WITH ID

func (this *Instances) NewInstance(id uint32, service *Service, groups []*ServiceGroup) (*ServiceInstance, error) {
	this.log.Debug2("<gaffer.instances.NewInstance>{ id=%v service=%v groups=%v }", id, service, groups)
	// Check incoming parameters
	if id == 0 || service == nil {
		return nil, gopi.ErrBadParameter
	}
	// Check service is not at instance_count == 0
	if service.InstanceCount_ == 0 {
		return nil, fmt.Errorf("Service %v is disabled", strconv.Quote(service.Name_))
	}
	// Check id is unused but in the ids table
	if this.IsUnusedIdentifier(id) == false {
		this.log.Debug2("IsUnusedIdentifier(%v) == false", id)
		return nil, gopi.ErrBadParameter
	}

	// Avoid race conditions
	this.Lock()
	defer this.Unlock()

	// Create instance
	if instance, err := NewInstance(id, service, groups); err != nil {
		return nil, err
	} else if instance == nil {
		return nil, gopi.ErrAppError
	} else {
		this.instances[id] = instance
		delete(this.ids, id)
		return instance, nil
	}
}

func (this *Instances) DeleteInstance(instance *ServiceInstance) error {
	this.log.Debug2("<gaffer.instances.DeleteInstance>{ instance=%v }", instance)
	// Check incoming parameters
	if instance == nil {
		return gopi.ErrBadParameter
	}

	// Avoid race conditions
	this.Lock()
	defer this.Unlock()

	// Do delete
	if instance_, exists := this.instances[instance.Id_]; exists == false {
		return gopi.ErrNotFound
	} else if instance_ != instance {
		return gopi.ErrAppError
	} else {
		delete(this.instances, instance.Id_)
		return nil
	}
}

func (this *Instances) Start(instance *ServiceInstance, stdout, stderr chan<- []byte) error {
	this.log.Debug2("<gaffer.instances.Start>{ instance=%v }", instance)

	// Check parameters
	if instance == nil || stdout == nil || stderr == nil {
		return gopi.ErrBadParameter
	}

	return gopi.ErrNotFound
}

////////////////////////////////////////////////////////////////////////////////
// RETURN INSTANCES

func (this *Instances) GetInstances() []rpc.GafferServiceInstance {
	instances := make([]rpc.GafferServiceInstance, 0, len(this.instances))
	for _, instance := range this.instances {
		instances = append(instances, instance)
	}
	return instances
}

////////////////////////////////////////////////////////////////////////////////
// GENERATE IDENTIFIERS

// Returns an unused random identifier, or zero if an identifier
// could not be found
func (this *Instances) getUnusedIdentifier(rand bool) uint32 {
	// Avoid race conditions
	this.Lock()
	defer this.Unlock()

	// Check
	check := func(r uint32) bool {
		if r <= 0 || r > this.max_instances {
			return false
		} else if _, exists := this.instances[r]; exists {
			return false
		} else if _, exists := this.ids[r]; exists {
			return false
		} else {
			return true
		}
	}

	if rand {
		// Create a new identifier, try 20 times
		for i := 0; i < 20; i++ {
			r := uint32(this.r.Int31n(int32(this.max_instances + 1)))
			if check(r) {
				this.ids[r] = time.Now()
				return r
			}
		}
	} else {
		// Create a new identifier, incrementally
		for r := uint32(1); r <= this.max_instances; r++ {
			if check(r) {
				this.ids[r] = time.Now()
				return r
			}
		}
	}

	// Return zero - to show failed
	return 0
}

// Cleanup stale identifiers
func (this *Instances) CleanupIdentifiers() {
	// Avoid race conditions
	this.Lock()
	defer this.Unlock()

	for id, t := range this.ids {
		if _, exists := this.instances[id]; exists {
			continue
		} else if time.Now().Sub(t) >= this.delta_cleanup {
			this.log.Debug("Cleanup stale identifier %v", id)
			delete(this.ids, id)
		}
	}
}

func (this *Instances) GetUnusedIdentifier() uint32 {
	// Get an unused identifier, trying a second time (after cleanup)
	// when there is a clash
	if id := this.getUnusedIdentifier(true); id > 0 {
		return id
	}

	this.CleanupIdentifiers()
	if id := this.getUnusedIdentifier(true); id > 0 {
		return id
	} else {
		return this.getUnusedIdentifier(false)
	}
}

func (this *Instances) IsUnusedIdentifier(id uint32) bool {
	// Avoid race conditions
	this.Lock()
	defer this.Unlock()

	if id == 0 {
		return false
	} else if _, exists := this.instances[id]; exists == true {
		return false
	} else if _, exists := this.ids[id]; exists == false {
		return false
	} else {
		return true
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Instances) String() string {
	return fmt.Sprintf("<instances>{ instances=%v }", this.GetInstances())
}
