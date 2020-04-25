/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Service represents a service to be run
type service struct {
	name        string   // Name of the service
	sid         uint32   // Service ID
	path        string   // Path represents the path to the executable
	cwd         string   // Working directory on execution
	user, group string   // User and group for process
	args        []string // Process arguments
	enabled     bool     // Whether the service is enabled
	instances   uint32   // Minumum number of instances to run
}

////////////////////////////////////////////////////////////////////////////////
// VARIABLES

var (
	// Service name needs to start with an alpha character
	reName = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9\\-\\_\\.]{0,31}$")
)

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewServiceEx(path, user, group string, args []string) *service {
	return &service{
		name:      filepath.Base(path),
		path:      path,
		user:      user,
		group:     group,
		args:      args,
		instances: 1,
	}
}

func NewService(src rpc.GafferService, instances uint32) *service {
	// Return nil if the source is also nil
	if src == nil {
		return nil
	}

	// Make new version of src
	this := new(service)
	this.name = src.Name()
	this.sid = src.Sid()
	this.path = src.Path()
	this.cwd = src.Cwd()
	this.user = src.User()
	this.group = src.Group()
	this.args = nil
	if src.Args() != nil {
		this.args = make([]string, len(src.Args()))
		copy(this.args, src.Args())
	}
	this.enabled = src.Enabled()
	this.instances = instances

	// Return success
	return this
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION rpc.GafferService

func (this *service) Name() string {
	return this.name
}

func (this *service) Sid() uint32 {
	return this.sid
}

func (this *service) Path() string {
	return this.path
}

func (this *service) Cwd() string {
	return this.cwd
}

func (this *service) User() string {
	return this.user
}
func (this *service) Group() string {
	return this.group
}

func (this *service) Args() []string {
	return this.args
}

func (this *service) Enabled() bool {
	return this.enabled
}

////////////////////////////////////////////////////////////////////////////////
// VALID NAME

func (this *service) IsValidName(name string) bool {
	return reName.MatchString(name)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *service) String() string {
	str := "<GafferService"
	str += " sid=" + fmt.Sprint(this.sid)
	if this.name != "" {
		str += " name=" + strconv.Quote(this.name)
	}
	str += " path=" + strconv.Quote(this.path)
	if this.cwd != "" {
		str += " cwd=" + strconv.Quote(this.cwd)
	}
	if this.user != "" {
		str += " user=" + strconv.Quote(this.user)
	}
	if this.group != "" {
		str += " group=" + strconv.Quote(this.group)
	}
	if len(this.args) > 0 {
		str += " args=["
		for i, arg := range this.args {
			if i > 0 {
				str += " "
			}
			str += strconv.Quote(arg)
		}
		str += "]"
	}
	return str + ">"
}
