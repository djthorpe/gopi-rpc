/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// RPCServiceRecord implementation
type ServiceRecord struct {
	key     string
	name    string
	host    string
	service string
	port    uint
	txt     []string
	ipv4    []net.IP
	ipv6    []net.IP
	ts      time.Time
	ttl     time.Duration
}

// RPCEvent implementation
type Event struct {
	s gopi.Driver
	t gopi.RPCEventType
	r gopi.RPCServiceRecord
}

////////////////////////////////////////////////////////////////////////////////
// SERVICERECORD IMPLEMENTATION

func NewServiceRecord() *ServiceRecord {
	return &ServiceRecord{
		ts: time.Now(),
	}
}

func (this *ServiceRecord) SetKey(value string) {
	this.key = value
}

func (this *ServiceRecord) Key() string {
	return this.key
}

func (this *ServiceRecord) SetName(value string) {
	this.name = strings.TrimPrefix(value, ".")
}

func (this *ServiceRecord) Name() string {
	return this.name
}

func (this *ServiceRecord) SetService(value string) {
	this.service = value
}

func (this *ServiceRecord) Service() string {
	return this.service
}

func (this *ServiceRecord) Host() string {
	return this.host
}

func (this *ServiceRecord) Port() uint {
	return this.port
}

func (this *ServiceRecord) SetTTL(value time.Duration) {
	this.ttl = value
}

func (this *ServiceRecord) TTL() time.Duration {
	return this.ttl
}

func (this *ServiceRecord) Timestamp() time.Time {
	return this.ts
}

func (this *ServiceRecord) Text() []string {
	return this.txt
}

func (this *ServiceRecord) IP4() []net.IP {
	return this.ipv4
}

func (this *ServiceRecord) IP6() []net.IP {
	return this.ipv6
}

////////////////////////////////////////////////////////////////////////////////
// EVENT IMPLEMENTATION

func NewEvent(source gopi.Driver, type_ gopi.RPCEventType, service gopi.RPCServiceRecord) *Event {
	return &Event{s: source, t: type_, r: service}
}

// Return the type of event
func (this *Event) Type() gopi.RPCEventType {
	return this.t
}

// Return the service record
func (this *Event) ServiceRecord() gopi.RPCServiceRecord {
	return this.r
}

// Return name of event
func (*Event) Name() string {
	return "RPCEvent"
}

// Return source of event
func (this *Event) Source() gopi.Driver {
	return this.s
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Event) String() string {
	if this.r != nil {
		return fmt.Sprintf("<rpc.event>{ type=%v record=%v }", this.t, this.r)
	} else {
		return fmt.Sprintf("<rpc.event>{ type=%v }", this.t)
	}
}

func (s *ServiceRecord) String() string {
	p := ""
	if s.name != "" {
		p += fmt.Sprintf("name=%v ", strconv.Quote(s.name))
	}
	if s.service != "" {
		p += fmt.Sprintf("service=%v ", strconv.Quote(s.service))
	}
	if s.host != "" {
		p += fmt.Sprintf("service=%v ", strconv.Quote(s.host))
	}
	if s.port > 0 {
		p += fmt.Sprintf("port=%v ", s.port)
	}
	if len(s.ipv4) > 0 {
		p += fmt.Sprintf("ipv4=%v ", s.ipv4)
	}
	if len(s.ipv6) > 0 {
		p += fmt.Sprintf("ipv6=%v ", s.ipv6)
	}
	if s.ttl > 0 {
		p += fmt.Sprintf("ttl=%v ", s.ttl)
	}
	if len(s.txt) > 0 {
		p += "txt=["
		for _, txt := range s.txt {
			p += strconv.Quote(txt) + ","
		}
		p += "]"
	}
	return fmt.Sprintf("<gopi.RPCServiceRecord>{ %v }", strings.TrimSpace(p))
}
