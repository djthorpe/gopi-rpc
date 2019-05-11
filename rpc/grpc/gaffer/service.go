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

	// Frameworks
	gopi "github.com/djthorpe/gopi"
	rpc "github.com/djthorpe/gopi-rpc"
	grpc "github.com/djthorpe/gopi-rpc/sys/grpc"

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

	// Success
	return this, nil
}

func (this *service) Close() error {
	this.log.Debug("<grpc.service.gaffer>Close{ gaffer=%v }", this.gaffer)

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
func (this *service) AddGroup(_ context.Context, req *pb.AddGroupRequest) (*pb.Group, error) {
	this.log.Debug("<grpc.service.gaffer.AddGroup>{ req=%v }", req)

	if group, err := this.gaffer.AddGroupForName(req.Name); err != nil {
		return nil, err
	} else {
		return toProtoFromGroup(group), nil
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
