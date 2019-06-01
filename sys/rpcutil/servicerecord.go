/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package rpcutil

import (
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	dns "github.com/miekg/dns"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Record implements a gopi.RPCServiceRecord
type record struct {
	Key_     string            `json:"key"`
	Name_    string            `json:"name"`
	Host_    string            `json:"host"`
	Service_ string            `json:"service"`
	Port_    uint              `json:"port,omitempty"`
	Txt_     []string          `json:"txt,omitempty"`
	Ipv4_    []net.IP          `json:"ipv4,omitempty"`
	Ipv6_    []net.IP          `json:"ipv6,omitempty"`
	Ts_      *timestamp        `json:"ts,omitempty"`
	Ttl_     *duration         `json:"ttl,omitempty"`
	Source_  rpc.DiscoveryType `json:"source,omitempty"`
}

// duration type to read and write JSON better
type duration struct {
	Duration time.Duration
}

// timestamp type to read and write JSON better
type timestamp struct {
	Time time.Time
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VARIABLES

var (
	reServiceType           = regexp.MustCompile("^_[A-Za-z][A-Za-z0-9\\-]*\\._(tcp|udp)$")
	reSubType               = regexp.MustCompile("^[A-Za-z][A-Za-z0-9\\-]*$")
	qClassCacheFlush uint16 = 1 << 15
)

////////////////////////////////////////////////////////////////////////////////
// NEW

func (this *util) NewServiceRecord(source rpc.DiscoveryType) rpc.ServiceRecord {
	if source == rpc.DISCOVERY_TYPE_NONE {
		return nil
	}
	r := &record{
		Ipv4_:   make([]net.IP, 0, 1),
		Ipv6_:   make([]net.IP, 0, 1),
		Txt_:    make([]string, 0),
		Source_: source,
	}
	if source == rpc.DISCOVERY_TYPE_DNS {
		// Adding the timestamp allows record to expire
		r.Ts_ = &timestamp{time.Now()}
	}
	return r
}

////////////////////////////////////////////////////////////////////////////////
// SET

// Set service type, subtype and IP protocol
func (this *record) SetService(service, subtype string) error {
	if service == "" || reServiceType.MatchString(service) == false {
		return gopi.ErrBadParameter
	}
	if subtype != "" && reSubType.MatchString(subtype) == false {
		return gopi.ErrBadParameter
	}
	if subtype != "" {
		this.Service_ = "_" + subtype + "._sub." + service
	} else {
		this.Service_ = service
	}
	// Set key
	this.Key_ = this.Name_
	if this.Service() != "" {
		this.Key_ += "." + this.Service()
	}
	return nil
}

// SetName sets the name of the service instance
func (this *record) SetName(name string) error {
	this.Name_ = Quote(name)
	this.Key_ = this.Name_
	if this.Service() != "" {
		this.Key_ += "." + this.Service()
	}
	return nil
}

// SetPTR sets from PTR record
func (this *record) SetPTR(zone string, rr *dns.PTR) error {
	if rr == nil {
		return gopi.ErrBadParameter
	}

	// Sanitize zone
	zone = strings.Trim(zone, ".")

	// Set name, service type and TTL
	this.Key_ = strings.TrimSuffix(rr.Ptr, "."+zone+".")
	this.Service_ = strings.TrimSuffix(rr.Hdr.Name, "."+zone+".")
	this.Name_ = this.Key_

	// If not a discovery query then trim the name
	if this.Service_ != rpc.DISCOVERY_SERVICE_QUERY {
		this.Name_ = strings.TrimSuffix(this.Name_, "."+this.Service())
	}

	// Set TTL
	this.Ttl_ = &duration{time.Second * time.Duration(rr.Hdr.Ttl)}

	// Success
	return nil
}

func (this *record) SetSRV(zone string, rr *dns.SRV) error {
	if rr == nil {
		return gopi.ErrBadParameter
	} else if strings.HasSuffix(rr.Target, "."+zone+".") {
		this.Host_ = rr.Target
	} else {
		this.Host_ = strings.TrimPrefix(rr.Target, ".") + "." + zone + "."
	}
	this.Port_ = uint(rr.Port)
	return nil
}

func (this *record) SetTXT(txt []string) error {
	if len(txt) > 0 {
		this.Txt_ = txt
	} else {
		this.Txt_ = []string{}
	}

	// Success
	return nil
}

func (this *record) SetHostPort(addr string) error {
	if host, port_, err := net.SplitHostPort(addr); err != nil {
		return err
	} else if port, err := strconv.ParseUint(strings.TrimPrefix(port_, ":"), 10, 32); err != nil {
		return err
	} else {
		this.Host_ = host
		this.Port_ = uint(port)
	}

	// Success
	return nil
}

func (this *record) SetTTL(d time.Duration) error {
	this.Ttl_ = &duration{d.Truncate(time.Second)}
	return nil
}

func (this *record) AppendIP(ips ...net.IP) error {
	for _, ip := range ips {
		if ip == nil {
			return gopi.ErrBadParameter
		} else if ip4_ := ip.To4(); ip4_ != nil {
			this.Ipv4_ = append(this.Ipv4_, ip4_)
		} else if ip6_ := ip.To16(); ip6_ != nil {
			this.Ipv6_ = append(this.Ipv6_, ip6_)
		} else {
			return gopi.ErrBadParameter
		}
	}

	// Success
	return nil
}

func (this *record) AppendTXT(value ...string) error {
	if this.Txt_ == nil {
		this.Txt_ = make([]string, 0, len(value))
	}
	if len(value) > 0 {
		this.Txt_ = append(this.Txt_, value...)
	}
	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// GET

// Source returns the source of the record
func (this *record) Source() rpc.DiscoveryType {
	return this.Source_
}

func (this *record) Key() string {
	return this.Key_
}

func (this *record) Name() string {
	if name, err := Unquote(this.Name_); err != nil {
		return this.Name_
	} else {
		return name
	}
}

// Service returns the service type and protocol, including the underscores
// but removes the subtype information
func (this *record) Service() string {
	parts := strings.SplitN(this.Service_, "._sub.", 2)
	if len(parts) == 1 {
		return parts[0]
	} else if len(parts) == 2 {
		return parts[1]
	} else {
		return ""
	}
}

// Subtype returns the service subtype, but not including the underscore
func (this *record) Subtype() string {
	parts := strings.SplitN(this.Service_, "._sub.", 2)
	if len(parts) == 2 {
		return strings.TrimPrefix(parts[0], "_")
	} else {
		return ""
	}
}

func (this *record) Host() string {
	return this.Host_
}

func (this *record) Port() uint {
	return this.Port_
}

func (this *record) TTL() time.Duration {
	if this.Ttl_ == nil {
		return 0
	} else {
		return this.Ttl_.Duration
	}
}

func (this *record) Timestamp() time.Time {
	if this.Ts_ == nil {
		return time.Time{}
	} else {
		return this.Ts_.Time
	}
}

func (this *record) Text() []string {
	return this.Txt_
}

func (this *record) IP4() []net.IP {
	return this.Ipv4_
}

func (this *record) IP6() []net.IP {
	return this.Ipv6_
}

func (this *record) Expired() bool {
	if this.Ts_ == nil || this.Ts_.Time.IsZero() {
		return false
	} else if this.Ttl_ == nil {
		return false
	} else if this.Ttl_.Duration == 0 {
		return true
	} else {
		if time.Now().Sub(this.Ts_.Time) > this.Ttl_.Duration {
			return true
		} else {
			return false
		}
	}
}

func (this *record) PTR(zone string, ttl time.Duration) *dns.PTR {
	return &dns.PTR{
		Hdr: dns.RR_Header{
			Name:   this.Service_ + "." + zone + ".",
			Rrtype: dns.TypePTR,
			Class:  dns.ClassINET,
			Ttl:    uint32(ttl.Seconds()),
		},
		Ptr: this.Name_ + "." + zone + ".",
	}
}

func (this *record) SRV(zone string, ttl time.Duration) *dns.SRV {
	return &dns.SRV{
		Hdr: dns.RR_Header{
			Name:   this.Name_ + "." + zone + ".",
			Rrtype: dns.TypeSRV,
			Class:  dns.ClassINET | qClassCacheFlush,
			Ttl:    uint32(ttl.Seconds()),
		},
		Priority: 0,
		Weight:   0,
		Port:     uint16(this.Port_),
		Target:   this.Host_,
	}
}

func (this *record) TXT(zone string, ttl time.Duration) *dns.TXT {
	return &dns.TXT{
		Hdr: dns.RR_Header{
			Name:   this.Name_ + "." + zone + ".",
			Rrtype: dns.TypeTXT,
			Class:  dns.ClassINET,
			Ttl:    uint32(ttl.Seconds()),
		},
		Txt: this.Txt_,
	}
}

func (this *record) A(zone string, ttl time.Duration) []*dns.A {
	ips := this.IP4()
	sections := make([]*dns.A, len(ips))
	host := strings.TrimSuffix(this.Host_, ".")
	zone = "." + strings.Trim(zone, ".")
	if strings.HasSuffix(host, "."+zone) == false {
		host = host + zone
	}
	for i, ip := range ips {
		sections[i] = &dns.A{
			Hdr: dns.RR_Header{
				Name:   host + ".",
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl.Seconds()),
			},
			A: ip,
		}
	}
	return sections
}

func (this *record) AAAA(zone string, ttl time.Duration) []*dns.AAAA {
	ips := this.IP6()
	sections := make([]*dns.AAAA, len(ips))
	host := strings.TrimSuffix(this.Host_, ".")
	zone = "." + strings.Trim(zone, ".")
	if strings.HasSuffix(host, "."+zone) == false {
		host = host + zone
	}
	for i, ip := range ips {
		sections[i] = &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   host + ".",
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
				Ttl:    uint32(ttl.Seconds()),
			},
			AAAA: ip,
		}
	}
	return sections
}

////////////////////////////////////////////////////////////////////////////////
// Stringify

func (s *record) String() string {
	p := ""
	p += fmt.Sprintf("key=%v ", strconv.Quote(s.Key()))
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

func (this *duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(fmt.Sprint(this.Duration))), nil
}

func (this *duration) UnmarshalJSON(data []byte) error {
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

func (this *timestamp) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(this.Time.Format(time.RFC3339))), nil
}

func (this *timestamp) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if tmp, err := time.Parse(time.RFC3339, s); err != nil {
		return err
	} else {
		this.Time = tmp
		return nil
	}
}
