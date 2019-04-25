/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	iface "github.com/djthorpe/gopi-rpc"
	errors "github.com/djthorpe/gopi/util/errors"
	dns "github.com/miekg/dns"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Listener struct {
	sync.WaitGroup

	domain   string
	end      bool
	ipv4     *net.UDPConn
	ipv6     *net.UDPConn
	errors   chan<- error
	services chan<- *ServiceRecord
}

type ServiceRecord struct {
	Key       string
	Name      string
	Host      string
	Service   string
	Port      uint16
	TXT       []string
	IPv4      []net.IP
	IPv6      []net.IP
	Timestamp time.Time
	TTL       time.Duration
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VERIABLES

const (
	MDNS_DEFAULT_DOMAIN = "local."
	MDNS_SERVICE_QUERY  = "_services._dns-sd._udp"
)

var (
	MDNS_ADDR_IPV4 = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	MDNS_ADDR_IPV6 = &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}
)

////////////////////////////////////////////////////////////////////////////////
// INIT / DEINIT

func (this *Listener) Init(config Discovery, errors chan<- error, services chan<- *ServiceRecord) error {
	if config.Domain == "" {
		config.Domain = MDNS_DEFAULT_DOMAIN
	}
	if strings.HasSuffix(config.Domain, ".") == false {
		return gopi.ErrBadParameter
	}
	if errors == nil || services == nil {
		return gopi.ErrBadParameter
	}
	if config.Flags|iface.RPC_FLAG_IPV4 != 0 {
		this.ipv4, _ = net.ListenMulticastUDP("udp4", config.Interface, MDNS_ADDR_IPV4)
	}
	if config.Flags|iface.RPC_FLAG_IPV6 != 0 {
		this.ipv6, _ = net.ListenMulticastUDP("udp6", config.Interface, MDNS_ADDR_IPV6)
	}
	if this.ipv4 == nil && this.ipv6 == nil {
		return fmt.Errorf("No multicast listeners could be started")
	}

	// Set up the listener
	this.domain = config.Domain
	this.errors = errors
	this.services = services

	// Start listening to connections
	go this.recv_loop(this.ipv4)
	go this.recv_loop(this.ipv6)

	return nil
}

func (this *Listener) Destroy() error {

	// Indicate shutdown
	this.end = true

	// More than one error can be returned
	errs := errors.CompoundError{}

	// Close connections
	if this.ipv4 != nil {
		if err := this.ipv4.Close(); err != nil {
			errs.Add(err)
		}
	}
	if this.ipv6 != nil {
		if err := this.ipv6.Close(); err != nil {
			errs.Add(err)
		}
	}

	// Wait for go routines to end
	this.Wait()

	// Return compound errors
	return errs.ErrorOrSelf()
}

////////////////////////////////////////////////////////////////////////////////
// QUERY

