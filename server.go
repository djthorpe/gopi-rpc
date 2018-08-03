package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	// Frameworks
	"github.com/djthorpe/gopi"
	"github.com/miekg/dns"
)

type Listener struct {
	Interface *net.Interface
}

type ServiceRecord struct {
	Id        string
	Name      string
	Service   string
	Hostname  string
	IPv4      []net.IP
	IPv6      []net.IP
	Port      uint16
	TTL       time.Duration
	TXT       []string
	Timestamp time.Time
}

type listener struct {
	sync.Mutex
	log      gopi.Logger
	ipv4     *net.UDPConn
	ipv6     *net.UDPConn
	errors   chan error
	entries  map[string]*ServiceRecord
	shutdown bool
	ended    sync.WaitGroup
}

var (
	MDNS_ADDR_IPV4 = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	MDNS_ADDR_IPV6 = &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}
)

func (config Listener) Open(logger gopi.Logger) (gopi.Driver, error) {
	ipv4, _ := net.ListenMulticastUDP("udp4", config.Interface, MDNS_ADDR_IPV4)
	ipv6, _ := net.ListenMulticastUDP("udp6", config.Interface, MDNS_ADDR_IPV6)
	if ipv4 == nil && ipv6 == nil {
		return nil, fmt.Errorf("No multicast listeners could be started")
	} else {

		// Create listener
		this := new(listener)
		this.log = logger
		this.errors = make(chan error)
		this.shutdown = false
		this.entries = make(map[string]*ServiceRecord)

		// Start listening to connections
		if ipv4 != nil {
			this.ipv4 = ipv4
			go this.recv_loop(this.ipv4)
		}
		if ipv6 != nil {
			this.ipv6 = ipv6
			go this.recv_loop(this.ipv6)
		}

		// Receive errors and print them out
		go this.recv_errors()

		// TODO: Add in invalidation after TTL has been reached

		// Success
		return this, nil
	}
}

func (this *listener) Close() error {
	// Indicate shutdown
	this.shutdown = true
	// Close connections
	if this.ipv4 != nil {
		this.ipv4.Close()
	}
	if this.ipv6 != nil {
		this.ipv6.Close()
	}
	// Close errors channel
	close(this.errors)
	// Wait for go routines to end
	this.ended.Wait()
	// Return success
	return nil
}

// recv is a long running routine to receive packets from an interface
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
			this.errors <- err
		} else if service_record != nil {
			if service_record.TTL > 0 {
				if err := this.insert(service_record.Id, service_record); err != nil {
					this.errors <- err
				}
			} else {
				if err := this.remove(service_record.Id, service_record); err != nil {
					this.errors <- err
				}
			}
		}
	}
}

func (this *listener) insert(key string, record *ServiceRecord) error {
	this.Lock()
	defer this.Unlock()

	if _, exists := this.entries[key]; exists == false {
		// Add new entry
		fmt.Printf("ADD: %v\n", record)
		this.entries[key] = record
	} else {
		// Update existing entry
		fmt.Printf("UPD: %v\n", record)
		this.entries[key] = record
	}

	// Success
	return nil
}

func (this *listener) remove(key string, record *ServiceRecord) error {
	this.Lock()
	defer this.Unlock()

	if _, exists := this.entries[key]; exists == true {
		// Remove existing entry
		fmt.Printf("DEL: %v\n", record)
		delete(this.entries, key)
	}

	// Success
	return nil
}

func (this *listener) recv_errors() {
	// Indicate end of loop
	this.ended.Add(1)
	defer this.ended.Done()

	// Perform loop
	for {
		if e := <-this.errors; e == nil {
			break
		} else {
			this.log.Error("%v", e)
		}
	}
}

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
			entry.Id = rr.Ptr
			entry.Name = strings.TrimSuffix(strings.Replace(rr.Ptr, rr.Hdr.Name, "", -1), ".")
			entry.Service = rr.Hdr.Name
			entry.TTL = time.Duration(time.Second * time.Duration(rr.Hdr.Ttl))
		case *dns.SRV:
			entry.Hostname = rr.Target
			entry.Port = rr.Port
		case *dns.TXT:
			entry.TXT = rr.Txt
		}
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

func (this *ServiceRecord) complete() bool {
	if this.Id == "" {
		return false
	}
	return true
}
func (this *ServiceRecord) String() string {
	parts := make([]string, 0)
	if this.Id != "" {
		parts = append(parts, fmt.Sprintf("ID='%v'", this.Id))
	}
	if this.Name != "" {
		parts = append(parts, fmt.Sprintf("Name='%v'", this.Name))
	}
	if this.Service != "" {
		parts = append(parts, fmt.Sprintf("Service='%v'", this.Service))
	}
	if this.Hostname != "" {
		parts = append(parts, fmt.Sprintf("Hostname='%v'", this.Hostname))
	}
	if this.Port != 0 {
		parts = append(parts, fmt.Sprintf("Port=%v", this.Port))
	}
	parts = append(parts, fmt.Sprintf("TTL=%v", this.TTL))
	for _, addr := range this.IPv4 {
		parts = append(parts, fmt.Sprintf("Addr=%v", addr))
	}
	for _, addr := range this.IPv6 {
		parts = append(parts, fmt.Sprintf("Addr=%v", addr))
	}
	if len(this.TXT) > 0 {
		parts = append(parts, fmt.Sprintf("TXT='%v'", strings.Join(this.TXT, "|")))
	}
	return fmt.Sprintf("<ServiceRecord>{ %v ts=%v }", strings.Join(parts, " "), this.Timestamp)
}
