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
	"github.com/miekg/dns"
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
		ts:   time.Now(),
		ipv4: make([]net.IP, 2),
		ipv6: make([]net.IP, 2),
	}
}

func NewServiceRecordWithAddr(name, addr string) *ServiceRecord {
	if host, port_, err := net.SplitHostPort(addr); err != nil {
		return nil
	} else if port, err := strconv.ParseUint(strings.TrimPrefix(port_, ":"), 10, 32); err != nil {
		return nil
	} else if service := NewServiceRecord(); service == nil {
		return nil
	} else {
		service.name = name
		service.host = host
		service.port = uint(port)
		return service
	}
}

func (this *ServiceRecord) Key() string {
	return this.key
}

func (this *ServiceRecord) Name() string {
	return this.name
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

func (this *ServiceRecord) SetPTR(rr *dns.PTR) {
	this.key = rr.Ptr
	this.name = strings.Replace(rr.Ptr, rr.Hdr.Name, "", -1)
	this.service = rr.Hdr.Name
	this.ttl = time.Duration(time.Second * time.Duration(rr.Hdr.Ttl))
}

func (this *ServiceRecord) SetSRV(rr *dns.SRV) {
	this.host = rr.Target
	this.port = uint(rr.Port)
}

func (this *ServiceRecord) SetTXT(rr *dns.TXT) {
	this.txt = rr.Txt
}

func (this *ServiceRecord) AppendIP4(rr *dns.A) {
	this.ipv4 = append(this.ipv4, rr.A)
}

func (this *ServiceRecord) AppendIP6(rr *dns.AAAA) {
	this.ipv6 = append(this.ipv6, rr.AAAA)
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
