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
	rpc "github.com/djthorpe/gopi-rpc"
	errors "github.com/djthorpe/gopi/util/errors"
	dns "github.com/miekg/dns"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Listener struct {
	sync.WaitGroup

	domain    string
	end       bool
	ipv4      *net.UDPConn
	ipv6      *net.UDPConn
	errors    chan<- error
	services  chan<- *rpc.ServiceRecord
	questions chan<- string
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VERIABLES

const (
	MDNS_DEFAULT_DOMAIN = "local."
	MDNS_SERVICE_QUERY  = "_services._dns-sd._udp"
	DELTA_QUERY         = 5 * time.Second
)

var (
	MDNS_ADDR_IPV4 = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	MDNS_ADDR_IPV6 = &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}
)

////////////////////////////////////////////////////////////////////////////////
// INIT / DEINIT

func (this *Listener) Init(config Discovery, errors chan<- error, services chan<- *rpc.ServiceRecord, questions chan<- string) error {
	if config.Domain == "" {
		config.Domain = MDNS_DEFAULT_DOMAIN
	}
	if strings.HasSuffix(config.Domain, ".") == false {
		return gopi.ErrBadParameter
	}
	if errors == nil || services == nil {
		return gopi.ErrBadParameter
	}
	if config.Flags&gopi.RPC_FLAG_INET_V4 != 0 {
		this.ipv4, _ = net.ListenMulticastUDP("udp4", config.Interface, MDNS_ADDR_IPV4)
	}
	if config.Flags&gopi.RPC_FLAG_INET_V6 != 0 {
		this.ipv6, _ = net.ListenMulticastUDP("udp6", config.Interface, MDNS_ADDR_IPV6)
	}
	if this.ipv4 == nil && this.ipv6 == nil {
		return fmt.Errorf("No multicast listeners could be started")
	}

	// Set up the listener
	this.domain = config.Domain
	this.errors = errors
	this.services = services
	this.questions = questions

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

	// Wait for recv_loop go routines to end
	this.Wait()

	// Return compound errors
	return errs.ErrorOrSelf()
}

////////////////////////////////////////////////////////////////////////////////
// QUERY

func (this *Listener) Query(msg *dns.Msg, ctx context.Context) error {

	// Send out service message every five seconds
	ticker := time.NewTimer(1 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			if err := this.send(msg); err != nil {
				return err
			}
			// Restart timer to send query again
			ticker.Reset(DELTA_QUERY)
		case <-ctx.Done():
			ticker.Stop()
			return ctx.Err()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
// ANSWER

func (this *Listener) AnswerEnum(name string, ttl time.Duration) {
	fmt.Println("ANSWER", name, ttl)
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

// answer questions from remote
func (this *Listener) answer_questions(q dns.Question, from net.Addr) {
	domain := "." + strings.Trim(this.domain, ".") + "."
	if strings.HasSuffix(q.Name, domain) && this.questions != nil {
		this.questions <- strings.TrimSuffix(q.Name, domain)
	}
}

// parse packets into service records
func (this *Listener) parse_packet(packet []byte, from net.Addr) (*rpc.ServiceRecord, error) {
	var msg dns.Msg
	if err := msg.Unpack(packet); err != nil {
		return nil, err
	}
	if msg.Opcode != dns.OpcodeQuery {
		return nil, fmt.Errorf("Query with invalid Opcode %v (expected %v)", msg.Opcode, dns.OpcodeQuery)
	}
	if msg.Rcode != 0 {
		return nil, fmt.Errorf("Query with non-zero Rcode %v", msg.Rcode)
	}
	if msg.Truncated {
		return nil, fmt.Errorf("Support for DNS requests with high truncated bit not implemented")
	}

	// Deal with questions
	for _, q := range msg.Question {
		this.answer_questions(q, from)
	}
	if len(msg.Answer) == 0 {
		return nil, nil
	}

	// Make the entry
	entry := rpc.NewServiceRecord()

	// Process sections of the response
	sections := append(msg.Answer, msg.Ns...)
	sections = append(sections, msg.Extra...)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.PTR:
			entry.SetPTR(this.domain, rr)
		case *dns.SRV:
			entry.SetSRV(rr)
		case *dns.TXT:
			entry.SetTXT(rr)
		}
	}

	// Associate IPs in a second round
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.A:
			entry.AppendIP4(rr)
		case *dns.AAAA:
			entry.AppendIP6(rr)
		}
	}

	// Check the entry is valid
	if entry.Key() == "" {
		// Ensure entry is complete
		return nil, nil
	} else if strings.HasSuffix(entry.Key(), "."+this.domain) == false {
		// Domain doesn't match
		return nil, nil
	} else if entry.Service_ == MDNS_SERVICE_QUERY {
		// Strip domain
		entry.Name_ = strings.TrimPrefix(entry.Name_, ".") + "."
		if strings.HasSuffix(entry.Name(), "."+this.domain) == false {
			return nil, nil
		} else {
			entry.Name_ = strings.TrimSuffix(entry.Name(), "."+this.domain)
		}
	}

	// Success
	return entry, nil
}
