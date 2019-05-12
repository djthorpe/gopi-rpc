/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/gaffer"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type pb_service struct {
	pb *pb.Service
}

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

func fromProtoService(service *pb.Service) rpc.GafferService {
	return &pb_service{service}
}

func fromProtoServiceArray(services []*pb.Service) []rpc.GafferService {
	if services == nil {
		return nil
	}
	services_ := make([]rpc.GafferService, len(services))
	for i, service := range services {
		services_[i] = fromProtoService(service)
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
		Flags:   instance.Flags(),
		Env:     instance.Env(),
	}
}

func toProtoFromInstanceArray(instances []rpc.GafferServiceInstance, filter func(rpc.GafferServiceInstance) bool) []*pb.Instance {
	if instances == nil {
		return nil
	}
	instances_ := make([]*pb.Instance, 0, len(instances))
	for _, instance := range instances {
		if filter == nil || filter(instance) {
			instances_ = append(instances_, toProtoFromInstance(instance))
		}
	}
	return instances_
}

////////////////////////////////////////////////////////////////////////////////
// SERVICE IMPLEMENTATION

func (this *pb_service) Name() string {
	if this.pb == nil {
		return ""
	} else {
		return this.pb.Name
	}
}

func (this *pb_service) Path() string {
	if this.pb == nil {
		return ""
	} else {
		return this.pb.Path
	}
}

func (this *pb_service) Groups() []string {
	if this.pb == nil {
		return nil
	} else {
		return this.pb.Groups
	}
}

func (this *pb_service) Mode() rpc.GafferServiceMode {
	if this.pb == nil {
		return rpc.GAFFER_MODE_NONE
	} else {
		return rpc.GafferServiceMode(this.pb.Mode)
	}
}

func (this *pb_service) InstanceCount() uint {
	if this.pb == nil {
		return 0
	} else {
		return uint(this.pb.InstanceCount)
	}
}

func (this *pb_service) RunTime() time.Duration {
	if this.pb == nil {
		return 0
	} else if duration, err := ptypes.Duration(this.pb.RunTime); err != nil {
		return 0
	} else {
		return duration
	}
}

func (this *pb_service) IdleTime() time.Duration {
	if this.pb == nil {
		return 0
	} else if duration, err := ptypes.Duration(this.pb.IdleTime); err != nil {
		return 0
	} else {
		return duration
	}
}

func (this *pb_service) SetFlags(map[string]string) error {
	return gopi.ErrNotImplemented
}

func (this *pb_service) Flags() []string {
	if this.pb == nil {
		return nil
	} else {
		return this.pb.Flags
	}
}

func (this *pb_service) IsMemberOfGroup(group string) bool {
	if this.pb == nil {
		return false
	} else {
		for _, group_ := range this.pb.Groups {
			if group == group_ {
				return true
			}
		}
		return false
	}
}
