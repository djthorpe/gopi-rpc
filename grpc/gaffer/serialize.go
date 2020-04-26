/*
  Go Language Raspberry Pi Interface
  (c) Copyright David Thorpe 2016-2020
  All Rights Reserved
  For Licensing and Usage information, please see LICENSE.md
*/

package gaffer

import (
	"fmt"

	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc/v2"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/v2/protobuf/gaffer"
)

////////////////////////////////////////////////////////////////////////////////
// EVENT

type GafferKernelEvent interface {
	State() rpc.GafferState
	Process() rpc.GafferProcess
	Buf() []byte
	Error() error
}

func ErrorToString(err error) string {
	if err == nil {
		return ""
	} else {
		return err.Error()
	}
}

func ProtoFromEvent(event GafferKernelEvent) *pb.KernelProcessEvent {
	return &pb.KernelProcessEvent{
		State:   pb.KernelProcessEvent_State(event.State()),
		Process: ProtoFromProcess(event.Process()),
		Buf:     event.Buf(),
		Error:   ErrorToString(event.Error()),
	}
}

////////////////////////////////////////////////////////////////////////////////
// SERVICE AND SERVICE LIST

type pbservice struct {
	proto  *pb.Service
	fields map[string]bool
}

func NewMutable(service rpc.GafferService) rpc.MutableGafferService {
	return &pbservice{
		proto:  ProtoFromService(service),
		fields: make(map[string]bool),
	}
}

func (this *pbservice) Name() string {
	if this.proto == nil {
		return ""
	} else {
		return this.proto.Name
	}
}

func (this *pbservice) SetName(value string) rpc.MutableGafferService {
	this.proto.Name = value
	this.fields["name"] = true
	return this
}

func (this *pbservice) Sid() uint32 {
	if this.proto == nil {
		return 0
	} else {
		return this.proto.Sid
	}
}

func (this *pbservice) Path() string {
	if this.proto == nil {
		return ""
	} else {
		return this.proto.Path
	}
}

func (this *pbservice) Cwd() string {
	if this.proto == nil {
		return ""
	} else {
		return this.proto.Cwd
	}
}

func (this *pbservice) Args() []string {
	if this.proto == nil {
		return nil
	} else {
		return this.proto.Args
	}
}

func (this *pbservice) SetArgs(value []string) rpc.MutableGafferService {
	this.proto.Args = value
	this.fields["args"] = true
	return this
}

func (this *pbservice) User() string {
	if this.proto == nil {
		return ""
	} else {
		return this.proto.User
	}
}

func (this *pbservice) Group() string {
	if this.proto == nil {
		return ""
	} else {
		return this.proto.Group
	}
}

func (this *pbservice) Enabled() bool {
	if this.proto == nil {
		return false
	} else {
		return this.proto.Enabled
	}
}

func (this *pbservice) SetEnabled(value bool) rpc.MutableGafferService {
	this.proto.Enabled = value
	this.fields["enabled"] = true
	return this
}

func (this *pbservice) String() string {
	return "<GafferService " + fmt.Sprint(this.proto) + ">"
}

func (this *pbservice) Fields() []string {
	fields := make([]string, 0, len(this.fields))
	for field := range this.fields {
		fields = append(fields, field)
	}
	return fields
}

func ProtoToService(proto *pb.Service) rpc.GafferService {
	return &pbservice{proto, nil}
}

func ProtoFromService(service rpc.GafferService) *pb.Service {
	return &pb.Service{
		Name:    service.Name(),
		Sid:     service.Sid(),
		Path:    service.Path(),
		Cwd:     service.Cwd(),
		Args:    service.Args(),
		User:    service.User(),
		Group:   service.Group(),
		Enabled: service.Enabled(),
	}
}

func ProtoFromServiceList(services []rpc.GafferService) *pb.ServiceList {
	if services == nil {
		return nil
	}
	proto := &pb.ServiceList{
		Service: make([]*pb.Service, len(services)),
	}
	for i, service := range services {
		proto.Service[i] = ProtoFromService(service)
	}
	return proto
}

func ProtoFromServiceListOne(service rpc.GafferService) *pb.ServiceList {
	if service == nil {
		return nil
	}
	return &pb.ServiceList{
		Service: []*pb.Service{ProtoFromService(service)},
	}
}

func ProtoFromServiceId(sid uint32) *pb.ServiceId {
	return &pb.ServiceId{
		Sid: sid,
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROCESS

func ProtoFromProcess(process rpc.GafferProcess) *pb.KernelProcess {
	service := process.Service()
	return &pb.KernelProcess{
		Id:      ProtoFromProcessId(process.Id()),
		State:   pb.KernelProcess_State(process.State()),
		Service: ProtoFromService(service),
	}
}

func ProtoToProcess(pb *pb.KernelProcess) rpc.GafferProcess {
	return &protoProcess{pb}
}

func ProtoFromProcessList(process []rpc.GafferProcess) *pb.KernelProcessList {
	if process == nil {
		return nil
	}
	list := &pb.KernelProcessList{
		Process: make([]*pb.KernelProcess, len(process)),
	}
	for i, p := range process {
		list.Process[i] = ProtoFromProcess(p)
	}
	return list
}

func ProtoFromExecutablesList(executables []string) *pb.KernelExecutableList {
	list := &pb.KernelExecutableList{
		Executable: make([]string, len(executables)),
	}
	for i, exec := range executables {
		list.Executable[i] = exec
	}
	return list
}

////////////////////////////////////////////////////////////////////////////////
// GafferProcess interface implemenation

type protoProcess struct {
	pb *pb.KernelProcess
}

func (this *protoProcess) Id() uint32 {
	if this.pb == nil {
		return 0
	} else {
		return this.pb.GetId().Id
	}
}

func (this *protoProcess) Service() rpc.GafferService {
	if this.pb == nil {
		return nil
	} else {
		return ProtoToService(this.pb.Service)
	}
}

func (this *protoProcess) State() rpc.GafferState {
	if this.pb == nil {
		return rpc.GAFFER_STATE_NONE
	} else {
		return rpc.GafferState(this.pb.GetState())
	}
}

func (this *protoProcess) String() string {
	return "<GafferProcess" + fmt.Sprint(this.pb) + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PROCESS ID

func ProtoFromProcessId(id uint32) *pb.ProcessId {
	return &pb.ProcessId{Id: id}
}
