/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type castdevice struct {
	name string
	ip4  []net.IP
	ip6  []net.IP
	txt  map[string]string
	port uint
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewCastDevice(service gopi.RPCServiceRecord) *castdevice {
	this := new(castdevice)
	this.name = service.Name()
	this.ip4 = service.IP4()
	this.ip6 = service.IP6()
	this.port = service.Port()
	this.txt = make(map[string]string)
	for _, txt := range service.Text() {
		if pair := strings.SplitN(txt, "=", 2); len(pair) == 2 {
			this.txt[pair[0]] = pair[1]
		}
	}
	if this.Id() == "" {
		return nil
	} else {
		return this
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *castdevice) String() string {
	return fmt.Sprintf("<GoogleCastDevice>{ id=%v name=%v model=%v service=%v state=%v }", this.Id(), strconv.Quote(this.Name()), strconv.Quote(this.Model()), strconv.Quote(this.Service()), this.State())
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *castdevice) Id() string {
	return this.txt["id"]
}

func (this *castdevice) Name() string {
	if value, exists := this.txt["fn"]; exists == false || value == "" {
		return "Unknown"
	} else {
		return value
	}
}

func (this *castdevice) Model() string {
	if value, exists := this.txt["md"]; exists == false || value == "" {
		return "Unknown"
	} else {
		return value
	}
}

func (this *castdevice) Service() string {
	if value, exists := this.txt["rs"]; exists == false || value == "" {
		return ""
	} else {
		return value
	}
}

func (this *castdevice) State() uint {
	if value, exists := this.txt["st"]; exists == false || value == "" {
		return 0
	} else if value_, err := strconv.ParseUint(value, 10, 32); err != nil {
		return 0
	} else {
		return uint(value_)
	}
}

func (this *castdevice) EqualsState(other *castdevice) bool {
	if this.Model() != other.Model() {
		return false
	}
	if this.Service() != other.Service() {
		return false
	}
	if this.Name() != other.Name() {
		return false
	}
	if this.State() != other.State() {
		return false
	}
	return true
}
