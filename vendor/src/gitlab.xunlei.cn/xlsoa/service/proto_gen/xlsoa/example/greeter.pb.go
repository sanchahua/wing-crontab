// Code generated by protoc-gen-go.
// source: proto/xlsoa/example/greeter.proto
// DO NOT EDIT!

package xlsoa_example

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// The request message containing the user's name.
type HelloRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *HelloRequest) Reset()                    { *m = HelloRequest{} }
func (m *HelloRequest) String() string            { return proto.CompactTextString(m) }
func (*HelloRequest) ProtoMessage()               {}
func (*HelloRequest) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *HelloRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

// The response message containing the greetings
type HelloReply struct {
	Message string `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
}

func (m *HelloReply) Reset()                    { *m = HelloReply{} }
func (m *HelloReply) String() string            { return proto.CompactTextString(m) }
func (*HelloReply) ProtoMessage()               {}
func (*HelloReply) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *HelloReply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*HelloRequest)(nil), "xlsoa.example.HelloRequest")
	proto.RegisterType((*HelloReply)(nil), "xlsoa.example.HelloReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Greeter service

type GreeterClient interface {
	// Sends a greeting
	SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error)
}

type greeterClient struct {
	cc *grpc.ClientConn
}

func NewGreeterClient(cc *grpc.ClientConn) GreeterClient {
	return &greeterClient{cc}
}

func (c *greeterClient) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error) {
	out := new(HelloReply)
	err := grpc.Invoke(ctx, "/xlsoa.example.greeter/SayHello", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Greeter service

type GreeterServer interface {
	// Sends a greeting
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
}

func RegisterGreeterServer(s *grpc.Server, srv GreeterServer) {
	s.RegisterService(&_Greeter_serviceDesc, srv)
}

func _Greeter_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/xlsoa.example.greeter/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).SayHello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Greeter_serviceDesc = grpc.ServiceDesc{
	ServiceName: "xlsoa.example.greeter",
	HandlerType: (*GreeterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Greeter_SayHello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/xlsoa/example/greeter.proto",
}

func init() { proto.RegisterFile("proto/xlsoa/example/greeter.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 156 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x52, 0x2c, 0x28, 0xca, 0x2f,
	0xc9, 0xd7, 0xaf, 0xc8, 0x29, 0xce, 0x4f, 0xd4, 0x4f, 0xad, 0x48, 0xcc, 0x2d, 0xc8, 0x49, 0xd5,
	0x4f, 0x2f, 0x4a, 0x4d, 0x2d, 0x49, 0x2d, 0xd2, 0x03, 0xcb, 0x09, 0xf1, 0x82, 0x25, 0xf5, 0xa0,
	0x92, 0x4a, 0x4a, 0x5c, 0x3c, 0x1e, 0xa9, 0x39, 0x39, 0xf9, 0x41, 0xa9, 0x85, 0xa5, 0xa9, 0xc5,
	0x25, 0x42, 0x42, 0x5c, 0x2c, 0x79, 0x89, 0xb9, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41,
	0x60, 0xb6, 0x92, 0x1a, 0x17, 0x17, 0x54, 0x4d, 0x41, 0x4e, 0xa5, 0x90, 0x04, 0x17, 0x7b, 0x6e,
	0x6a, 0x71, 0x71, 0x62, 0x3a, 0x4c, 0x11, 0x8c, 0x6b, 0xe4, 0xcf, 0xc5, 0x0e, 0xb5, 0x4b, 0xc8,
	0x85, 0x8b, 0x23, 0x38, 0xb1, 0x12, 0xac, 0x4b, 0x48, 0x5a, 0x0f, 0xc5, 0x4a, 0x3d, 0x64, 0xfb,
	0xa4, 0x24, 0xb1, 0x4b, 0x16, 0xe4, 0x54, 0x2a, 0x31, 0x24, 0xb1, 0x81, 0x9d, 0x6c, 0x0c, 0x08,
	0x00, 0x00, 0xff, 0xff, 0xf7, 0x2b, 0xf8, 0x3a, 0xd7, 0x00, 0x00, 0x00,
}
