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
// SERVICE

func ProtoToService(proto *pb.KernelService) rpc.GafferService {
	if proto == nil {
		return rpc.GafferService{}
	} else {
		return rpc.GafferService{
			Name:  proto.GetName(),
			Path:  proto.GetPath(),
			Args:  proto.GetArgs(),
			Cwd:   proto.GetCwd(),
			User:  proto.GetUser(),
			Group: proto.GetGroup(),
			Sid:   proto.GetSid(),
		}
	}
}

func ProtoFromService(service rpc.GafferService) *pb.KernelService {
	return &pb.KernelService{
		Name:  service.Name,
		Path:  service.Path,
		Args:  service.Args,
		Cwd:   service.Cwd,
		User:  service.User,
		Group: service.Group,
		Sid:   service.Sid,
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROCESS

func ProtoFromProcess(process rpc.GafferProcess) *pb.KernelProcess {
	service := process.Service()
	return &pb.KernelProcess{
		Id:      ProtoFromProcessId(process.Id(), service.Sid),
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
		return rpc.GafferService{}
	} else {
		return ProtoToService(this.pb.GetService())
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

func ProtoFromProcessId(id, sid uint32) *pb.KernelProcessId {
	return &pb.KernelProcessId{Id: id, Sid: sid}
}
