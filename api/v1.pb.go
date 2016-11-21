// Code generated by protoc-gen-go.
// source: v1.proto
// DO NOT EDIT!

/*
Package api is a generated protocol buffer package.

It is generated from these files:
        v1.proto

It has these top-level messages:
        RunRequest
        RunReply
        StopTaskRequest
        StopTaskReply
        ResourceRequest
        ResourceReply
        ResourceIDRequest
*/
package api

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
//const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type RunRequest struct {
        ImageName string `protobuf:"bytes,1,opt,name=imageName" json:"imageName,omitempty"`
        HostName  string `protobuf:"bytes,2,opt,name=hostName" json:"hostName,omitempty"`
        TaskType  string `protobuf:"bytes,3,opt,name=taskType" json:"taskType,omitempty"`
}

func (m *RunRequest) Reset()                    { *m = RunRequest{} }
func (m *RunRequest) String() string            { return proto.CompactTextString(m) }
func (*RunRequest) ProtoMessage()               {}
func (*RunRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type RunReply struct {
        InstanceId string `protobuf:"bytes,1,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
}

func (m *RunReply) Reset()                    { *m = RunReply{} }
func (m *RunReply) String() string            { return proto.CompactTextString(m) }
func (*RunReply) ProtoMessage()               {}
func (*RunReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type StopTaskRequest struct {
        HostName string `protobuf:"bytes,1,opt,name=hostName" json:"hostName,omitempty"`
        TaskType string `protobuf:"bytes,2,opt,name=taskType" json:"taskType,omitempty"`
}

func (m *StopTaskRequest) Reset()                    { *m = StopTaskRequest{} }
func (m *StopTaskRequest) String() string            { return proto.CompactTextString(m) }
func (*StopTaskRequest) ProtoMessage()               {}
func (*StopTaskRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type StopTaskReply struct {
        InstanceId string `protobuf:"bytes,1,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
}

func (m *StopTaskReply) Reset()                    { *m = StopTaskReply{} }
func (m *StopTaskReply) String() string            { return proto.CompactTextString(m) }
func (*StopTaskReply) ProtoMessage()               {}
func (*StopTaskReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type ResourceRequest struct {
}

func (m *ResourceRequest) Reset()                    { *m = ResourceRequest{} }
func (m *ResourceRequest) String() string            { return proto.CompactTextString(m) }
func (*ResourceRequest) ProtoMessage()               {}
func (*ResourceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type ResourceReply struct {
        ID string `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
}

func (m *ResourceReply) Reset()                    { *m = ResourceReply{} }
func (m *ResourceReply) String() string            { return proto.CompactTextString(m) }
func (*ResourceReply) ProtoMessage()               {}
func (*ResourceReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type ResourceIDRequest struct {
        // Types that are valid to be assigned to Key:
        //      *ResourceIDRequest_ID
        //      *ResourceIDRequest_Name
        Key isResourceIDRequest_Key `protobuf_oneof:"Key"`
}

func (m *ResourceIDRequest) Reset()                    { *m = ResourceIDRequest{} }
func (m *ResourceIDRequest) String() string            { return proto.CompactTextString(m) }
func (*ResourceIDRequest) ProtoMessage()               {}
func (*ResourceIDRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type isResourceIDRequest_Key interface {
        isResourceIDRequest_Key()
}

type ResourceIDRequest_ID struct {
        ID string `protobuf:"bytes,1,opt,name=ID,oneof"`
}
type ResourceIDRequest_Name struct {
        Name string `protobuf:"bytes,2,opt,name=Name,oneof"`
}

func (*ResourceIDRequest_ID) isResourceIDRequest_Key()   {}
func (*ResourceIDRequest_Name) isResourceIDRequest_Key() {}

func (m *ResourceIDRequest) GetKey() isResourceIDRequest_Key {
        if m != nil {
                return m.Key
        }
        return nil
}

func (m *ResourceIDRequest) GetID() string {
        if x, ok := m.GetKey().(*ResourceIDRequest_ID); ok {
                return x.ID
        }
        return ""
}

func (m *ResourceIDRequest) GetName() string {
        if x, ok := m.GetKey().(*ResourceIDRequest_Name); ok {
                return x.Name
        }
        return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*ResourceIDRequest) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
        return _ResourceIDRequest_OneofMarshaler, _ResourceIDRequest_OneofUnmarshaler, _ResourceIDRequest_OneofSizer, []interface{}{
                (*ResourceIDRequest_ID)(nil),
                (*ResourceIDRequest_Name)(nil),
        }
}

func _ResourceIDRequest_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
        m := msg.(*ResourceIDRequest)
        // Key
        switch x := m.Key.(type) {
        case *ResourceIDRequest_ID:
                b.EncodeVarint(1<<3 | proto.WireBytes)
                b.EncodeStringBytes(x.ID)
        case *ResourceIDRequest_Name:
                b.EncodeVarint(2<<3 | proto.WireBytes)
                b.EncodeStringBytes(x.Name)
        case nil:
        default:
                return fmt.Errorf("ResourceIDRequest.Key has unexpected type %T", x)
        }
        return nil
}

func _ResourceIDRequest_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
        m := msg.(*ResourceIDRequest)
        switch tag {
        case 1: // Key.ID
                if wire != proto.WireBytes {
                        return true, proto.ErrInternalBadWireType
                }
                x, err := b.DecodeStringBytes()
                m.Key = &ResourceIDRequest_ID{x}
                return true, err
        case 2: // Key.Name
                if wire != proto.WireBytes {
                        return true, proto.ErrInternalBadWireType
                }
                x, err := b.DecodeStringBytes()
                m.Key = &ResourceIDRequest_Name{x}
                return true, err
        default:
                return false, nil
        }
}

func _ResourceIDRequest_OneofSizer(msg proto.Message) (n int) {
        m := msg.(*ResourceIDRequest)
        // Key
        switch x := m.Key.(type) {
        case *ResourceIDRequest_ID:
                n += proto.SizeVarint(1<<3 | proto.WireBytes)
                n += proto.SizeVarint(uint64(len(x.ID)))
                n += len(x.ID)
        case *ResourceIDRequest_Name:
                n += proto.SizeVarint(2<<3 | proto.WireBytes)
                n += proto.SizeVarint(uint64(len(x.Name)))
                n += len(x.Name)
        case nil:
        default:
                panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
        }
        return n
}

func init() {
        proto.RegisterType((*RunRequest)(nil), "api.RunRequest")
        proto.RegisterType((*RunReply)(nil), "api.RunReply")
        proto.RegisterType((*StopTaskRequest)(nil), "api.StopTaskRequest")
        proto.RegisterType((*StopTaskReply)(nil), "api.StopTaskReply")
        proto.RegisterType((*ResourceRequest)(nil), "api.ResourceRequest")
        proto.RegisterType((*ResourceReply)(nil), "api.ResourceReply")
        proto.RegisterType((*ResourceIDRequest)(nil), "api.ResourceIDRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
//const _ = grpc.SupportPackageIsVersion3

// Client API for Instance service

type InstanceClient interface {
        Run(ctx context.Context, in *RunRequest, opts ...grpc.CallOption) (*RunReply, error)
        StopTask(ctx context.Context, in *StopTaskRequest, opts ...grpc.CallOption) (*StopTaskReply, error)
}

type instanceClient struct {
        cc *grpc.ClientConn
}

func NewInstanceClient(cc *grpc.ClientConn) InstanceClient {
        return &instanceClient{cc}
}

func (c *instanceClient) Run(ctx context.Context, in *RunRequest, opts ...grpc.CallOption) (*RunReply, error) {
        out := new(RunReply)
        err := grpc.Invoke(ctx, "/api.Instance/Run", in, out, c.cc, opts...)
        if err != nil {
                return nil, err
        }
        return out, nil
}

func (c *instanceClient) StopTask(ctx context.Context, in *StopTaskRequest, opts ...grpc.CallOption) (*StopTaskReply, error) {
        out := new(StopTaskReply)
        err := grpc.Invoke(ctx, "/api.Instance/StopTask", in, out, c.cc, opts...)
        if err != nil {
                return nil, err
        }
        return out, nil
}

// Server API for Instance service

type InstanceServer interface {
        Run(context.Context, *RunRequest) (*RunReply, error)
        StopTask(context.Context, *StopTaskRequest) (*StopTaskReply, error)
}

func RegisterInstanceServer(s *grpc.Server, srv InstanceServer) {
        s.RegisterService(&_Instance_serviceDesc, srv)
}

func _Instance_Run_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
        in := new(RunRequest)
        if err := dec(in); err != nil {
                return nil, err
        }
        if interceptor == nil {
                return srv.(InstanceServer).Run(ctx, in)
        }
        info := &grpc.UnaryServerInfo{
                Server:     srv,
                FullMethod: "/api.Instance/Run",
        }
        handler := func(ctx context.Context, req interface{}) (interface{}, error) {
                return srv.(InstanceServer).Run(ctx, req.(*RunRequest))
        }
        return interceptor(ctx, in, info, handler)
}

func _Instance_StopTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
        in := new(StopTaskRequest)
        if err := dec(in); err != nil {
                return nil, err
        }
        if interceptor == nil {
                return srv.(InstanceServer).StopTask(ctx, in)
        }
        info := &grpc.UnaryServerInfo{
                Server:     srv,
                FullMethod: "/api.Instance/StopTask",
        }
        handler := func(ctx context.Context, req interface{}) (interface{}, error) {
                return srv.(InstanceServer).StopTask(ctx, req.(*StopTaskRequest))
        }
        return interceptor(ctx, in, info, handler)
}

var _Instance_serviceDesc = grpc.ServiceDesc{
        ServiceName: "api.Instance",
        HandlerType: (*InstanceServer)(nil),
        Methods: []grpc.MethodDesc{
                {
                        MethodName: "Run",
                        Handler:    _Instance_Run_Handler,
                },
                {
                        MethodName: "StopTask",
                        Handler:    _Instance_StopTask_Handler,
                },
        },
        Streams:  []grpc.StreamDesc{},
        Metadata: fileDescriptor0,
}

// Client API for Resource service

type ResourceClient interface {
        Register(ctx context.Context, in *ResourceRequest, opts ...grpc.CallOption) (*ResourceReply, error)
        Unregister(ctx context.Context, in *ResourceIDRequest, opts ...grpc.CallOption) (*ResourceReply, error)
}

type resourceClient struct {
        cc *grpc.ClientConn
}

func NewResourceClient(cc *grpc.ClientConn) ResourceClient {
        return &resourceClient{cc}
}

func (c *resourceClient) Register(ctx context.Context, in *ResourceRequest, opts ...grpc.CallOption) (*ResourceReply, error) {
        out := new(ResourceReply)
        err := grpc.Invoke(ctx, "/api.Resource/Register", in, out, c.cc, opts...)
        if err != nil {
                return nil, err
        }
        return out, nil
}

func (c *resourceClient) Unregister(ctx context.Context, in *ResourceIDRequest, opts ...grpc.CallOption) (*ResourceReply, error) {
        out := new(ResourceReply)
        err := grpc.Invoke(ctx, "/api.Resource/Unregister", in, out, c.cc, opts...)
        if err != nil {
                return nil, err
        }
        return out, nil
}

// Server API for Resource service

type ResourceServer interface {
        Register(context.Context, *ResourceRequest) (*ResourceReply, error)
        Unregister(context.Context, *ResourceIDRequest) (*ResourceReply, error)
}

func RegisterResourceServer(s *grpc.Server, srv ResourceServer) {
        s.RegisterService(&_Resource_serviceDesc, srv)
}

func _Resource_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
        in := new(ResourceRequest)
        if err := dec(in); err != nil {
                return nil, err
        }
        if interceptor == nil {
                return srv.(ResourceServer).Register(ctx, in)
        }
        info := &grpc.UnaryServerInfo{
                Server:     srv,
                FullMethod: "/api.Resource/Register",
        }
        handler := func(ctx context.Context, req interface{}) (interface{}, error) {
                return srv.(ResourceServer).Register(ctx, req.(*ResourceRequest))
        }
        return interceptor(ctx, in, info, handler)
}

func _Resource_Unregister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
        in := new(ResourceIDRequest)
        if err := dec(in); err != nil {
                return nil, err
        }
        if interceptor == nil {
                return srv.(ResourceServer).Unregister(ctx, in)
        }
        info := &grpc.UnaryServerInfo{
                Server:     srv,
                FullMethod: "/api.Resource/Unregister",
        }
        handler := func(ctx context.Context, req interface{}) (interface{}, error) {
                return srv.(ResourceServer).Unregister(ctx, req.(*ResourceIDRequest))
        }
        return interceptor(ctx, in, info, handler)
}

var _Resource_serviceDesc = grpc.ServiceDesc{
        ServiceName: "api.Resource",
        HandlerType: (*ResourceServer)(nil),
        Methods: []grpc.MethodDesc{
                {
                        MethodName: "Register",
                        Handler:    _Resource_Register_Handler,
                },
                {
                        MethodName: "Unregister",
                        Handler:    _Resource_Unregister_Handler,
                },
        },
        Streams:  []grpc.StreamDesc{},
        Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("v1.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
        // 319 bytes of a gzipped FileDescriptorProto
        0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x52, 0xc1, 0x4a, 0xc3, 0x40,
        0x10, 0x6d, 0x13, 0x95, 0x74, 0x24, 0xc6, 0x2e, 0x45, 0x4a, 0x10, 0x2a, 0x7b, 0x51, 0x10, 0x82,
        0x56, 0xf0, 0xe0, 0xb1, 0xf4, 0x60, 0x10, 0x3c, 0xc4, 0x7a, 0x96, 0xb4, 0x1d, 0x6a, 0xa8, 0x4d,
        0xd6, 0xec, 0x46, 0xc8, 0xc5, 0x6f, 0x77, 0xb3, 0xee, 0x76, 0x6d, 0x10, 0xf1, 0x96, 0x79, 0x33,
        0xf3, 0xde, 0xdb, 0x37, 0x01, 0xef, 0xe3, 0x3a, 0x62, 0x65, 0x21, 0x0a, 0xe2, 0xa6, 0x2c, 0xa3,
        0x73, 0x80, 0xa4, 0xca, 0x13, 0x7c, 0xaf, 0x90, 0x0b, 0x72, 0x0a, 0xbd, 0x6c, 0x93, 0xae, 0xf0,
        0x31, 0xdd, 0xe0, 0xb0, 0x7b, 0xd6, 0xbd, 0xe8, 0x25, 0x16, 0x20, 0x21, 0x78, 0xaf, 0x05, 0x17,
        0xaa, 0xe9, 0xa8, 0xe6, 0xb6, 0x6e, 0x7a, 0x22, 0xe5, 0xeb, 0x59, 0xcd, 0x70, 0xe8, 0x7e, 0xf7,
        0x4c, 0x4d, 0x2f, 0xc1, 0x53, 0x1a, 0xec, 0xad, 0x26, 0x23, 0x38, 0xcc, 0x72, 0x2e, 0xd2, 0x7c,
        0x81, 0x2f, 0xd9, 0x52, 0x6b, 0x80, 0x81, 0xe2, 0x25, 0x8d, 0x21, 0x78, 0x12, 0x05, 0x9b, 0xc9,
        0x65, 0xe3, 0xea, 0xa7, 0x6e, 0xf7, 0x0f, 0x5d, 0xa7, 0xa5, 0x7b, 0x05, 0xbe, 0xa5, 0xfa, 0x97,
        0x78, 0x1f, 0x82, 0x04, 0x79, 0x51, 0x95, 0x0b, 0xd4, 0xe2, 0x74, 0x04, 0xbe, 0x85, 0x1a, 0x92,
        0x23, 0x70, 0xe2, 0xa9, 0xde, 0x95, 0x5f, 0x74, 0x02, 0x7d, 0x33, 0x10, 0x4f, 0x8d, 0xe5, 0x63,
        0x3b, 0x74, 0xdf, 0x69, 0xc6, 0xc8, 0x00, 0xf6, 0x6c, 0x70, 0x12, 0x53, 0xd5, 0x64, 0x1f, 0xdc,
        0x07, 0xac, 0xc7, 0x6b, 0xf0, 0x62, 0xed, 0x82, 0x9c, 0x83, 0x2b, 0xd3, 0x22, 0x41, 0x24, 0xcf,
        0x13, 0xd9, 0xdb, 0x84, 0xbe, 0x05, 0xa4, 0x0d, 0xda, 0x21, 0xb7, 0xe0, 0x99, 0xe7, 0x91, 0x81,
        0x6a, 0xb6, 0x82, 0x0b, 0x49, 0x0b, 0x55, 0x7b, 0xe3, 0x4f, 0x79, 0x0e, 0x6d, 0xb8, 0xe1, 0x48,
        0x70, 0x95, 0x71, 0x81, 0xa5, 0xe6, 0x68, 0xbd, 0x5f, 0x73, 0xec, 0x44, 0x20, 0xb5, 0xef, 0x00,
        0x9e, 0xf3, 0xd2, 0x6c, 0x9e, 0xec, 0xcc, 0x6c, 0x53, 0xf8, 0x7d, 0x77, 0x7e, 0xa0, 0x7e, 0xbf,
        0x9b, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd1, 0xbc, 0x56, 0x61, 0x8a, 0x02, 0x00, 0x00,
}
