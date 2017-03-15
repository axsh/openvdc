// Code generated by protoc-gen-go.
// source: cluster.proto
// DO NOT EDIT!

package model

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Console_Transport int32

const (
	Console_SSH Console_Transport = 0
)

var Console_Transport_name = map[int32]string{
	0: "SSH",
}
var Console_Transport_value = map[string]int32{
	"SSH": 0,
}

func (x Console_Transport) String() string {
	return proto.EnumName(Console_Transport_name, int32(x))
}
func (Console_Transport) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{0, 0} }

type NodeState_State int32

const (
	NodeState_REGISTERED NodeState_State = 0
)

var NodeState_State_name = map[int32]string{
	0: "REGISTERED",
}
var NodeState_State_value = map[string]int32{
	"REGISTERED": 0,
}

func (x NodeState_State) String() string {
	return proto.EnumName(NodeState_State_name, int32(x))
}
func (NodeState_State) EnumDescriptor() ([]byte, []int) { return fileDescriptor1, []int{3, 0} }

type Console struct {
	Type     Console_Transport `protobuf:"varint,1,opt,name=type,enum=model.Console_Transport" json:"type,omitempty"`
	BindAddr string            `protobuf:"bytes,2,opt,name=bind_addr,json=bindAddr" json:"bind_addr,omitempty"`
}

func (m *Console) Reset()                    { *m = Console{} }
func (m *Console) String() string            { return proto.CompactTextString(m) }
func (*Console) ProtoMessage()               {}
func (*Console) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *Console) GetType() Console_Transport {
	if m != nil {
		return m.Type
	}
	return Console_SSH
}

func (m *Console) GetBindAddr() string {
	if m != nil {
		return m.BindAddr
	}
	return ""
}

type ExecutorNode struct {
	Id        string                     `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	CreatedAt *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	Console   *Console                   `protobuf:"bytes,3,opt,name=console" json:"console,omitempty"`
	GrpcAddr  string                     `protobuf:"bytes,4,opt,name=grpc_addr,json=grpcAddr" json:"grpc_addr,omitempty"`
	LastState *NodeState                 `protobuf:"bytes,5,opt,name=last_state,json=lastState" json:"last_state,omitempty"`
}

func (m *ExecutorNode) Reset()                    { *m = ExecutorNode{} }
func (m *ExecutorNode) String() string            { return proto.CompactTextString(m) }
func (*ExecutorNode) ProtoMessage()               {}
func (*ExecutorNode) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{1} }

func (m *ExecutorNode) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *ExecutorNode) GetCreatedAt() *google_protobuf.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func (m *ExecutorNode) GetConsole() *Console {
	if m != nil {
		return m.Console
	}
	return nil
}

func (m *ExecutorNode) GetGrpcAddr() string {
	if m != nil {
		return m.GrpcAddr
	}
	return ""
}

func (m *ExecutorNode) GetLastState() *NodeState {
	if m != nil {
		return m.LastState
	}
	return nil
}

type SchedulerNode struct {
	Id        string                     `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	CreatedAt *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
}

func (m *SchedulerNode) Reset()                    { *m = SchedulerNode{} }
func (m *SchedulerNode) String() string            { return proto.CompactTextString(m) }
func (*SchedulerNode) ProtoMessage()               {}
func (*SchedulerNode) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{2} }

func (m *SchedulerNode) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *SchedulerNode) GetCreatedAt() *google_protobuf.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

type NodeState struct {
	State     NodeState_State            `protobuf:"varint,1,opt,name=state,enum=model.NodeState_State" json:"state,omitempty"`
	CreatedAt *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
}

func (m *NodeState) Reset()                    { *m = NodeState{} }
func (m *NodeState) String() string            { return proto.CompactTextString(m) }
func (*NodeState) ProtoMessage()               {}
func (*NodeState) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{3} }

func (m *NodeState) GetState() NodeState_State {
	if m != nil {
		return m.State
	}
	return NodeState_REGISTERED
}

func (m *NodeState) GetCreatedAt() *google_protobuf.Timestamp {
	if m != nil {
		return m.CreatedAt
	}
	return nil
}

func init() {
	proto.RegisterType((*Console)(nil), "model.Console")
	proto.RegisterType((*ExecutorNode)(nil), "model.ExecutorNode")
	proto.RegisterType((*SchedulerNode)(nil), "model.SchedulerNode")
	proto.RegisterType((*NodeState)(nil), "model.NodeState")
	proto.RegisterEnum("model.Console_Transport", Console_Transport_name, Console_Transport_value)
	proto.RegisterEnum("model.NodeState_State", NodeState_State_name, NodeState_State_value)
}

