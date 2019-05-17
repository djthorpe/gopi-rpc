/*
	Gaffer: Microservice Manager
	(c) Copyright David Thorpe 2019
	All Rights Reserved

	For Licensing and Usage information, please see LICENSE
*/

package gaffer

import (
	"context"
	"fmt"
	"strings"

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"
	event "github.com/djthorpe/gopi/util/event"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/gaffer"
	empty "github.com/golang/protobuf/ptypes/empty"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Service struct {
	Server gopi.RPCServer
	Gaffer rpc.Gaffer
}

type service struct {
	log    gopi.Logger
	gaffer rpc.Gaffer

	event.Tasks
}

////////////////////////////////////////////////////////////////////////////////
// OPEN AND CLOSE

// Open the server
func (config Service) Open(log gopi.Logger) (gopi.Driver, error) {
	log.Debug("<grpc.service.gaffer>Open{ server=%v gaffer=%v }", config.Server, config.Gaffer)

	this := new(service)
	this.log = log
	this.gaffer = config.Gaffer

	if this.gaffer == nil {
		return nil, gopi.ErrBadParameter
	}

	// Register service with GRPC server
	pb.RegisterGafferServer(config.Server.(grpc.GRPCServer).GRPCServer(), this)

	// Start background task which reports on all events (for debugging)
	this.Tasks.Start(this.EventTask)

	// Success
	return this, nil
}

func (this *service) Close() error {
	this.log.Debug("<grpc.service.gaffer>Close{ gaffer=%v }", this.gaffer)

	// Stop background tasks
	if err := this.Tasks.Close(); err != nil {
		return err
	}

	// Release resources
	this.gaffer = nil

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// RPCService implementation

func (this *service) CancelRequests() error {
	// No need to cancel any requests since none are streaming
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Stringify

func (this *service) String() string {
	return fmt.Sprintf("grpc.service.gaffer{}")
}

////////////////////////////////////////////////////////////////////////////////
// RPC Methods

func (this *service) Ping(context.Context, *empty.Empty) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.gaffer.Ping>{ }")
	return &empty.Empty{}, nil
}

// ListExecutables returns a list of executables which can be made into services
func (this *service) ListExecutables(context.Context, *empty.Empty) (*pb.ListExecutablesReply, error) {
	this.log.Debug("<grpc.service.gaffer.ListExecutables>{ }")

	recursive := true
	executables := this.gaffer.GetExecutables(recursive)

	return &pb.ListExecutablesReply{
		Path: executables,
	}, nil
}

// List services
func (this *service) ListServices(_ context.Context, req *pb.RequestFilter) (*pb.ListServicesReply, error) {
	this.log.Debug("<grpc.service.gaffer.ListServices>{ req=%v }", req)

	// Where one service is required
	if req.Type == pb.RequestFilter_SERVICE {
		service := this.gaffer.GetServiceForName(req.Value)
		if service == nil {
			return &pb.ListServicesReply{
				Service: []*pb.Service{},
			}, nil
		} else {
			return &pb.ListServicesReply{
				Service: []*pb.Service{toProtoFromService(service)},
			}, nil
		}
	}

	// Obtain all services
	services := this.gaffer.GetServices()
	if len(services) == 0 {
		return &pb.ListServicesReply{}, nil
	}

	if req.Type == pb.RequestFilter_GROUP {
		// Where services should be returned filtered by group name
		return &pb.ListServicesReply{
			Service: toProtoFromServiceArray(services, func(s rpc.GafferService) bool {
				return s.IsMemberOfGroup(req.Value)
			}),
		}, nil
	} else if req.Type == pb.RequestFilter_NONE {
		// Where all services should be returned (no filtering)
		return &pb.ListServicesReply{
			Service: toProtoFromServiceArray(services, nil),
		}, nil
	} else {
		return nil, gopi.ErrNotImplemented
	}
}

// List groups
func (this *service) ListGroups(_ context.Context, req *pb.RequestFilter) (*pb.ListGroupsReply, error) {
	this.log.Debug("<grpc.service.gaffer.ListGroups>{ req=%v }", req)

	// Where one group is required
	if req.Type == pb.RequestFilter_GROUP {
		groups := this.gaffer.GetGroupsForNames([]string{req.Value})
		if len(groups) == 0 {
			return &pb.ListGroupsReply{
				Group: []*pb.Group{},
			}, nil
		} else {
			return &pb.ListGroupsReply{
				Group: toProtoFromGroupArray(groups, nil),
			}, nil
		}
	}

	// Where groups of a service are required
	if req.Type == pb.RequestFilter_SERVICE {
		if service := this.gaffer.GetServiceForName(req.Value); service == nil {
			return nil, gopi.ErrNotFound
		} else {
			groups := this.gaffer.GetGroupsForNames(service.Groups())
			return &pb.ListGroupsReply{
				Group: toProtoFromGroupArray(groups, nil),
			}, nil
		}
	}

	// Obtain all groups
	if req.Type == pb.RequestFilter_NONE {
		groups := this.gaffer.GetGroups()
		return &pb.ListGroupsReply{
			Group: toProtoFromGroupArray(groups, nil),
		}, nil
	}

	return nil, gopi.ErrNotImplemented
}

// List instances
func (this *service) ListInstances(_ context.Context, req *pb.RequestFilter) (*pb.ListInstancesReply, error) {
	this.log.Debug("<grpc.service.gaffer.ListInstances>{ req=%v }", req)

	// Get instances
	instances := this.gaffer.GetInstances()

	// Obtain all instances
	if req.Type == pb.RequestFilter_NONE {
		return &pb.ListInstancesReply{
			Instance: toProtoFromInstanceArray(instances, nil),
		}, nil
	}

	// Obtain instances for a service
	if req.Type == pb.RequestFilter_SERVICE {
		if service := this.gaffer.GetServiceForName(req.Value); service == nil {
			return nil, gopi.ErrNotFound
		} else {
			return &pb.ListInstancesReply{
				Instance: toProtoFromInstanceArray(instances, func(i rpc.GafferServiceInstance) bool {
					return i.Service() == service
				}),
			}, nil
		}
	}

	// Obtain instances for services in a particular group
	if req.Type == pb.RequestFilter_SERVICE {
		return &pb.ListInstancesReply{
			Instance: toProtoFromInstanceArray(instances, func(i rpc.GafferServiceInstance) bool {
				return i.Service().IsMemberOfGroup(req.Value)
			}),
		}, nil
	}

	// Filter not implemented
	return nil, gopi.ErrNotImplemented
}

// Add a service
func (this *service) AddService(_ context.Context, req *pb.AddServiceRequest) (*pb.Service, error) {
	this.log.Debug("<grpc.service.gaffer.AddService>{ req=%v }", req)

	if service, err := this.gaffer.AddServiceForPath(req.Path); err != nil {
		return nil, err
	} else {
		return toProtoFromService(service), nil
	}
}

// Add a group
func (this *service) AddGroup(_ context.Context, req *pb.NameRequest) (*pb.Group, error) {
	this.log.Debug("<grpc.service.gaffer.AddGroup>{ req=%v }", req)

	if group, err := this.gaffer.AddGroupForName(req.Name); err != nil {
		return nil, err
	} else {
		return toProtoFromGroup(group), nil
	}
}

// Remove a service
func (this *service) RemoveService(_ context.Context, req *pb.NameRequest) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.gaffer.RemoveService>{ req=%v }", req)

	if err := this.gaffer.RemoveServiceForName(req.Name); err != nil {
		return nil, err
	} else {
		return &empty.Empty{}, nil
	}
}

// Remove a group
func (this *service) RemoveGroup(_ context.Context, req *pb.NameRequest) (*empty.Empty, error) {
	this.log.Debug("<grpc.service.gaffer.RemoveGroup>{ req=%v }", req)

	if err := this.gaffer.RemoveGroupForName(req.Name); err != nil {
		return nil, err
	} else {
		return &empty.Empty{}, nil
	}
}

// Get an Instance ID
func (this *service) GetInstanceId(context.Context, *empty.Empty) (*pb.InstanceId, error) {
	this.log.Debug("<grpc.service.gaffer.GetInstanceId>{}")

	if id := this.gaffer.GenerateInstanceId(); id == 0 {
		return nil, gopi.ErrOutOfOrder
	} else {
		return &pb.InstanceId{Id: id}, nil
	}
}

// Start an Instance given service name and ID
func (this *service) StartInstance(_ context.Context, req *pb.StartInstanceRequest) (*pb.Instance, error) {
	this.log.Debug("<grpc.service.gaffer.StartInstance>{ req=%v }", req)

	if instance, err := this.gaffer.StartInstanceForServiceName(req.Service, req.Id); err != nil {
		return nil, err
	} else {
		return toProtoFromInstance(instance), nil
	}
}

func (this *service) StopInstance(_ context.Context, req *pb.InstanceId) (*pb.Instance, error) {
	this.log.Debug("<grpc.service.gaffer.StopInstance>{ req=%v }", req)

	if instance := this.gaffer.GetInstanceForId(req.Id); instance == nil {
		return nil, gopi.ErrNotFound
	} else if err := this.gaffer.StopInstanceForId(req.Id); err != nil {
		return nil, err
	} else {
		return toProtoFromInstance(instance), nil
	}

}

// Set group flags
func (this *service) SetGroupFlags(_ context.Context, req *pb.SetTuplesRequest) (*pb.Group, error) {
	this.log.Debug("<grpc.service.gaffer.SetGroupFlags>{ req=%v }", req)

	if err := this.gaffer.SetGroupFlagsForName(req.Name, fromProtoTuples(req.Tuples)); err != nil {
		return nil, err
	} else if groups := this.gaffer.GetGroupsForNames([]string{req.Name}); len(groups) == 0 {
		return nil, gopi.ErrNotFound
	} else if len(groups) > 1 {
		return nil, gopi.ErrAppError
	} else {
		return toProtoFromGroup(groups[0]), nil
	}
}

// Set group env
func (this *service) SetGroupEnv(_ context.Context, req *pb.SetTuplesRequest) (*pb.Group, error) {
	this.log.Debug("<grpc.service.gaffer.SetGroupEnv>{ req=%v }", req)

	if err := this.gaffer.SetGroupEnvForName(req.Name, fromProtoTuples(req.Tuples)); err != nil {
		return nil, err
	} else if groups := this.gaffer.GetGroupsForNames([]string{req.Name}); len(groups) == 0 {
		return nil, gopi.ErrNotFound
	} else if len(groups) > 1 {
		return nil, gopi.ErrAppError
	} else {
		return toProtoFromGroup(groups[0]), nil
	}
}

// Set service flags
func (this *service) SetServiceFlags(_ context.Context, req *pb.SetTuplesRequest) (*pb.Service, error) {
	this.log.Debug("<grpc.service.gaffer.SetServiceFlags>{ req=%v }", req)

	if err := this.gaffer.SetServiceFlagsForName(req.Name, fromProtoTuples(req.Tuples)); err != nil {
		return nil, err
	} else if service := this.gaffer.GetServiceForName(req.Name); service == nil {
		return nil, gopi.ErrNotFound
	} else {
		return toProtoFromService(service), nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// BACKGROUND TASKS

func (this *service) EventTask(start chan<- event.Signal, stop <-chan event.Signal) error {
	start <- gopi.DONE
	events := this.gaffer.Subscribe()
FOR_LOOP:
	for {
		select {
		case evt := <-events:
			if evt_, ok := evt.(rpc.GafferEvent); ok {
				this.EventPrint(evt_)
			} else {
				this.log.Warn("Ignoring: %v", evt)
			}
		case <-stop:
			break FOR_LOOP
		}
	}

	this.gaffer.Unsubscribe(events)

	// Success
	return nil
}

func (this *service) EventPrint(evt rpc.GafferEvent) {
	switch evt.Type() {
	case rpc.GAFFER_EVENT_LOG_STDERR, rpc.GAFFER_EVENT_LOG_STDOUT:
		line := strings.Trim(string(evt.Data()), "\n")
		this.log.Info("%v[%v]: %v", evt.Service().Name(), evt.Instance().Id(), line)
	default:
		this.log.Info("%v", evt)
	}
}
