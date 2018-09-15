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
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ServiceRecord struct {
	Timestamp     time.Time
	Key           string
	Name          string
	Host          string
	ServiceDomain string
	Port          uint16
	TTL           time.Duration
	TXT           []string
	IPv4          []net.IP
	IPv6          []net.IP
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *ServiceRecord) complete() bool {
	if this.Key == "" {
		return false
	}
	return true
}

func (this *ServiceRecord) String() string {
	parts := make([]string, 0)
	if this.Key != "" {
		parts = append(parts, fmt.Sprintf("Key='%v'", this.Key))
	}
	if this.Name != "" {
		parts = append(parts, fmt.Sprintf("Name='%v'", this.Name))
	}
	if this.ServiceDomain != "" {
		parts = append(parts, fmt.Sprintf("ServiceDomain='%v'", this.ServiceDomain))
	}
	if this.Host != "" {
		parts = append(parts, fmt.Sprintf("Host='%v'", this.Host))
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
	return fmt.Sprintf("<ServiceRecord>{ %v ts=%v }", strings.Join(parts, " "), this.Timestamp.Format(time.Kitchen))
}
