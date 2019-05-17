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

type Client struct {
	pb.GafferClient
	conn gopi.RPCClientConn
}

////////////////////////////////////////////////////////////////////////////////
// NEW

func NewGafferClient(conn gopi.RPCClientConn) gopi.RPCClient {
	return &Client{pb.NewGafferClient(conn.(grpc.GRPCClientConn).GRPCConn()), conn}
}

func (this *Client) NewContext() context.Context {
	if this.conn.Timeout() == 0 {
		return context.Background()
	} else {
		ctx, _ := context.WithTimeout(context.Background(), this.conn.Timeout())
		return ctx
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (this *Client) Conn() gopi.RPCClientConn {
	return this.conn
}

////////////////////////////////////////////////////////////////////////////////
// CALLS

func (this *Client) Ping() error {
	this.conn.Lock()
	defer this.conn.Unlock()

	// Perform ping
	if _, err := this.GafferClient.Ping(this.NewContext(), &empty.Empty{}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Client) ListExecutables() ([]string, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.ListExecutables(this.NewContext(), &empty.Empty{}); err != nil {
		return nil, err
	} else {
		return reply.Path, nil
	}
}

func (this *Client) ListServices() ([]rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.ListServices(this.NewContext(), &pb.RequestFilter{
		Type: pb.RequestFilter_NONE,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoServiceArray(reply.Service), nil
	}
}

func (this *Client) ListServicesForGroup(group string) ([]rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.ListServices(this.NewContext(), &pb.RequestFilter{
		Type:  pb.RequestFilter_GROUP,
		Value: group,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoServiceArray(reply.Service), nil
	}
}

func (this *Client) GetService(service string) (rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.ListServices(this.NewContext(), &pb.RequestFilter{
		Type:  pb.RequestFilter_SERVICE,
		Value: service,
	}); err != nil {
		return nil, err
	} else if len(reply.Service) == 0 {
		return nil, gopi.ErrNotFound
	} else {
		return fromProtoService(reply.Service[0]), nil
	}
}

func (this *Client) ListGroups() ([]rpc.GafferServiceGroup, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.ListGroups(this.NewContext(), &pb.RequestFilter{
		Type: pb.RequestFilter_NONE,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoGroupArray(reply.Group), nil
	}
}

func (this *Client) ListGroupsForService(service string) ([]rpc.GafferServiceGroup, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.ListGroups(this.NewContext(), &pb.RequestFilter{
		Type:  pb.RequestFilter_SERVICE,
		Value: service,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoGroupArray(reply.Group), nil
	}
}

func (this *Client) GetGroup(group string) (rpc.GafferServiceGroup, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.ListGroups(this.NewContext(), &pb.RequestFilter{
		Type:  pb.RequestFilter_GROUP,
		Value: group,
	}); err != nil {
		return nil, err
	} else if len(reply.Group) == 0 {
		return nil, gopi.ErrNotFound
	} else {
		return fromProtoGroup(reply.Group[0]), nil
	}
}

func (this *Client) ListInstances() ([]rpc.GafferServiceInstance, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.ListInstances(this.NewContext(), &pb.RequestFilter{}); err != nil {
		return nil, err
	} else {
		return fromProtoInstanceArray(reply.Instance), nil
	}
}

func (this *Client) AddServiceForPath(path string, groups []string) (rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.AddService(this.NewContext(), &pb.ServiceRequest{
		Name:   path,
		Groups: groups,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoService(reply), nil
	}
}

func (this *Client) AddGroupForName(name string) (rpc.GafferServiceGroup, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.AddGroup(this.NewContext(), &pb.NameRequest{
		Name: name,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoGroup(reply), nil
	}
}

func (this *Client) RemoveServiceForName(name string) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	if _, err := this.GafferClient.RemoveService(this.NewContext(), &pb.NameRequest{
		Name: name,
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Client) RemoveGroupForName(name string) error {
	this.conn.Lock()
	defer this.conn.Unlock()

	if _, err := this.GafferClient.RemoveGroup(this.NewContext(), &pb.NameRequest{
		Name: name,
	}); err != nil {
		return err
	} else {
		return nil
	}
}

func (this *Client) GetInstanceId() (uint32, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.GetInstanceId(this.NewContext(), &empty.Empty{}); err != nil {
		return 0, err
	} else {
		return reply.Id, nil
	}
}

func (this *Client) StartInstance(service string, id uint32) (rpc.GafferServiceInstance, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.StartInstance(this.NewContext(), &pb.StartInstanceRequest{
		Id:      id,
		Service: service,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoInstance(reply), nil
	}
}

func (this *Client) StopInstance(id uint32) (rpc.GafferServiceInstance, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.StopInstance(this.NewContext(), &pb.InstanceId{
		Id: id,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoInstance(reply), nil
	}
}

func (this *Client) SetFlagsForService(service string, tuples rpc.Tuples) (rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.SetServiceFlags(this.NewContext(), &pb.SetTuplesRequest{
		Name:   service,
		Tuples: toProtoTuples(tuples),
	}); err != nil {
		return nil, err
	} else {
		return fromProtoService(reply), nil
	}
}

func (this *Client) SetFlagsForGroup(group string, tuples rpc.Tuples) (rpc.GafferServiceGroup, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.SetGroupFlags(this.NewContext(), &pb.SetTuplesRequest{
		Name:   group,
		Tuples: toProtoTuples(tuples),
	}); err != nil {
		return nil, err
	} else {
		return fromProtoGroup(reply), nil
	}
}

func (this *Client) SetEnvForGroup(group string, tuples rpc.Tuples) (rpc.GafferServiceGroup, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.SetGroupEnv(this.NewContext(), &pb.SetTuplesRequest{
		Name:   group,
		Tuples: toProtoTuples(tuples),
	}); err != nil {
		return nil, err
	} else {
		return fromProtoGroup(reply), nil
	}
}

func (this *Client) SetServiceGroups(service string, groups []string) (rpc.GafferService, error) {
	this.conn.Lock()
	defer this.conn.Unlock()

	if reply, err := this.GafferClient.SetServiceParameters(this.NewContext(), &pb.ServiceRequest{
		Name:   service,
		Groups: groups,
	}); err != nil {
		return nil, err
	} else {
		return fromProtoService(reply), nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Client) String() string {
	return fmt.Sprintf("<rpc.service.gaffer.Client>{ conn=%v }", this.conn)
}
