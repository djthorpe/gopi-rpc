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
	"sync/atomic"
	"time"

	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"

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

	util      rpc.Util
	domain    string
	end       int32
	ifaces    []*net.Interface
	ip4       *ipv4.PacketConn
	ip6       *ipv6.PacketConn
	errors    chan<- error
	services  chan<- rpc.ServiceRecord
	questions chan<- string
}

////////////////////////////////////////////////////////////////////////////////
// GLOBAL VERIABLES

const (
	MDNS_DEFAULT_DOMAIN = "local."
	DELTA_QUERY         = 5 * time.Second
)

var (
	MDNS_ADDR_IPV4 = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	MDNS_ADDR_IPV6 = &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}
)

////////////////////////////////////////////////////////////////////////////////
// INIT / DEINIT

func (this *Listener) Init(config Discovery, errors chan<- error, services chan<- rpc.ServiceRecord, questions chan<- string) error {
	if config.Domain == "" {
		config.Domain = MDNS_DEFAULT_DOMAIN
	}
	if strings.HasSuffix(config.Domain, ".") == false {
		return gopi.ErrBadParameter
	}
	if errors == nil || services == nil {
		return gopi.ErrBadParameter
	}
	if config.Util == nil {
		return gopi.ErrBadParameter
	} else {
		this.util = config.Util
	}
	if ifaces, err := listMulticastInterfaces(config.Interface); err != nil {
		return err
	} else {
		this.ifaces = ifaces
	}
	if config.Flags&gopi.RPC_FLAG_INET_V4 != 0 {
		if ip4, err := joinUdp4Multicast(this.ifaces, MDNS_ADDR_IPV4); err != nil {
			return err
		} else {
			this.ip4 = ip4
		}
	}
	if config.Flags&gopi.RPC_FLAG_INET_V6 != 0 {
		if ip6, err := joinUdp6Multicast(this.ifaces, MDNS_ADDR_IPV6); err != nil {
			return err
		} else {
			this.ip6 = ip6
		}
	}
	if this.ip4 == nil && this.ip6 == nil {
		return fmt.Errorf("No multicast listeners could be started")
	}

	// Set up the listener
	this.domain = config.Domain
	this.errors = errors
	this.services = services
	this.questions = questions

	// Start listening to connections
	go this.recv_loop4(this.ip4)
	go this.recv_loop6(this.ip6)

	return nil
}

