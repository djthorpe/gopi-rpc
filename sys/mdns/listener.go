/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2017
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package mdns

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

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Listener struct {
	Interface *net.Interface
}

type listener struct {
	sync.Mutex
	log      gopi.Logger
	ipv4     *net.UDPConn
	ipv6     *net.UDPConn
	entries  map[string]*gopi.RPCServiceRecord
	shutdown bool
	ended    sync.WaitGroup
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VERIABLES

var (
	MDNS_ADDR_IPV4 = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	MDNS_ADDR_IPV6 = &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}
)

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

func (config Listener) Open(logger gopi.Logger) (gopi.Driver, error) {
	logger.Debug("<rpc.mdns.Open>{ interface=%v }", config.Interface)

	ipv4, _ := net.ListenMulticastUDP("udp4", config.Interface, MDNS_ADDR_IPV4)
	ipv6, _ := net.ListenMulticastUDP("udp6", config.Interface, MDNS_ADDR_IPV6)
	if ipv4 == nil && ipv6 == nil {
		return nil, fmt.Errorf("No multicast listeners could be started")
	} else {

		// Create listener
		this := new(listener)
		this.log = logger
		this.shutdown = false
		this.entries = make(map[string]*gopi.RPCServiceRecord)

		// Start listening to connections
		if ipv4 != nil {
			this.ipv4 = ipv4
			go this.recv_loop(this.ipv4)
		}
		if ipv6 != nil {
			this.ipv6 = ipv6
			go this.recv_loop(this.ipv6)
		}

		// TODO: Add in invalidation after TTL has been reached

		// Success
		return this, nil
	}
}

func (this *listener) Close() error {
	this.log.Debug("<rpc.mdns.Close>{ }")

	// Indicate shutdown
	this.shutdown = true

	// Close connections
	if this.ipv4 != nil {
		this.ipv4.Close()
	}
	if this.ipv6 != nil {
		this.ipv6.Close()
	}

	// Wait for go routines to end
	this.ended.Wait()

	// Empty entries
	this.entries = nil

	// Return success
	return nil
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
				if err := this.insert(service_record.Id, service_record); err != nil {
					this.log.Warn("rpc.mdns.recv_loop: %v", err)
				}
			} else {
				if err := this.remove(service_record.Id, service_record); err != nil {
					this.log.Warn("rpc.mdns.recv_loop: %v", err)
				}
			}
		}
	}
}

// parse packets into service records
func (this *listener) parse(packet []byte, from net.Addr) (*gopi.RPCServiceRecord, error) {
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
	entry := &gopi.RPCServiceRecord{
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

func (this *listener) insert(key string, record *gopi.RPCServiceRecord) error {
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

func (this *listener) remove(key string, record *gopi.RPCServiceRecord) error {
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
