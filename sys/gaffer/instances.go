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
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	flags         *gopi.Flags
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
	this.flags = config.AppFlags

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
	this.log.Debug("<gaffer.instances.Destroy>{ instances=%v }", this.GetInstances(rpc.GAFFER_INSTANCE_ANY))

	// TODO: Stop instances

	// Release resources
	this.ids = nil
	this.instances = nil

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// CREATE INSTANCE WITH ID

func (this *Instances) NewInstance(id uint32, service *Service, groups []*ServiceGroup, root string) (*ServiceInstance, error) {
	this.log.Debug2("<gaffer.instances.NewInstance>{ id=%v service=%v groups=%v root=%v }", id, service, groups, strconv.Quote(root))
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

	// Obtain path to the executable
	path := filepath.Join(root, service.Path())
	if stat, err := os.Stat(path); os.IsNotExist(err) {
		return nil, gopi.ErrNotFound
	} else if stat.Mode().IsRegular() == false {
		return nil, fmt.Errorf("Not a regular file: %v", service.Path())
	} else if isExecutableFileAtPath(path) != nil {
		return nil, fmt.Errorf("Not an executable file: %v", service.Path())
	}

	// Create instance
	if instance, err := NewInstance(id, service, groups, path, this.TupleExpander); err != nil {
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

func (this *Instances) Start(instance *ServiceInstance, ch chan<- rpc.GafferEvent) error {
	this.log.Debug2("<gaffer.instances.Start>{ instance=%v }", instance)
	this.Lock()
	defer this.Unlock()

	// Check parameters
	if instance == nil {
		return gopi.ErrBadParameter
	}

	if err := instance.process.Start(instance.stdout, instance.stderr, instance.stop); err != nil {
		return err
	}

	if instance.process.cmd != nil {
		this.log.Debug("%v %v", instance.process.cmd.Path, strings.Join(instance.Flags().Flags(), " "))
	}

	// Start goroutines for receiving data from stdout and stderr
	go this.processLog(instance, instance.stdout, rpc.GAFFER_EVENT_LOG_STDOUT, ch)
	go this.processLog(instance, instance.stderr, rpc.GAFFER_EVENT_LOG_STDERR, ch)
	go this.processStop(instance, instance.stop, ch)

	if ch != nil {
		// Send start signal
		ch <- NewEventWithInstance(nil, rpc.GAFFER_EVENT_INSTANCE_RUN, instance)
	}

	// Set start
	instance.Start_ = time.Now()

	// Return success
	return nil
}

func (this *Instances) Stop(instance *ServiceInstance) error {
	this.log.Debug2("<gaffer.instances.Stop>{ instance=%v }", instance)
	this.Lock()
	defer this.Unlock()

	// Check parameters
	if instance == nil {
		return gopi.ErrBadParameter
	}

	// Stop the process
	if err := instance.process.Stop(); err != nil {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// RETURN INSTANCES

func (this *Instances) GetInstances(flags rpc.GafferInstanceStatus) []rpc.GafferServiceInstance {
	this.Lock()
	defer this.Unlock()

	instances := make([]rpc.GafferServiceInstance, 0, len(this.instances))
	for _, instance := range this.instances {
		if flags == rpc.GAFFER_INSTANCE_ANY {
			instances = append(instances, instance)
		} else if flags&instance.Status() != 0 {
			instances = append(instances, instance)
		}
	}
	return instances
}

func (this *Instances) GetInstancesForServiceName(service string) []rpc.GafferServiceInstance {
	this.Lock()
	defer this.Unlock()

	instances := make([]rpc.GafferServiceInstance, 0, len(this.instances))
	for _, instance := range this.instances {
		if instance.Service_.Name_ == service {
			instances = append(instances, instance)
		}
	}
	return instances
}

func (this *Instances) GetInstanceForId(id uint32) *ServiceInstance {
	this.Lock()
	defer this.Unlock()

	if instance, exists := this.instances[id]; exists == false {
		return nil
	} else {
		return instance
	}
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
// PROCESS LOGS AND STOP SIGNAL

func (this *Instances) processLog(instance *ServiceInstance, in <-chan []byte, t rpc.GafferEventType, out chan<- rpc.GafferEvent) {
	for {
		if buf := <-in; buf == nil {
			break
		} else {
			out <- NewEventWithInstanceData(nil, t, instance, buf)
		}
	}
}

func (this *Instances) processStop(instance *ServiceInstance, in <-chan error, out chan<- rpc.GafferEvent) {
	for {
		if err := <-in; err == nil {
			break
		} else if err == ErrSuccess {
			// Set stop
			instance.Stop_ = time.Now()
			// Emit stop event
			out <- NewEventWithInstance(nil, rpc.GAFFER_EVENT_INSTANCE_STOP_OK, instance)
		} else {
			// Set stop
			instance.Stop_ = time.Now()
			// Emit stop event
			out <- NewEventWithInstanceData(nil, rpc.GAFFER_EVENT_INSTANCE_STOP_ERROR, instance, []byte(err.Error()))
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Instances) String() string {
	return fmt.Sprintf("<instances>{ instances=%v }", this.GetInstances(rpc.GAFFER_INSTANCE_ANY))
}

////////////////////////////////////////////////////////////////////////////////
// RESOLVER

func (this *Instances) TupleExpander(key string) string {
	this.log.Debug("<instances>TupleExpander{ key=%v flags=%v }", strconv.Quote(key))
	switch key {
	case "rpc.port":
		// Return an unused port
		if port, err := getUnusedPort(); err != nil {
			this.log.Warn("getUnusedPort: %v", err)
			return "0"
		} else {
			return fmt.Sprint(port)
		}
	case "rpc.sslcert", "rpc.sslkey":
		// Return flag argument
		if this.flags != nil {
			if value, exists := this.flags.GetString(key); exists {
				return value
			}
		}
		fallthrough
	default:
		return "${" + key + "}"
	}
}

func getUnusedPort() (int, error) {
	if addr, err := net.ResolveTCPAddr("tcp", "localhost:0"); err != nil {
		return 0, err
	} else if listen, err := net.ListenTCP("tcp", addr); err != nil {
		return 0, err
	} else {
		defer listen.Close()
		return listen.Addr().(*net.TCPAddr).Port, nil
	}
}