func (this *Listener) Destroy() error {

	// Indicate shutdown
	if !atomic.CompareAndSwapInt32(&this.end, 0, 1) {
		return nil
	}

	// More than one error can be returned
	errs := errors.CompoundError{}

	// Close connections
	if this.ip4 != nil {
		if err := this.ip4.Close(); err != nil {
			errs.Add(err)
		}
	}
	if this.ip6 != nil {
		if err := this.ip6.Close(); err != nil {
			errs.Add(err)
		}
	}

	// Wait for recv_loop go routines to end
	this.Wait()

	// Release resources
	this.ip4 = nil
	this.ip6 = nil

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
	if this.ip4 != nil {
		params += fmt.Sprintf("ip4=%v ", this.ip4.LocalAddr())
	}
	if this.ip6 != nil {
		params += fmt.Sprintf("ip6=%v ", this.ip6.LocalAddr())
	}
	for i, iface := range this.ifaces {
		if i == 0 {
			params += "iface="
		} else {
			params += ","
		}
		params += strconv.Quote(iface.Name)
	}
	return fmt.Sprintf("<Listener>{ %v }", strings.TrimSpace(params))
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// recv_loop is a long running routine to receive packets from an interface
func (this *Listener) recv_loop4(conn *ipv4.PacketConn) {
	// Sanity check
	if conn == nil {
		return
	}

	// Indicate end of loop
	this.Add(1)
	defer this.Done()

	// Perform loop
	buf := make([]byte, 65536)
	for atomic.LoadInt32(&this.end) == 0 {
		if n, cm, from, err := conn.ReadFrom(buf); err != nil {
			continue
		} else if cm == nil {
			continue
		} else if service, err := this.parse_packet(buf[:n], cm.IfIndex, from); err != nil {
			this.errors <- err
		} else if service != nil {
			this.services <- service
		}
	}
}

func (this *Listener) recv_loop6(conn *ipv6.PacketConn) {
	// Sanity check
	if conn == nil {
		return
	}

	// Indicate end of loop
	this.Add(1)
	defer this.Done()

	// Perform loop
	buf := make([]byte, 65536)
	for atomic.LoadInt32(&this.end) == 0 {
		if n, cm, from, err := conn.ReadFrom(buf); err != nil {
			continue
		} else if cm == nil {
			continue
		} else if service, err := this.parse_packet(buf[:n], cm.IfIndex, from); err != nil {
			this.errors <- err
		} else if service != nil {
			this.services <- service
		}
	}
}

// send is used to multicast a query out
func (this *Listener) send(q *dns.Msg) error {
	return gopi.ErrNotImplemented
	/*
		if buf, err := q.Pack(); err != nil {
			return err
		} else {
			if this.ip4 != nil {
				if _, err := this.ip4.WriteToUDP(buf, MDNS_ADDR_IPV4); err != nil {
					return err
				}
			}
			if this.ip6 != nil {
				if _, err := this.ip6.WriteToUDP(buf, MDNS_ADDR_IPV6); err != nil {
					return err
				}
			}
		}
		return nil
	*/
}

// answer questions from remote
func (this *Listener) answer_questions(q dns.Question, ifIndex int, from net.Addr) {
	domain := "." + strings.Trim(this.domain, ".") + "."
	if strings.HasSuffix(q.Name, domain) && this.questions != nil {
		// TODO: Handle question on particular interface
		fmt.Println("answer_questions:", q)
		fmt.Println("    from:", from, "ifIndex:", ifIndex)
		this.questions <- strings.TrimSuffix(q.Name, domain)
	}
}

// parse packets into service records
func (this *Listener) parse_packet(packet []byte, ifIndex int, from net.Addr) (rpc.ServiceRecord, error) {
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

	// Deal with questions, and return nil if no answers
	for _, q := range msg.Question {
		this.answer_questions(q, ifIndex, from)
	}
	if len(msg.Answer) == 0 {
		return nil, nil
	}

	// Make the entry
	record := this.util.NewServiceRecord(rpc.DISCOVERY_TYPE_DNS)

	// Process sections of the response
	sections := append(msg.Answer, msg.Ns...)
	sections = append(sections, msg.Extra...)
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.PTR:
			if err := record.SetPTR(this.domain, rr); err != nil {
				return nil, err
			}
		case *dns.SRV:
			if err := record.SetSRV(rr); err != nil {
				return nil, err
			}
		case *dns.TXT:
			if err := record.SetTXT(rr.Txt); err != nil {
				return nil, err
			}
		}
	}

	// If this is a reverse lookup (.ip6.arpa. or .in-addr.arpa.) then ignore it
	if strings.HasSuffix(record.Service(), ".ip6.arpa.") {
		return nil, nil
	} else if strings.HasSuffix(record.Service(), ".in-addr.arpa.") {
		return nil, nil
	}

	// Associate IPs in a second round
	for _, answer := range sections {
		switch rr := answer.(type) {
		case *dns.A:
			if err := record.AppendIP(rr.A); err != nil {
				return nil, err
			}
		case *dns.AAAA:
			if err := record.AppendIP(rr.AAAA); err != nil {
				return nil, err
			}
		}
	}

	// Check the entry is valid
	if record.Key() == "" {
		return nil, nil
	}

	// Success
	return record, nil
}

