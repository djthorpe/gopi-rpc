/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"time"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/gaffer"
	ptypes "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type pb_service struct {
	pb *pb.Service
}

type pb_group struct {
	pb *pb.Group
}

type pb_instance struct {
	pb *pb.Instance
}

type pb_event struct {
	pb *pb.GafferEvent
}

////////////////////////////////////////////////////////////////////////////////
// SERVICES

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
		Flags:         toProtoTuples(service.Flags()),
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
// GROUPS

func toProtoFromGroup(group rpc.GafferServiceGroup) *pb.Group {
	if group == nil {
		return nil
	}
	return &pb.Group{
		Name:  group.Name(),
		Flags: toProtoTuples(group.Flags()),
		Env:   toProtoTuples(group.Env()),
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

func fromProtoGroup(group *pb.Group) rpc.GafferServiceGroup {
	return &pb_group{group}
}

func fromProtoGroupArray(groups []*pb.Group) []rpc.GafferServiceGroup {
	if groups == nil {
		return nil
	}
	groups_ := make([]rpc.GafferServiceGroup, len(groups))
	for i, group := range groups {
		groups_[i] = fromProtoGroup(group)
	}
	return groups_
}

////////////////////////////////////////////////////////////////////////////////
// INSTANCES

func toProtoFromInstance(instance rpc.GafferServiceInstance) *pb.Instance {
	if instance == nil {
		return nil
	}
	if start_ts, err := ptypes.TimestampProto(instance.Start()); err != nil {
		return nil
	} else if stop_ts, err := ptypes.TimestampProto(instance.Stop()); err != nil {
		return nil
	} else {
		return &pb.Instance{
			Id:       instance.Id(),
			Service:  toProtoFromService(instance.Service()),
			Flags:    toProtoTuples(instance.Flags()),
			Env:      toProtoTuples(instance.Env()),
			StartTs:  start_ts,
			StopTs:   stop_ts,
			ExitCode: instance.ExitCode(),
		}
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

func fromProtoInstance(instance *pb.Instance) rpc.GafferServiceInstance {
	return &pb_instance{instance}
}

func fromProtoInstanceArray(instances []*pb.Instance) []rpc.GafferServiceInstance {
	if instances == nil {
		return nil
	}
	instances_ := make([]rpc.GafferServiceInstance, len(instances))
	for i, instance := range instances {
		instances_[i] = fromProtoInstance(instance)
	}
	return instances_
}

////////////////////////////////////////////////////////////////////////////////
// EVENTS

func toProtoEvent(evt rpc.GafferEvent) *pb.GafferEvent {
	if evt == nil {
		return nil
	}
	return &pb.GafferEvent{
		Type:     pb.GafferEvent_Type(evt.Type()),
		Service:  toProtoFromService(evt.Service()),
		Group:    toProtoFromGroup(evt.Group()),
		Instance: toProtoFromInstance(evt.Instance()),
		Data:     evt.Data(),
		Ts:       ptypes.TimestampNow(),
	}
}

func fromProtoEvent(evt *pb.GafferEvent) rpc.GafferEvent {
	if evt == nil {
		return nil
	}
	return &pb_event{evt}
}

////////////////////////////////////////////////////////////////////////////////
// TUPLES

func toProtoTuples(tuples rpc.Tuples) *pb.Tuples {
	proto := &pb.Tuples{
		Tuples: make([]*pb.Tuple, tuples.Len()),
	}
	for i, key := range tuples.Keys() {
		proto.Tuples[i] = &pb.Tuple{
			Key:   key,
			Value: tuples.StringForKey(key),
		}
	}
	return proto
}

func fromProtoTuples(proto *pb.Tuples) rpc.Tuples {
	if proto == nil {
		return rpc.Tuples{}
	}

	// Copy over the tuples
	tuples := rpc.Tuples{}
	for _, tuple := range proto.Tuples {
		if err := tuples.SetStringForKey(tuple.Key, tuple.Value); err != nil {
			return rpc.Tuples{}
		}
	}

	// Return tuples
	return tuples
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

func (this *pb_service) Flags() rpc.Tuples {
	if this.pb == nil {
		return rpc.Tuples{}
	} else {
		return fromProtoTuples(this.pb.Flags)
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

////////////////////////////////////////////////////////////////////////////////
// GROUP IMPLEMENTATION

func (this *pb_group) Name() string {
	if this.pb == nil {
		return ""
	} else {
		return this.pb.Name
	}
}

func (this *pb_group) Flags() rpc.Tuples {
	if this.pb == nil {
		return rpc.Tuples{}
	} else {
		return fromProtoTuples(this.pb.Flags)
	}
}

func (this *pb_group) Env() rpc.Tuples {
	if this.pb == nil {
		return rpc.Tuples{}
	} else {
		return fromProtoTuples(this.pb.Env)
	}
}

////////////////////////////////////////////////////////////////////////////////
// INSTANCE IMPLEMENTATION

func (this *pb_instance) Id() uint32 {
	if this.pb == nil {
		return 0
	} else {
		return this.pb.Id
	}
}

func (this *pb_instance) Service() rpc.GafferService {
	if this.pb == nil {
		return nil
	} else {
		return fromProtoService(this.pb.Service)
	}
}

func (this *pb_instance) Flags() rpc.Tuples {
	if this.pb == nil {
		return rpc.Tuples{}
	} else {
		return fromProtoTuples(this.pb.Flags)
	}
}

func (this *pb_instance) Env() rpc.Tuples {
	if this.pb == nil {
		return rpc.Tuples{}
	} else {
		return fromProtoTuples(this.pb.Env)
	}
}

func (this *pb_instance) Start() time.Time {
	if this.pb == nil {
		return time.Time{}
	} else if ts, err := ptypes.Timestamp(this.pb.StartTs); err != nil {
		return time.Time{}
	} else {
		return ts
	}
}

func (this *pb_instance) Stop() time.Time {
	if this.pb == nil {
		return time.Time{}
	} else if ts, err := ptypes.Timestamp(this.pb.StopTs); err != nil {
		return time.Time{}
	} else {
		return ts
	}
}

func (this *pb_instance) ExitCode() int64 {
	if this.pb == nil {
		return 0
	} else {
		return this.pb.ExitCode
	}
}

////////////////////////////////////////////////////////////////////////////////
// EVENT IMPLEMENTATION

func (this *pb_event) Source() gopi.Driver {
	return nil
}

func (this *pb_event) Name() string {
	return "GafferEvent"
}

func (this *pb_event) Type() rpc.GafferEventType {
	if this.pb == nil {
		return rpc.GAFFER_EVENT_NONE
	} else {
		return rpc.GafferEventType(this.pb.Type)
	}
}

func (this *pb_event) Service() rpc.GafferService {
	if this.pb == nil {
		return nil
	} else if this.pb.Service != nil {
		return fromProtoService(this.pb.Service)
	} else if this.pb.Instance.Service != nil {
		return fromProtoService(this.pb.Instance.Service)
	} else {
		return nil
	}
}

func (this *pb_event) Group() rpc.GafferServiceGroup {
	if this.pb == nil {
		return nil
	} else {
		return fromProtoGroup(this.pb.Group)
	}
}

func (this *pb_event) Instance() rpc.GafferServiceInstance {
	if this.pb == nil {
		return nil
	} else {
		return fromProtoInstance(this.pb.Instance)
	}
}

func (this *pb_event) Data() []byte {
	if this.pb == nil {
		return nil
	} else {
		return this.pb.Data
	}
}
