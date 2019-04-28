/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package discovery

import (
	"github.com/golang/protobuf/ptypes"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/discovery"
)

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
		}
	}
}

func protoToServiceRecord(proto *pb.ServiceRecord) gopi.RPCServiceRecord {
	if proto == nil {
		return nil
	} else if ttl, err := ptypes.Duration(proto.Ttl); err != nil {
		return nil
	} else {
		r := rpc.NewServiceRecord()
		r.Name_ = proto.Name
		r.Service_ = proto.Service
		r.Host_ = proto.Host
		r.Port_ = uint(proto.Port)
		r.Txt_ = proto.Txt
		r.Ttl_ = &rpc.Duration{ttl}
		return r
	}
}

//TODO: repeated string ip4 = 6;
//TODO: repeated string ip6 = 7;

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
