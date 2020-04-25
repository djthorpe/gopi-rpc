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

type services struct {
	sync.Mutex
	Log     gopi.Logger
	service map[uint32]*service
	flag    map[uint32]bool
	user    *user.User
}

////////////////////////////////////////////////////////////////////////////////
// INIT AND CLOSE

func (this *services) Init(config Gaffer, log gopi.Logger) error {
	// Seed random number
	rand.Seed(time.Now().Unix())

	// Set parameters
	this.Log = log
	this.service = make(map[uint32]*service)

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
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Release resources
	this.Log = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

func (this *services) Services() []rpc.GafferService {
	services := make([]rpc.GafferService, 0, len(this.service))
	for _, service := range this.service {
		services = append(services, service)
	}
	return services
}

func (this *services) Update(src rpc.GafferService, fields []string) (rpc.GafferService, error) {
	if src == nil {
		return nil, gopi.ErrBadParameter.WithPrefix("service")
	} else if len(fields) == 0 {
		return nil, gopi.ErrBadParameter.WithPrefix("fields")
	}

	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Retrieve a service by ID
	if service, exists := this.service[src.Sid()]; exists == false {
		return nil, gopi.ErrNotFound.WithPrefix("sid")
	} else {
		// Create a copy
		dst := NewService(service, service.instances)
		modified := false
		for _, field := range fields {
			switch field {
			case "enabled":
				if src.Enabled() != dst.enabled {
					dst.enabled = src.Enabled()
					modified = true
				}
			case "name":
				if dst.IsValidName(src.Name()) == false {
					return nil, gopi.ErrBadParameter.WithPrefix("name")
				}
				if src.Name() != dst.name {
					dst.name = src.Name()
					modified = true
				}
			default:
				return nil, gopi.ErrBadParameter.WithPrefix(field)
			}
		}

		// Nothing modified
		if modified == false {
			return nil, rpc.ERROR_NOT_MODIFIED
		}

		// Commit the transaction, return success
		this.service[src.Sid()] = dst
		return dst, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// AND AND DISABLE SERVICES

func (this *services) modify(executables []string) bool {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Modified is set to true if new service executable is added
	// or a service executable is removed
	modified := false
	flag := make(map[uint32]bool, len(this.service))

	// Iterate services
	for _, exec := range executables {
		if services := this.servicesWithPath(exec); len(services) > 0 {
			for _, service := range services {
				flag[service.sid] = true
			}
		} else if service := this.add(exec); service == nil {
			this.Log.Warn("Unable to add service", strconv.Quote(exec))
		} else {
			this.Log.Info("Added:", service)
			flag[service.sid] = true
			modified = true
		}
	}
	// If any flags are still false, then this service should be disabled
	for _, service := range this.service {
		if _, exists := flag[service.sid]; exists == false {
			if service.enabled == true {
				this.Log.Info("TODO: Disable service:", service)
				service.enabled = false
				modified = true
			}
		}
	}

	return modified
}

////////////////////////////////////////////////////////////////////////////////
// SERVICE IMPLEMENTATION

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *services) servicesWithPath(path string) []*service {
	services := make([]*service, 0)
	for _, service := range this.service {
		if service.path == path {
			services = append(services, service)
		}
	}
	return services
}

func (this *services) add(executable string) *service {
	// Obtain a new Session ID
	sid := this.newId()
	if sid == 0 {
		return nil
	}

	// Add a new service from an executable
	service := &service{
		name:  filepath.Base(executable),
		sid:   sid,
		path:  executable,
		cwd:   this.user.HomeDir,
		user:  this.user.Uid,
		group: this.user.Gid,
	}
	this.service[sid] = service

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
		if _, exists := this.service[rand]; exists == false && rand > 0 {
			return rand
		} else {
			mod <<= 1
		}
	}
	return 0
}
