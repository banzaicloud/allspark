// Code generated by protoc-gen-go.
// source: allspark.proto
// DO NOT EDIT!

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	allspark.proto

It has these top-level messages:
	Params
	Msg
*/
package pb

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

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Params struct {
}

func (m *Params) Reset()                    { *m = Params{} }
func (m *Params) String() string            { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()               {}
func (*Params) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Msg struct {
	Response string `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

func (m *Msg) Reset()                    { *m = Msg{} }
func (m *Msg) String() string            { return proto.CompactTextString(m) }
func (*Msg) ProtoMessage()               {}
func (*Msg) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Msg) GetResponse() string {
	if m != nil {
		return m.Response
	}
	return ""
}

func init() {
	proto.RegisterType((*Params)(nil), "params")
	proto.RegisterType((*Msg)(nil), "msg")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Allspark service

type AllsparkClient interface {
	Incoming(ctx context.Context, in *Params, opts ...grpc.CallOption) (*Msg, error)
}

type allsparkClient struct {
	cc *grpc.ClientConn
}

func NewAllsparkClient(cc *grpc.ClientConn) AllsparkClient {
	return &allsparkClient{cc}
}

func (c *allsparkClient) Incoming(ctx context.Context, in *Params, opts ...grpc.CallOption) (*Msg, error) {
	out := new(Msg)
	err := grpc.Invoke(ctx, "/allspark/Incoming", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Allspark service

type AllsparkServer interface {
	Incoming(context.Context, *Params) (*Msg, error)
}

func RegisterAllsparkServer(s *grpc.Server, srv AllsparkServer) {
	s.RegisterService(&_Allspark_serviceDesc, srv)
}

func _Allspark_Incoming_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Params)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AllsparkServer).Incoming(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/allspark/Incoming",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AllsparkServer).Incoming(ctx, req.(*Params))
	}
	return interceptor(ctx, in, info, handler)
}

var _Allspark_serviceDesc = grpc.ServiceDesc{
	ServiceName: "allspark",
	HandlerType: (*AllsparkServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Incoming",
			Handler:    _Allspark_Incoming_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "allspark.proto",
}

func init() { proto.RegisterFile("allspark.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 118 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4b, 0xcc, 0xc9, 0x29,
	0x2e, 0x48, 0x2c, 0xca, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x57, 0xe2, 0xe0, 0x62, 0x2b, 0x48,
	0x2c, 0x4a, 0xcc, 0x2d, 0x56, 0x52, 0xe4, 0x62, 0xce, 0x2d, 0x4e, 0x17, 0x92, 0xe2, 0xe2, 0x28,
	0x4a, 0x2d, 0x2e, 0xc8, 0xcf, 0x2b, 0x4e, 0x95, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x0c, 0x82, 0xf3,
	0x8d, 0x54, 0xb9, 0x38, 0x60, 0xda, 0x85, 0x24, 0xb9, 0x38, 0x3c, 0xf3, 0x92, 0xf3, 0x73, 0x33,
	0xf3, 0xd2, 0x85, 0xd8, 0xf5, 0x20, 0x66, 0x48, 0xb1, 0xe8, 0xe5, 0x16, 0xa7, 0x3b, 0xb1, 0x44,
	0x31, 0x15, 0x24, 0x25, 0xb1, 0x81, 0x2d, 0x30, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x4f, 0x54,
	0x54, 0xb7, 0x72, 0x00, 0x00, 0x00,
}
