/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	// Frameworks
	rpc "github.com/djthorpe/gopi-rpc"

	// Protocol buffers
	pb "github.com/djthorpe/gopi-rpc/rpc/protobuf/googlecast"
	ptypes "github.com/golang/protobuf/ptypes"
)

////////////////////////////////////////////////////////////////////////////////
// TO PROTOBUF

func toProtobufGoogleCastDevice(device rpc.GoogleCastDevice) *pb.GoogleCastDevice {
	if device == nil {
		return nil
	}
	return &pb.GoogleCastDevice{
		Id:      device.Id(),
		Name:    device.Name(),
		Model:   device.Model(),
		Service: device.Service(),
		State:   uint32(device.State()),
	}
}

func toProtobufGoogleCastDeviceReply(devices []rpc.GoogleCastDevice) *pb.GoogleCastDeviceReply {
	if devices == nil {
		return nil
	}
	reply := make([]*pb.GoogleCastDevice, len(devices))
	for i, device := range devices {
		reply[i] = toProtobufGoogleCastDevice(device)
	}
	return &pb.GoogleCastDeviceReply{
		Device: reply,
	}
}

func toProtoGoogleCastEvent(evt rpc.GoogleCastEvent) *pb.GoogleCastEvent {
	if evt == nil {
		return nil
	}
	if ts, err := ptypes.TimestampProto(evt.Timestamp()); err != nil {
		return nil
	} else {
		reply := &pb.GoogleCastEvent{
			Type:   pb.GoogleCastEvent_EventType(evt.Type()),
			Device: toProtobufGoogleCastDevice(evt.Device()),
			Ts:     ts,
		}
		return reply
	}
}

////////////////////////////////////////////////////////////////////////////////
// FROM PROTOBUF

func fromProtobufGoogleCastDeviceReply(proto []*pb.GoogleCastDevice) []rpc.GoogleCastDevice {
	if proto == nil {
		return nil
	}
	devices := make([]rpc.GoogleCastDevice, len(proto))
	for i, device := range proto {
		devices[i] = fromProtobufGoogleCastDevice(device)
	}
	return devices
}

func fromProtobufGoogleCastDevice(proto *pb.GoogleCastDevice) rpc.GoogleCastDevice {
	if proto == nil {
		return nil
	}
	return &castdevice{proto}
}

////////////////////////////////////////////////////////////////////////////////
// GoogleCastDevice IMPLEMENTATION

type castdevice struct {
	pb *pb.GoogleCastDevice
}

func (this *castdevice) Id() string {
	return this.pb.Id
}

func (this *castdevice) Model() string {
	return this.pb.Model
}

func (this *castdevice) Name() string {
	return this.pb.Name
}

func (this *castdevice) Service() string {
	return this.pb.Service
}

func (this *castdevice) State() uint {
	return uint(this.pb.State)
}