func (this *Listener) Query(ctx context.Context) error {
	msg := new(dns.Msg)
	msg.SetQuestion(MDNS_SERVICE_QUERY+"."+this.domain, dns.TypePTR)
	msg.RecursionDesired = false

	// Send out service message
	ticker := time.NewTimer(1 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if err := this.send(msg); err != nil {
				return err
			}
			// Restart timer after 5 seconds
			ticker.Reset(5 * time.Second)
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this Listener) String() string {
	params := ""
	if this.domain != "" {
		params += fmt.Sprintf("domain=%v ", strconv.Quote(this.domain))
	}
	if this.ipv4 != nil {
		params += fmt.Sprintf("ip4=%v ", this.ipv4.LocalAddr())
	}
	if this.ipv6 != nil {
		params += fmt.Sprintf("ip6=%v ", this.ipv6.LocalAddr())
	}
	return fmt.Sprintf("<Listener>{ %v }", strings.TrimSpace(params))
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// recv_loop is a long running routine to receive packets from an interface
func (this *Listener) recv_loop(conn *net.UDPConn) {
	// Sanity check
	if conn == nil {
		return
	}

	// Indicate end of loop
	this.Add(1)
	defer this.Done()

	// Perform loop
	buf := make([]byte, 65536)
	for this.end == false {
		if n, from, err := conn.ReadFrom(buf); err != nil {
			continue
		} else if service, err := this.parse_packet(buf[:n], from); err != nil {
			this.errors <- err
		} else if service != nil {
			this.services <- service
		}
	}
}

// send is used to multicast a query out
func (this *Listener) send(q *dns.Msg) error {
	if buf, err := q.Pack(); err != nil {
		return err
	} else {
		if this.ipv4 != nil {
			if _, err := this.ipv4.WriteToUDP(buf, MDNS_ADDR_IPV4); err != nil {
				return err
			}
		}
		if this.ipv6 != nil {
			if _, err := this.ipv6.WriteToUDP(buf, MDNS_ADDR_IPV6); err != nil {
				return err
			}
		}
	}
	return nil
}

// parse packets into service records
func (this *Listener) parse_packet(packet []byte, from net.Addr) (*ServiceRecord, error) {
	var msg dns.Msg
	if err := msg.Unpack(packet); err != nil {
		return nil, err
	}
	if msg.Opcode != dns.OpcodeQuery {
		return nil, fmt.Errorf("Query with non-zero Opcode %v", msg.Opcode)
	}
	if msg.Rcode != 0 {
		return nil, fmt.Errorf("Query with non-zero Rcode %v", msg.Rcode)
	}
	if msg.Truncated {
		return nil, fmt.Errorf("Support for DNS requests with high truncated bit not implemented")
	}

	// Make the entry
	entry := &ServiceRecord{
		Timestamp: time.Now(),
	}

	// Process sections of the response
	sections := append(msg.Answer, msg.Ns...)
	sections = append(sections, msg.Extra...)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.PTR:
			// Obtain the name and service
			entry.Key = rr.Ptr
			entry.Name = strings.TrimSuffix(strings.Replace(rr.Ptr, rr.Hdr.Name, "", -1), ".")
			entry.Service = rr.Hdr.Name
			entry.TTL = time.Duration(time.Second * time.Duration(rr.Hdr.Ttl))
		case *dns.SRV:
			entry.Host = rr.Target
			entry.Port = rr.Port
		case *dns.TXT:
			entry.TXT = rr.Txt
		}
	}

	// Check the entry ServiceDomain matches this domain
	if strings.HasSuffix(entry.Service, "."+this.domain) {
		entry.Service = strings.TrimSuffix(entry.Service, "."+this.domain)
	} else {
		return nil, nil
	}

	// Associate IPs in a second round
	entry.IPv4 = make([]net.IP, 0)
	entry.IPv6 = make([]net.IP, 0)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.A:
			entry.IPv4 = append(entry.IPv4, rr.A)
		case *dns.AAAA:
			entry.IPv6 = append(entry.IPv4, rr.AAAA)
		}
	}
	// Ensure entry is complete
	if entry.complete() {
		return entry, nil
	} else {
		return nil, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *ServiceRecord) complete() bool {
	if this.Key == "" {
		return false
	}
	return true
}