// multicastSend sends a multicast response packet to a particular interface
// or all interfaces if 0
func (this *Listener) multicastSend(msg *dns.Msg, ifIndex int) error {
	var buf []byte
	if msg == nil {
		return gopi.ErrBadParameter
	} else if buf_, err := msg.Pack(); err != nil {
		return err
	} else {
		buf = buf_
	}
	if this.ip4 != nil {
		var cm ipv4.ControlMessage
		if ifIndex != 0 {
			cm.IfIndex = ifIndex
			this.ip4.WriteTo(buf, &cm, MDNS_ADDR_IPV4)
		} else {
			for _, intf := range this.ifaces {
				cm.IfIndex = intf.Index
				this.ip4.WriteTo(buf, &cm, MDNS_ADDR_IPV4)
			}
		}
	}
	if this.ip6 != nil {
		var cm ipv6.ControlMessage
		if ifIndex != 0 {
			cm.IfIndex = ifIndex
			this.ip6.WriteTo(buf, &cm, MDNS_ADDR_IPV6)
		} else {
			for _, intf := range this.ifaces {
				cm.IfIndex = intf.Index
				this.ip6.WriteTo(buf, &cm, MDNS_ADDR_IPV6)
			}
		}
	}
	// Success
	return nil
}

func joinUdp6Multicast(ifaces []*net.Interface, addr *net.UDPAddr) (*ipv6.PacketConn, error) {
	if len(ifaces) == 0 {
		return nil, gopi.ErrBadParameter
	} else if conn, err := net.ListenUDP("udp6", addr); err != nil {
		return nil, err
	} else if packet_conn := ipv6.NewPacketConn(conn); packet_conn == nil {
		return nil, conn.Close()
	} else {
		packet_conn.SetControlMessage(ipv6.FlagInterface, true)
		errs := &errors.CompoundError{}
		for _, iface := range ifaces {
			if err := packet_conn.JoinGroup(iface, &net.UDPAddr{IP: addr.IP}); err != nil {
				errs.Add(fmt.Errorf("%v: %v", iface.Name, err))
			}
		}
		if errs.Success() {
			return packet_conn, nil
		}
		errs.Add(conn.Close())
		return nil, errs.ErrorOrSelf()
	}
}

func joinUdp4Multicast(ifaces []*net.Interface, addr *net.UDPAddr) (*ipv4.PacketConn, error) {
	if len(ifaces) == 0 {
		return nil, gopi.ErrBadParameter
	} else if conn, err := net.ListenUDP("udp4", addr); err != nil {
		return nil, err
	} else if packet_conn := ipv4.NewPacketConn(conn); packet_conn == nil {
		return nil, conn.Close()
	} else {
		packet_conn.SetControlMessage(ipv4.FlagInterface, true)
		errs := &errors.CompoundError{}
		for _, ifi := range ifaces {
			if err := packet_conn.JoinGroup(ifi, &net.UDPAddr{IP: addr.IP}); err != nil {
				errs.Add(fmt.Errorf("%v: %v", ifi.Name, err))
			}
		}
		if errs.Success() {
			return packet_conn, nil
		}
		errs.Add(conn.Close())
		return nil, errs.ErrorOrSelf()
	}
}

func listMulticastInterfaces(iface *net.Interface) ([]*net.Interface, error) {
	if iface != nil {
		if (iface.Flags&net.FlagUp) > 0 && (iface.Flags&net.FlagMulticast) > 0 {
			return []*net.Interface{iface}, nil
		} else {
			return nil, fmt.Errorf("Interface %v is not up and/or multicast-enabled", iface.Name)
		}
	}
	if ifaces, err := net.Interfaces(); err != nil {
		return nil, err
	} else {
		interfaces := make([]*net.Interface, 0, len(ifaces))
		for _, ifi := range ifaces {
			if (ifi.Flags & net.FlagUp) == 0 {
				continue
			}
			if (ifi.Flags & net.FlagMulticast) > 0 {
				ifi2 := ifi
				interfaces = append(interfaces, &ifi2)
			}
		}
		if len(interfaces) > 0 {
			return interfaces, nil
		} else {
			return nil, fmt.Errorf("No multicast-enabled interface found")
		}
	}
}
