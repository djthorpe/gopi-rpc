/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2019
  All Rights Reserved

  Documentation http://djthorpe.github.io/gopi/
  For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"fmt"
	"net"
	"strings"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// ServiceRecord
type ServiceRecord struct {
	key     string
	name    string
	host    string
	service string
	port    uint16
	txt     []string
	ipv4    []net.IP
	ipv6    []net.IP
	ts      time.Time
	ttl     time.Duration
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *ServiceRecord) String() string {
	p := make([]string, 0, 5)
	if s.name != "" {
		p = append(p, fmt.Sprintf("name=\"%v\"", s.name))
	}
	if s.service != "" {
		p = append(p, fmt.Sprintf("type=%v", s.service))
	}
	if s.port > 0 {
		p = append(p, fmt.Sprintf("port=%v", s.port))
	}
	if s.host != "" {
		p = append(p, fmt.Sprintf("host=%v", s.host))
	}
	if len(s.ipv4) > 0 {
		p = append(p, fmt.Sprintf("ip4=%v", s.ipv4))
	}
	if len(s.ipv6) > 0 {
		p = append(p, fmt.Sprintf("ip6=%v", s.ipv6))
	}
	if s.ttl > 0 {
		p = append(p, fmt.Sprintf("ttl=%v", s.ttl))
	}
	if len(s.txt) > 0 {
		p = append(p, fmt.Sprintf("txt=%v", s.txt))
	}
	return fmt.Sprintf("<gopi.RPCServiceRecord>{ %v }", strings.Join(p, " "))
}