/*
// EnumerateServiceNames browses for services available on the domain
// and publishes the service names as events
func (this *listener) EnumerateServiceNames(ctx context.Context) error {
	msg := new(dns.Msg)
	msg.SetQuestion(MDNS_SERVICE_QUERY+"."+this.domain, dns.TypePTR)
	msg.RecursionDesired = false

	// Send out service message
	ticker := time.NewTimer(1 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if err := this.send(msg); err != nil {
				return err
			}
			// Restart timer after 5 seconds
			ticker.Reset(5 * time.Second)
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		}
	}
}

func (this *listener) String() string {
	return fmt.Sprintf("<rpc.mdns.Server>{ domain='%v' }", this.domain)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// recv_loop is a long running routine to receive packets from an interface
func (this *listener) recv_loop(conn *net.UDPConn) {
	// Sanity check
	if conn == nil {
		return
	}

	// Indicate end of loop
	this.ended.Add(1)
	defer this.ended.Done()

	// Perform loop
	buf := make([]byte, 65536)
	for this.shutdown == false {
		if n, from, err := conn.ReadFrom(buf); err != nil {
			continue
		} else if service_record, err := this.parse(buf[:n], from); err != nil {
			this.log.Warn("rpc.mdns.recv_loop: %v", err)
		} else if service_record != nil {
			if service_record.TTL > 0 {
				this.log.Debug2("rpc.mdns.insert{ service_record=%v }", service_record)
				if err := this.insert(service_record.Key, service_record); err != nil {
					this.log.Warn("rpc.mdns.recv_loop: %v", err)
				}
			} else {
				this.log.Debug2("rpc.mdns.remove{ service_record=%v }", service_record)
				if err := this.remove(service_record.Key, service_record); err != nil {
					this.log.Warn("rpc.mdns.recv_loop: %v", err)
				}
			}
		}
	}
}

// parse packets into service records
func (this *listener) parse(packet []byte, from net.Addr) (*ServiceRecord, error) {
	var msg dns.Msg
	if err := msg.Unpack(packet); err != nil {
		return nil, err
	}
	if msg.Opcode != dns.OpcodeQuery {
		return nil, fmt.Errorf("Query with non-zero Opcode %v", msg.Opcode)
	}
	if msg.Rcode != 0 {
		return nil, fmt.Errorf("Query with non-zero Rcode %v", msg.Rcode)
	}
	if msg.Truncated {
		return nil, fmt.Errorf("Support for DNS requests with high truncated bit not implemented")
	}

	// Make the entry
	entry := &ServiceRecord{
		Timestamp: time.Now(),
	}

	// Process sections of the response
	sections := append(msg.Answer, msg.Ns...)
	sections = append(sections, msg.Extra...)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.PTR:
			// Obtain the name and service
			entry.Key = rr.Ptr
			entry.Name = strings.TrimSuffix(strings.Replace(rr.Ptr, rr.Hdr.Name, "", -1), ".")
			entry.Service = rr.Hdr.Name
			entry.TTL = time.Duration(time.Second * time.Duration(rr.Hdr.Ttl))
		case *dns.SRV:
			entry.Host = rr.Target
			entry.Port = rr.Port
		case *dns.TXT:
			entry.TXT = rr.Txt
		}
	}

	// Check the entry ServiceDomain matches this domain
	if strings.HasSuffix(entry.Service, "."+this.domain) {
		entry.Service = strings.TrimSuffix(entry.Service, "."+this.domain)
	} else {
		return nil, nil
	}

	// Associate IPs in a second round
	entry.IPv4 = make([]net.IP, 0)
	entry.IPv6 = make([]net.IP, 0)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.A:
			entry.IPv4 = append(entry.IPv4, rr.A)
		case *dns.AAAA:
			entry.IPv6 = append(entry.IPv4, rr.AAAA)
		}
	}
	// Ensure entry is complete
	if entry.complete() {
		return entry, nil
	} else {
		return nil, nil
	}
}

func (this *listener) insert(key string, record *ServiceRecord) error {
	this.Lock()
	defer this.Unlock()

	if record.Service == MDNS_SERVICE_QUERY {
		if _, exists := this.entries[key]; exists == false {
			domain := "." + this.domain
			if strings.HasSuffix(record.Key, domain) {
				this.emit_service_event(strings.TrimSuffix(record.Key, domain))
			}
			this.entries[key] = record
		}
	} else if _, exists := this.entries[key]; exists == false {
		// Add new entry
		fmt.Printf("ADD: %v\n", record)
		this.entries[key] = record
	} else {
		// Update existing entry
		if record.equals(this.entries[key]) == false {
			fmt.Printf("UPD: %v\n", record)
		}
		this.entries[key] = record
	}

	// Success
	return nil
}

func (this *listener) remove(key string, record *ServiceRecord) error {
	this.Lock()
	defer this.Unlock()

	if record.Service == MDNS_SERVICE_QUERY {
		delete(this.entries, key)
	} else if _, exists := this.entries[key]; exists == true {
		// Remove existing entry
		fmt.Printf("DEL: %v\n", record)
		delete(this.entries, key)
	}

	// Success
	return nil
}

func (this *listener) ttl_expire(start chan<- event.Signal, stop <-chan event.Signal) error {
	this.log.Debug("START ttl_expire")
	start <- gopi.DONE

	timer := time.NewTicker(500 * time.Millisecond)

FOR_LOOP:
	for {
		select {
		case <-timer.C:
			// look for expiring TTL records in a very non-optimal way
			expired_keys := make([]string, 0, 1)
			for _, entry := range this.entries {
				if time.Now().After(entry.Timestamp.Add(entry.TTL)) {
					expired_keys = append(expired_keys, entry.Key)
				}
			}
			for _, key := range expired_keys {
				fmt.Printf("EXP: %v\n", this.entries[key])
				delete(this.entries, key)
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	this.log.Debug("STOP ttl_expire")
	return nil
}

func (this *listener) emit_service_event(service_name string) {
	this.Emit(rpc.NewEvent(gopi.RPC_EVENT_SERVICE_NAME))
}
*/
