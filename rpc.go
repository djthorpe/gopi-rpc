/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpc

import (
	"encoding/json"
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
	Key_     string    `json:"key"`
	Name_    string    `json:"name"`
	Host_    string    `json:"host"`
	Service_ string    `json:"service"`
	Port_    uint      `json:"port"`
	Txt_     []string  `json:"txt"`
	Ipv4_    []net.IP  `json:"ipv4"`
	Ipv6_    []net.IP  `json:"ipv6"`
	Ts_      time.Time `json:"ts"`
	Ttl_     *Duration `json:"ttl"`
	Local_   bool      `json:"local"`
}

// RPCEvent implementation
type Event struct {
	s gopi.Driver
	t gopi.RPCEventType
	r gopi.RPCServiceRecord
}

// Duration type to read and write JSON better
type Duration struct {
	Duration time.Duration
}

// DiscoveryType is either DNS (using DNS-SD) or DB (using internal database)
type DiscoveryType uint

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	DISCOVERY_TYPE_NONE DiscoveryType = 0
	DISCOVERY_TYPE_DNS  DiscoveryType = 1
	DISCOVERY_TYPE_DB   DiscoveryType = 2
)

////////////////////////////////////////////////////////////////////////////////
// gopi.RPCServiceRecord IMPLEMENTATION

func NewServiceRecord() *ServiceRecord {
	return &ServiceRecord{
		Ts_:   time.Now(),
		Ipv4_: make([]net.IP, 0, 1),
		Ipv6_: make([]net.IP, 0, 1),
		Txt_:  make([]string, 0),
	}
}

func (this *ServiceRecord) Key() string {
	return this.Key_
}

func (this *ServiceRecord) Name() string {
	return this.Name_
}

func (this *ServiceRecord) Service() string {
	return this.Service_
}

func (this *ServiceRecord) Host() string {
	return this.Host_
}

func (this *ServiceRecord) Port() uint {
	return this.Port_
}

func (this *ServiceRecord) TTL() time.Duration {
	if this.Ttl_ == nil {
		return 0
	} else {
		return this.Ttl_.Duration
	}
}

func (this *ServiceRecord) Timestamp() time.Time {
	return this.Ts_
}

func (this *ServiceRecord) Text() []string {
	return this.Txt_
}

func (this *ServiceRecord) IP4() []net.IP {
	return this.Ipv4_
}

func (this *ServiceRecord) IP6() []net.IP {
	return this.Ipv6_
}

func (this *ServiceRecord) SetPTR(zone string, rr *dns.PTR) {
	this.Key_ = rr.Ptr
	this.Name_ = strings.TrimSuffix(strings.Replace(rr.Ptr, rr.Hdr.Name, "", -1), ".")
	this.Service_ = rr.Hdr.Name
	this.Ttl_ = &Duration{time.Second * time.Duration(rr.Hdr.Ttl)}

	// Sanitize zone and service
	if zone != "" {
		zone = "." + strings.Trim(zone, ".") + "."
		this.Service_ = strings.TrimSuffix(this.Service_, zone)
	}
}

func (this *ServiceRecord) SetSRV(rr *dns.SRV) {
	this.Host_ = rr.Target
	this.Port_ = uint(rr.Port)
}

func (this *ServiceRecord) SetTXT(rr *dns.TXT) {
	this.Txt_ = rr.Txt
}

func (this *ServiceRecord) AppendIP4(rr *dns.A) {
	this.Ipv4_ = append(this.Ipv4_, rr.A)
}

func (this *ServiceRecord) AppendIP6(rr *dns.AAAA) {
	this.Ipv6_ = append(this.Ipv6_, rr.AAAA)
}

func (this *ServiceRecord) Expired() bool {
	if this.Ts_.IsZero() {
		return false
	} else if this.Ttl_.Duration == 0 {
		return true
	} else {
		if time.Now().Sub(this.Ts_) > this.Ttl_.Duration {
			return true
		} else {
			return false
		}
	}
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
	if s.Name_ != "" {
		p += fmt.Sprintf("name=%v ", strconv.Quote(s.Name_))
	}
	if s.Service_ != "" {
		p += fmt.Sprintf("service=%v ", strconv.Quote(s.Service_))
	}
	if s.Host_ != "" {
		p += fmt.Sprintf("host=%v ", strconv.Quote(s.Host_))
	}
	if s.Port_ > 0 {
		p += fmt.Sprintf("port=%v ", s.Port_)
	}
	if len(s.Ipv4_) > 0 {
		p += fmt.Sprintf("ipv4=%v ", s.Ipv4_)
	}
	if len(s.Ipv6_) > 0 {
		p += fmt.Sprintf("ipv6=%v ", s.Ipv6_)
	}
	if s.Ttl_ != nil && s.Ttl_.Duration > 0 {
		p += fmt.Sprintf("ttl=%v ", s.Ttl_.Duration)
	}
	if len(s.Txt_) > 0 {
		p += "txt=["
		for i, txt := range s.Txt_ {
			if i > 0 {
				p += ","
			}
			p += strconv.Quote(txt)
		}
		p += "]"
	}

	return fmt.Sprintf("<gopi.RPCServiceRecord>{ %v }", strings.TrimSpace(p))
}

////////////////////////////////////////////////////////////////////////////////
// JSON SERIALIZATION

func (this *Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(fmt.Sprint(this.Duration))), nil
}

func (this *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if tmp, err := time.ParseDuration(s); err != nil {
		return err
	} else {
		this.Duration = tmp
		return nil
	}
}