func init() { proto.RegisterFile("cluster.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 374 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x92, 0xb1, 0xce, 0xd3, 0x30,
	0x14, 0x85, 0x9b, 0xb6, 0xa1, 0xe4, 0x42, 0xa3, 0xca, 0x42, 0x10, 0x15, 0x21, 0xaa, 0x4c, 0x1d,
	0x2a, 0x47, 0x2a, 0x13, 0x62, 0x2a, 0x10, 0x01, 0x0b, 0x83, 0xd3, 0xa9, 0x4b, 0xe5, 0xd8, 0x26,
	0x0d, 0x4a, 0xe2, 0xc8, 0x76, 0x50, 0x79, 0x08, 0x1e, 0x8e, 0x37, 0x42, 0xb6, 0xdb, 0x02, 0xff,
	0xfa, 0xff, 0xdb, 0xd5, 0xbd, 0x27, 0xe7, 0x7c, 0x47, 0x31, 0xcc, 0x59, 0x33, 0x68, 0x23, 0x14,
	0xee, 0x95, 0x34, 0x12, 0x85, 0xad, 0xe4, 0xa2, 0x59, 0xbe, 0xab, 0x6a, 0x73, 0x1a, 0x4a, 0xcc,
	0x64, 0x9b, 0x55, 0xb2, 0xa1, 0x5d, 0x95, 0xb9, 0x7b, 0x39, 0x7c, 0xcb, 0x7a, 0xf3, 0xb3, 0x17,
	0x3a, 0x33, 0x75, 0x2b, 0xb4, 0xa1, 0x6d, 0xff, 0x77, 0xf2, 0x1e, 0xe9, 0x77, 0x98, 0x7d, 0x90,
	0x9d, 0x96, 0x8d, 0x40, 0x1b, 0x98, 0x5a, 0x75, 0x12, 0xac, 0x82, 0x75, 0xbc, 0x4d, 0xb0, 0x73,
	0xc7, 0x97, 0x2b, 0xde, 0x2b, 0xda, 0xe9, 0x5e, 0x2a, 0x43, 0x9c, 0x0a, 0xbd, 0x84, 0xa8, 0xac,
	0x3b, 0x7e, 0xa4, 0x9c, 0xab, 0x64, 0xbc, 0x0a, 0xd6, 0x11, 0x79, 0x6c, 0x17, 0x3b, 0xce, 0x55,
	0xfa, 0x0c, 0xa2, 0x9b, 0x1e, 0xcd, 0x60, 0x52, 0x14, 0x9f, 0x17, 0xa3, 0xf4, 0x77, 0x00, 0x4f,
	0xf3, 0xb3, 0x60, 0x83, 0x91, 0xea, 0xab, 0xe4, 0x02, 0xc5, 0x30, 0xae, 0xb9, 0xcb, 0x8b, 0xc8,
	0xb8, 0xe6, 0xe8, 0x2d, 0x00, 0x53, 0x82, 0x1a, 0xc1, 0x8f, 0xd4, 0x38, 0xd3, 0x27, 0xdb, 0x25,
	0xae, 0xa4, 0xac, 0x1a, 0x81, 0xaf, 0x9d, 0xf0, 0xfe, 0x5a, 0x81, 0x44, 0x17, 0xf5, 0xce, 0xa0,
	0x35, 0xcc, 0x98, 0x27, 0x4d, 0x26, 0xee, 0xbb, 0xf8, 0x7f, 0x7e, 0x72, 0x3d, 0x5b, 0xf0, 0x4a,
	0xf5, 0xcc, 0x83, 0x4f, 0x3d, 0xb8, 0x5d, 0x58, 0x70, 0x94, 0x01, 0x34, 0x54, 0x9b, 0xa3, 0x36,
	0xd4, 0x88, 0x24, 0x74, 0x4e, 0x8b, 0x8b, 0x93, 0x45, 0x2e, 0xec, 0x9e, 0x44, 0x56, 0xe3, 0xc6,
	0xf4, 0x00, 0xf3, 0x82, 0x9d, 0x04, 0x1f, 0x1a, 0xf1, 0xd0, 0x9d, 0xd2, 0x5f, 0x01, 0x44, 0xb7,
	0x50, 0xb4, 0x81, 0xd0, 0x53, 0xf9, 0xff, 0xf3, 0xfc, 0x2e, 0x15, 0xf6, 0x6c, 0x5e, 0x74, 0x9f,
	0xd8, 0x17, 0x10, 0xfa, 0xc4, 0x18, 0x80, 0xe4, 0x9f, 0xbe, 0x14, 0xfb, 0x9c, 0xe4, 0x1f, 0x17,
	0xa3, 0xf7, 0xaf, 0x0f, 0xaf, 0xfe, 0x79, 0x6a, 0xf4, 0xac, 0x4f, 0x99, 0xec, 0x45, 0xf7, 0x83,
	0xb3, 0xcc, 0xb1, 0x94, 0x8f, 0x9c, 0xf1, 0x9b, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x28, 0xc7,
	0xe4, 0x8e, 0xa8, 0x02, 0x00, 0x00,
}
