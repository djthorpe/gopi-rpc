/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"github.com/golang/protobuf/ptypes"
	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/gaffer"
)

////////////////////////////////////////////////////////////////////////////////
// Services

func toProtoFromService(service rpc.GafferService) *pb.Service {
	if service == nil {
		return nil
	}
	return &pb.Service{
		Name:          service.Name(),
		Path:          service.Path(),
		Groups:        service.Groups(),
		Mode:          pb.Service_ServiceMode(service.Mode()),
		InstanceCount: uint32(service.InstanceCount()),
		RunTime:       ptypes.DurationProto(service.RunTime()),
		IdleTime:      ptypes.DurationProto(service.IdleTime()),
		Flags:         service.Flags(),
	}
}

func toProtoFromServiceArray(services []rpc.GafferService, filter func(rpc.GafferService) bool) []*pb.Service {
	if services == nil {
		return nil
	}
	services_ := make([]*pb.Service, 0, len(services))
	for _, service := range services {
		if filter == nil || filter(service) {
			services_ = append(services_, toProtoFromService(service))
		}
	}
	return services_
}

////////////////////////////////////////////////////////////////////////////////
// Groups

func toProtoFromGroup(group rpc.GafferServiceGroup) *pb.Group {
	if group == nil {
		return nil
	}
	return &pb.Group{
		Name:  group.Name(),
		Flags: group.Flags(),
		Env:   group.Env(),
	}
}

func toProtoFromGroupArray(groups []rpc.GafferServiceGroup, filter func(rpc.GafferServiceGroup) bool) []*pb.Group {
	if groups == nil {
		return nil
	}
	groups_ := make([]*pb.Group, 0, len(groups))
	for _, group := range groups {
		if filter == nil || filter(group) {
			groups_ = append(groups_, toProtoFromGroup(group))
		}
	}
	return groups_
}

////////////////////////////////////////////////////////////////////////////////
// Instances

func toProtoFromInstance(instance rpc.GafferServiceInstance) *pb.Instance {
	if instance == nil {
		return nil
	}
	return &pb.Instance{
		Id:      instance.Id(),
		Service: toProtoFromService(instance.Service()),
	}
}
