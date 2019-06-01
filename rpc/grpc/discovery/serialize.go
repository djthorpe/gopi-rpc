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
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/discovery"
	ptypes "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// OPAQUE SERVICE RECORD

type servicerecord struct {
	r *pb.ServiceRecord
}

func (this *servicerecord) Name() string {
	if this.r != nil {
		return this.r.Name
	} else {
		return ""
	}
}

func (this *servicerecord) Service() string {
	if this.r != nil {
		return this.r.Service
	} else {
		return ""
	}
}

func (this *servicerecord) Subtype() string {
	if this.r != nil {
		return this.r.Subtype
	} else {
		return ""
	}
}

func (this *servicerecord) Host() string {
	if this.r != nil {
		return this.r.Host
	} else {
		return ""
	}
}

func (this *servicerecord) Port() uint {
	if this.r != nil {
		return uint(this.r.Port)
	} else {
		return 0
	}
}

func (this *servicerecord) Text() []string {
	if this.r != nil {
		return this.r.Txt
	} else {
		return nil
	}
}

func (this *servicerecord) TTL() time.Duration {
	if this.r != nil {
		if d, err := ptypes.Duration(this.r.Ttl); err == nil {
			return d
		}
	}
	return 0
}

func (this *servicerecord) IP4() []net.IP {
	if this.r != nil && this.r.Ip4 != nil {
		ip4 := make([]net.IP, len(this.r.Ip4))
		for i, ip := range this.r.Ip4 {
			ip4[i] = net.ParseIP(ip)
		}
		return ip4
	} else {
		return nil
	}
}

func (this *servicerecord) IP6() []net.IP {
	if this.r != nil && this.r.Ip6 != nil {
		ip6 := make([]net.IP, len(this.r.Ip6))
		for i, ip := range this.r.Ip6 {
			ip6[i] = net.ParseIP(ip)
		}
		return ip6
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// OPAQUE EVENT

type rpcevent struct {
	r *pb.Event
	c gopi.RPCClientConn
}

func (this *rpcevent) Source() gopi.Driver {
	return this.c
}

func (this *rpcevent) Name() string {
	return "RPCEvent"
}

func (this *rpcevent) Type() gopi.RPCEventType {
	if this.r == nil {
		return gopi.RPC_EVENT_NONE
	}
	switch this.r.Type {
	case pb.EventType_EVENT_ADD:
		return gopi.RPC_EVENT_SERVICE_ADDED
	case pb.EventType_EVENT_EXPIRE:
		return gopi.RPC_EVENT_SERVICE_EXPIRED
	case pb.EventType_EVENT_NAME:
		return gopi.RPC_EVENT_SERVICE_NAME
	case pb.EventType_EVENT_REMOVE:
		return gopi.RPC_EVENT_SERVICE_REMOVED
	case pb.EventType_EVENT_CHANGE:
		return gopi.RPC_EVENT_SERVICE_UPDATED
	default:
		return gopi.RPC_EVENT_NONE
	}
}

func (this *rpcevent) ServiceRecord() gopi.RPCServiceRecord {
	if this.r == nil {
		return nil
	} else {
		return protoToServiceRecord(this.r.Service)
	}
}

func (this *rpcevent) String() string {
	return fmt.Sprintf("<protobuf.RPCEvent>{ %v }", this.r)
}

////////////////////////////////////////////////////////////////////////////////
// DISCOVERY TYPE

func protoToDiscoveryType(t pb.DiscoveryType) rpc.DiscoveryType {
	return rpc.DiscoveryType(t)
}

func protoFromDiscoveryType(t rpc.DiscoveryType) pb.DiscoveryType {
	return pb.DiscoveryType(t)
}

////////////////////////////////////////////////////////////////////////////////
// SERVICE RECORDS

func protoFromIP(ips []net.IP) []string {
	reply := make([]string, len(ips))
	for i, ip := range ips {
		reply[i] = ip.String()
	}
	return reply
}

func protoFromServiceRecord(service gopi.RPCServiceRecord) *pb.ServiceRecord {
	if service == nil {
		return nil
	} else {

		return &pb.ServiceRecord{
			Name:    service.Name(),
			Service: service.Service(),
			Port:    uint32(service.Port()),
			Host:    service.Host(),
			Txt:     service.Text(),
			Ttl:     ptypes.DurationProto(service.TTL()),
			Ip4:     protoFromIP(service.IP4()),
			Ip6:     protoFromIP(service.IP6()),
		}
	}
}

func protoToServiceRecord(proto *pb.ServiceRecord) gopi.RPCServiceRecord {
	return &servicerecord{proto}
}

func protoFromServiceRecords(records []gopi.RPCServiceRecord) []*pb.ServiceRecord {
	if records == nil {
		return nil
	}
	reply := make([]*pb.ServiceRecord, len(records))
	for i, record := range records {
		reply[i] = protoFromServiceRecord(record)
	}
	return reply
}

func protoToServiceRecords(proto []*pb.ServiceRecord) []gopi.RPCServiceRecord {
	if proto == nil {
		return nil
	}
	records := make([]gopi.RPCServiceRecord, len(proto))
	for i, record := range proto {
		records[i] = protoToServiceRecord(record)
	}
	return records
}

////////////////////////////////////////////////////////////////////////////////
// RPCEVENT

func protoFromEventType(type_ gopi.RPCEventType) pb.EventType {
	switch type_ {
	case gopi.RPC_EVENT_SERVICE_ADDED:
		return pb.EventType_EVENT_ADD
	case gopi.RPC_EVENT_SERVICE_EXPIRED:
		return pb.EventType_EVENT_EXPIRE
	case gopi.RPC_EVENT_SERVICE_NAME:
		return pb.EventType_EVENT_NAME
	case gopi.RPC_EVENT_SERVICE_REMOVED:
		return pb.EventType_EVENT_REMOVE
	case gopi.RPC_EVENT_SERVICE_UPDATED:
		return pb.EventType_EVENT_CHANGE
	default:
		return pb.EventType_EVENT_NONE
	}
}

func protoFromEvent(event gopi.RPCEvent) *pb.Event {
	if event == nil {
		return nil
	}
	return &pb.Event{
		Type:    protoFromEventType(event.Type()),
		Service: protoFromServiceRecord(event.ServiceRecord()),
	}
}

func protoToEvent(proto *pb.Event, conn gopi.RPCClientConn) gopi.RPCEvent {
	if proto == nil {
		return nil
	}
	return &rpcevent{proto, conn}
}
