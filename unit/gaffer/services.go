/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	// Frameworks

	"math/rand"
	"os/user"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	rpc "github.com/djthorpe/gopi-rpc/v2"
	gopi "github.com/djthorpe/gopi/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	rpc.GafferService
	Enabled bool
	flag    bool
}

type services struct {
	sync.Mutex
	Log      gopi.Logger
	services map[uint32]*Service
	user     *user.User
}

////////////////////////////////////////////////////////////////////////////////
// INIT AND CLOSE

func (this *services) Init(config Gaffer, log gopi.Logger) error {
	// Seed random number
	rand.Seed(time.Now().Unix())

	// Set parameters
	this.Log = log
	this.services = make(map[uint32]*Service)

	// Set user and group
	if user, err := user.Current(); err != nil {
		return err
	} else {
		this.user = user
	}

	// Return success
	return nil
}

func (this *services) Close() error {
	// Release resources
	this.Log = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// AND AND DISABLE SERVICES

func (this *services) List() []*Service {
	services := make([]*Service, 0, len(this.services))
	for _, service := range this.services {
		services = append(services, service)
	}
	return services
}

func (this *services) Modify(executables []string) bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Modified is set to true if new service executable is added
	// or a service executable is removed
	modified := false

	// Set flag to false on all services
	for _, s := range this.services {
		s.flag = false
	}
	// Iterate services
	for _, exec := range executables {
		if services := this.servicesWithPath(exec); len(services) > 0 {
			for _, service := range services {
				service.flag = true
			}
		} else if service := this.add(exec); service == nil {
			this.Log.Warn("Unable to add service", strconv.Quote(exec))
		} else {
			this.Log.Info("Added:", service)
			service.flag = true
			modified = true
		}
	}
	// If any flags are still false, then this service should be disabled
	for _, service := range this.services {
		if service.flag == false && service.Enabled == true {
			this.Log.Info("TODO: Disable service:", service)
			service.Enabled = false
			modified = true
		}
	}

	return modified
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *services) servicesWithPath(path string) []*Service {
	services := make([]*Service, 0)
	for _, service := range this.services {
		if service.Path == path {
			services = append(services, service)
		}
	}
	return services
}

func (this *services) add(executable string) *Service {
	// Obtain a new Session ID
	sid := this.newId()
	if sid == 0 {
		return nil
	}

	// Add a new service from an executable
	service := &Service{
		GafferService: rpc.GafferService{
			Name:  filepath.Base(executable),
			Path:  executable,
			Cwd:   this.user.HomeDir,
			User:  this.user.Uid,
			Group: this.user.Gid,
			Sid:   sid,
		},
		Enabled: false,
	}
	this.services[sid] = service

	// Return the service
	return service
}

func (this *services) newId() uint32 {
	// Try to get a unique id 25 times before failing
	// the first id's would be between 1 and 63 and the second
	// between 1 and 127 and so forth.
	mod := uint32(64)
	for i := 0; i < 25; i++ {
		rand := uint32(rand.Int31()) % mod
		if _, exists := this.services[rand]; exists == false && rand > 0 {
			return rand
		} else {
			mod <<= 1
		}
	}
	return 0
}
