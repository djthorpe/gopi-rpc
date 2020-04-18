// Code generated by protoc-gen-go. DO NOT EDIT.
// source: gaffer/service.proto

package gaffer

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// A service defintion is used to start one or more processes
type Service struct {
	Sid                  uint32   `protobuf:"varint,1,opt,name=sid,proto3" json:"sid,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Path                 string   `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	Cwd                  string   `protobuf:"bytes,4,opt,name=cwd,proto3" json:"cwd,omitempty"`
	Args                 []string `protobuf:"bytes,5,rep,name=args,proto3" json:"args,omitempty"`
	User                 string   `protobuf:"bytes,6,opt,name=user,proto3" json:"user,omitempty"`
	Group                string   `protobuf:"bytes,7,opt,name=group,proto3" json:"group,omitempty"`
	Enabled              bool     `protobuf:"varint,8,opt,name=enabled,proto3" json:"enabled,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Service) Reset()         { *m = Service{} }
func (m *Service) String() string { return proto.CompactTextString(m) }
func (*Service) ProtoMessage()    {}
func (*Service) Descriptor() ([]byte, []int) {
	return fileDescriptor_d89f813df1b8899b, []int{0}
}

func (m *Service) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Service.Unmarshal(m, b)
}
func (m *Service) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Service.Marshal(b, m, deterministic)
}
func (m *Service) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Service.Merge(m, src)
}
func (m *Service) XXX_Size() int {
	return xxx_messageInfo_Service.Size(m)
}
func (m *Service) XXX_DiscardUnknown() {
	xxx_messageInfo_Service.DiscardUnknown(m)
}

var xxx_messageInfo_Service proto.InternalMessageInfo

func (m *Service) GetSid() uint32 {
	if m != nil {
		return m.Sid
	}
	return 0
}

func (m *Service) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Service) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *Service) GetCwd() string {
	if m != nil {
		return m.Cwd
	}
	return ""
}

func (m *Service) GetArgs() []string {
	if m != nil {
		return m.Args
	}
	return nil
}

func (m *Service) GetUser() string {
	if m != nil {
		return m.User
	}
	return ""
}

func (m *Service) GetGroup() string {
	if m != nil {
		return m.Group
	}
	return ""
}

func (m *Service) GetEnabled() bool {
	if m != nil {
		return m.Enabled
	}
	return false
}

// Return a list of processes
type ServiceList struct {
	Service              []*Service `protobuf:"bytes,1,rep,name=service,proto3" json:"service,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ServiceList) Reset()         { *m = ServiceList{} }
func (m *ServiceList) String() string { return proto.CompactTextString(m) }
func (*ServiceList) ProtoMessage()    {}
func (*ServiceList) Descriptor() ([]byte, []int) {
	return fileDescriptor_d89f813df1b8899b, []int{1}
}

func (m *ServiceList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceList.Unmarshal(m, b)
}
func (m *ServiceList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceList.Marshal(b, m, deterministic)
}
func (m *ServiceList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceList.Merge(m, src)
}
func (m *ServiceList) XXX_Size() int {
	return xxx_messageInfo_ServiceList.Size(m)
}
func (m *ServiceList) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceList.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceList proto.InternalMessageInfo

func (m *ServiceList) GetService() []*Service {
	if m != nil {
		return m.Service
	}
	return nil
}

func init() {
	proto.RegisterType((*Service)(nil), "gaffer.Service")
	proto.RegisterType((*ServiceList)(nil), "gaffer.ServiceList")
}

func init() {
	proto.RegisterFile("gaffer/service.proto", fileDescriptor_d89f813df1b8899b)
}

var fileDescriptor_d89f813df1b8899b = []byte{
	// 268 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x86, 0x59, 0xd3, 0x26, 0x71, 0x8a, 0x28, 0x63, 0x91, 0xa1, 0x5e, 0x42, 0x4f, 0xf1, 0xb2,
	0x81, 0x0a, 0xea, 0x0b, 0x88, 0x17, 0x0f, 0x12, 0x9f, 0x20, 0x69, 0x36, 0x6b, 0xa0, 0x4d, 0xc2,
	0x6e, 0xa2, 0xf4, 0xa1, 0x7c, 0x47, 0x99, 0xdd, 0xe4, 0x22, 0xf4, 0xf6, 0xcf, 0x37, 0xff, 0xbf,
	0x3b, 0x33, 0xb0, 0xd6, 0x45, 0x5d, 0x2b, 0x93, 0x59, 0x65, 0xbe, 0x9b, 0xbd, 0x92, 0xbd, 0xe9,
	0x86, 0x0e, 0x43, 0x4f, 0x37, 0xf7, 0xba, 0xeb, 0xf4, 0x41, 0x65, 0x8e, 0x96, 0x63, 0x9d, 0xa9,
	0x63, 0x3f, 0x9c, 0xbc, 0x69, 0xfb, 0x2b, 0x20, 0xfa, 0xf4, 0x31, 0xbc, 0x81, 0xc0, 0x36, 0x15,
	0x89, 0x44, 0xa4, 0x57, 0x39, 0x4b, 0x44, 0x58, 0xb4, 0xc5, 0x51, 0xd1, 0x45, 0x22, 0xd2, 0xcb,
	0xdc, 0x69, 0x66, 0x7d, 0x31, 0x7c, 0x51, 0xe0, 0x19, 0x6b, 0x4e, 0xee, 0x7f, 0x2a, 0x5a, 0x38,
	0xc4, 0x92, 0x5d, 0x85, 0xd1, 0x96, 0x96, 0x49, 0xc0, 0x2e, 0xd6, 0xcc, 0x46, 0xab, 0x0c, 0x85,
	0x3e, 0xc9, 0x1a, 0xd7, 0xb0, 0xd4, 0xa6, 0x1b, 0x7b, 0x8a, 0x1c, 0xf4, 0x05, 0x12, 0x44, 0xaa,
	0x2d, 0xca, 0x83, 0xaa, 0x28, 0x4e, 0x44, 0x1a, 0xe7, 0x73, 0xb9, 0x7d, 0x81, 0xd5, 0x34, 0xee,
	0x7b, 0x63, 0x07, 0x7c, 0x80, 0x68, 0x5a, 0x9a, 0x44, 0x12, 0xa4, 0xab, 0xdd, 0xb5, 0xf4, 0x5b,
	0xcb, 0xc9, 0x95, 0xcf, 0xfd, 0xdd, 0x09, 0xc2, 0x37, 0xd7, 0xc2, 0x27, 0x58, 0x7c, 0x34, 0xad,
	0xc6, 0x3b, 0xe9, 0x2f, 0x23, 0xe7, 0xcb, 0xc8, 0x57, 0xbe, 0xcc, 0xe6, 0x0c, 0xc7, 0x67, 0x88,
	0xa7, 0x57, 0xed, 0xd9, 0xec, 0xed, 0xbf, 0xff, 0x79, 0xca, 0x32, 0x74, 0xa6, 0xc7, 0xbf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x31, 0xb3, 0x0e, 0xf2, 0xa8, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// GafferClient is the client API for Gaffer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GafferClient interface {
	// Simple ping method to show server is "up"
	Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
	// Return services
	Services(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ServiceList, error)
}

type gafferClient struct {
	cc grpc.ClientConnInterface
}

func NewGafferClient(cc grpc.ClientConnInterface) GafferClient {
	return &gafferClient{cc}
}

func (c *gafferClient) Ping(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/gaffer.Gaffer/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *gafferClient) Services(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ServiceList, error) {
	out := new(ServiceList)
	err := c.cc.Invoke(ctx, "/gaffer.Gaffer/Services", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GafferServer is the server API for Gaffer service.
type GafferServer interface {
	// Simple ping method to show server is "up"
	Ping(context.Context, *empty.Empty) (*empty.Empty, error)
	// Return services
	Services(context.Context, *empty.Empty) (*ServiceList, error)
}

// UnimplementedGafferServer can be embedded to have forward compatible implementations.
type UnimplementedGafferServer struct {
}

func (*UnimplementedGafferServer) Ping(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (*UnimplementedGafferServer) Services(ctx context.Context, req *empty.Empty) (*ServiceList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Services not implemented")
}

func RegisterGafferServer(s *grpc.Server, srv GafferServer) {
	s.RegisterService(&_Gaffer_serviceDesc, srv)
}

func _Gaffer_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GafferServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gaffer.Gaffer/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GafferServer).Ping(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Gaffer_Services_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GafferServer).Services(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gaffer.Gaffer/Services",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GafferServer).Services(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _Gaffer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gaffer.Gaffer",
	HandlerType: (*GafferServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _Gaffer_Ping_Handler,
		},
		{
			MethodName: "Services",
			Handler:    _Gaffer_Services_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gaffer/service.proto",
}