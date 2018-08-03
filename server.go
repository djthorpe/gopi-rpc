package main

import (
	"fmt"
	"net"
	"os"
	"sync"

	// Frameworks
	"github.com/miekg/dns"
)

type Listener struct {
	ipv4   *net.UDPConn
	ipv6   *net.UDPConn
	errors chan error
	ended  sync.WaitGroup
}

var (
	MDNS_ADDR_IPV4 = &net.UDPAddr{IP: net.ParseIP("224.0.0.251"), Port: 5353}
	MDNS_ADDR_IPV6 = &net.UDPAddr{IP: net.ParseIP("ff02::fb"), Port: 5353}
)

func NewListener(iface *net.Interface) (*Listener, error) {
	ipv4, _ := net.ListenMulticastUDP("udp4", iface, MDNS_ADDR_IPV4)
	ipv6, _ := net.ListenMulticastUDP("udp6", iface, MDNS_ADDR_IPV6)
	if ipv4 == nil && ipv6 == nil {
		return nil, fmt.Errorf("No multicast listeners could be started")
	} else {
		// Create listener
		this := new(Listener)
		// Create error channel
		this.errors = make(chan error)
		// Start listening to connections
		if ipv4 != nil {
			this.ipv4 = ipv4
			this.ended.Add(1)
			go this.recv_loop(this.ipv4)
		}
		if ipv6 != nil {
			this.ipv6 = ipv6
			this.ended.Add(1)
			go this.recv_loop(this.ipv6)
		}
		// Receive errors and print them out
		this.ended.Add(1)
		go this.recv_errors()
		// Success
		return this, nil
	}
}

func (this *Listener) Close() error {
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
func (this *Listener) recv_loop(conn *net.UDPConn) {
	if conn == nil {
		return
	}
	buf := make([]byte, 65536)
	fmt.Println("START OF recv")
	for {
		if n, from, err := conn.ReadFrom(buf); err != nil && n == 0 {
			this.errors <- err
			break
		} else if err != nil {
			this.errors <- err
			continue
		} else if err := this.parse(buf[:n], from); err != nil {
			this.errors <- err
		}
	}
	fmt.Println("END OF recv")
	this.ended.Done()
}

func (this *Listener) recv_errors() {
	fmt.Println("START OF errs")
	for {
		if e := <-this.errors; e == nil {
			break
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", e)
		}
	}
	fmt.Println("END OF errs")
	this.ended.Done()
}

func (this *Listener) parse(packet []byte, from net.Addr) error {
	var msg dns.Msg
	if err := msg.Unpack(packet); err != nil {
		return err
	}
	if msg.Opcode != dns.OpcodeQuery {
		return fmt.Errorf("Query with non-zero Opcode %v", msg.Opcode)
	}
	if msg.Rcode != 0 {
		return fmt.Errorf("Query with non-zero Rcode %v", msg.Rcode)
	}
	if msg.Truncated {
		return fmt.Errorf("Support for DNS requests with high truncated bit not implemented")
	}

	fmt.Println(msg)
	return nil
}
