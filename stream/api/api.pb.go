// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/bifrost/stream/api/api.proto

package stream_api

import (
	fmt "fmt"
	math "math"

	accept "github.com/aperturerobotics/bifrost/stream/api/accept"
	dial "github.com/aperturerobotics/bifrost/stream/api/dial"
	rpc "github.com/aperturerobotics/bifrost/stream/api/rpc"
	forwarding "github.com/aperturerobotics/bifrost/stream/forwarding"
	listening "github.com/aperturerobotics/bifrost/stream/listening"
	exec "github.com/aperturerobotics/controllerbus/controller/exec"
	proto "github.com/golang/protobuf/proto"
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

// ForwardStreamsRequest is the request type for ForwardStreams.
type ForwardStreamsRequest struct {
	ForwardingConfig     *forwarding.Config `protobuf:"bytes,1,opt,name=forwarding_config,json=forwardingConfig,proto3" json:"forwarding_config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *ForwardStreamsRequest) Reset()         { *m = ForwardStreamsRequest{} }
func (m *ForwardStreamsRequest) String() string { return proto.CompactTextString(m) }
func (*ForwardStreamsRequest) ProtoMessage()    {}
func (*ForwardStreamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc527d05dedd29fb, []int{0}
}

func (m *ForwardStreamsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ForwardStreamsRequest.Unmarshal(m, b)
}
func (m *ForwardStreamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ForwardStreamsRequest.Marshal(b, m, deterministic)
}
func (m *ForwardStreamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ForwardStreamsRequest.Merge(m, src)
}
func (m *ForwardStreamsRequest) XXX_Size() int {
	return xxx_messageInfo_ForwardStreamsRequest.Size(m)
}
func (m *ForwardStreamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ForwardStreamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ForwardStreamsRequest proto.InternalMessageInfo

func (m *ForwardStreamsRequest) GetForwardingConfig() *forwarding.Config {
	if m != nil {
		return m.ForwardingConfig
	}
	return nil
}

// ForwardStreamsResponse is the response type for ForwardStreams.
type ForwardStreamsResponse struct {
	// ControllerStatus is the status of the forwarding controller.
	ControllerStatus     exec.ControllerStatus `protobuf:"varint,1,opt,name=controller_status,json=controllerStatus,proto3,enum=controller.exec.ControllerStatus" json:"controller_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *ForwardStreamsResponse) Reset()         { *m = ForwardStreamsResponse{} }
func (m *ForwardStreamsResponse) String() string { return proto.CompactTextString(m) }
func (*ForwardStreamsResponse) ProtoMessage()    {}
func (*ForwardStreamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc527d05dedd29fb, []int{1}
}

func (m *ForwardStreamsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ForwardStreamsResponse.Unmarshal(m, b)
}
func (m *ForwardStreamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ForwardStreamsResponse.Marshal(b, m, deterministic)
}
func (m *ForwardStreamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ForwardStreamsResponse.Merge(m, src)
}
func (m *ForwardStreamsResponse) XXX_Size() int {
	return xxx_messageInfo_ForwardStreamsResponse.Size(m)
}
func (m *ForwardStreamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ForwardStreamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ForwardStreamsResponse proto.InternalMessageInfo

func (m *ForwardStreamsResponse) GetControllerStatus() exec.ControllerStatus {
	if m != nil {
		return m.ControllerStatus
	}
	return exec.ControllerStatus_ControllerStatus_UNKNOWN
}

// ListenStreamsRequest is the request type for ListenStreams.
type ListenStreamsRequest struct {
	ListeningConfig      *listening.Config `protobuf:"bytes,1,opt,name=listening_config,json=listeningConfig,proto3" json:"listening_config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *ListenStreamsRequest) Reset()         { *m = ListenStreamsRequest{} }
func (m *ListenStreamsRequest) String() string { return proto.CompactTextString(m) }
func (*ListenStreamsRequest) ProtoMessage()    {}
func (*ListenStreamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc527d05dedd29fb, []int{2}
}

func (m *ListenStreamsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListenStreamsRequest.Unmarshal(m, b)
}
func (m *ListenStreamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListenStreamsRequest.Marshal(b, m, deterministic)
}
func (m *ListenStreamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListenStreamsRequest.Merge(m, src)
}
func (m *ListenStreamsRequest) XXX_Size() int {
	return xxx_messageInfo_ListenStreamsRequest.Size(m)
}
func (m *ListenStreamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListenStreamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListenStreamsRequest proto.InternalMessageInfo

func (m *ListenStreamsRequest) GetListeningConfig() *listening.Config {
	if m != nil {
		return m.ListeningConfig
	}
	return nil
}

// ListenStreamsResponse is the response type for ListenStreams.
type ListenStreamsResponse struct {
	// ControllerStatus is the status of the forwarding controller.
	ControllerStatus     exec.ControllerStatus `protobuf:"varint,1,opt,name=controller_status,json=controllerStatus,proto3,enum=controller.exec.ControllerStatus" json:"controller_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *ListenStreamsResponse) Reset()         { *m = ListenStreamsResponse{} }
func (m *ListenStreamsResponse) String() string { return proto.CompactTextString(m) }
func (*ListenStreamsResponse) ProtoMessage()    {}
func (*ListenStreamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc527d05dedd29fb, []int{3}
}

func (m *ListenStreamsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListenStreamsResponse.Unmarshal(m, b)
}
func (m *ListenStreamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListenStreamsResponse.Marshal(b, m, deterministic)
}
func (m *ListenStreamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListenStreamsResponse.Merge(m, src)
}
func (m *ListenStreamsResponse) XXX_Size() int {
	return xxx_messageInfo_ListenStreamsResponse.Size(m)
}
func (m *ListenStreamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListenStreamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListenStreamsResponse proto.InternalMessageInfo

func (m *ListenStreamsResponse) GetControllerStatus() exec.ControllerStatus {
	if m != nil {
		return m.ControllerStatus
	}
	return exec.ControllerStatus_ControllerStatus_UNKNOWN
}

// AcceptStreamRequest is the request type for AcceptStream.
type AcceptStreamRequest struct {
	// Config is the configuration for the accept.
	// The first packet will contain this value.
	Config *accept.Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	// Data is a data packet.
	Data                 *rpc.Data `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *AcceptStreamRequest) Reset()         { *m = AcceptStreamRequest{} }
func (m *AcceptStreamRequest) String() string { return proto.CompactTextString(m) }
func (*AcceptStreamRequest) ProtoMessage()    {}
func (*AcceptStreamRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc527d05dedd29fb, []int{4}
}

func (m *AcceptStreamRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AcceptStreamRequest.Unmarshal(m, b)
}
func (m *AcceptStreamRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AcceptStreamRequest.Marshal(b, m, deterministic)
}
func (m *AcceptStreamRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AcceptStreamRequest.Merge(m, src)
}
func (m *AcceptStreamRequest) XXX_Size() int {
	return xxx_messageInfo_AcceptStreamRequest.Size(m)
}
func (m *AcceptStreamRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_AcceptStreamRequest.DiscardUnknown(m)
}

var xxx_messageInfo_AcceptStreamRequest proto.InternalMessageInfo

func (m *AcceptStreamRequest) GetConfig() *accept.Config {
	if m != nil {
		return m.Config
	}
	return nil
}

func (m *AcceptStreamRequest) GetData() *rpc.Data {
	if m != nil {
		return m.Data
	}
	return nil
}

// AcceptStreamResponse is the response type for AcceptStream.
type AcceptStreamResponse struct {
	// Data is a data packet.
	Data                 *rpc.Data `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *AcceptStreamResponse) Reset()         { *m = AcceptStreamResponse{} }
func (m *AcceptStreamResponse) String() string { return proto.CompactTextString(m) }
func (*AcceptStreamResponse) ProtoMessage()    {}
func (*AcceptStreamResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc527d05dedd29fb, []int{5}
}

func (m *AcceptStreamResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AcceptStreamResponse.Unmarshal(m, b)
}
func (m *AcceptStreamResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AcceptStreamResponse.Marshal(b, m, deterministic)
}
func (m *AcceptStreamResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AcceptStreamResponse.Merge(m, src)
}
func (m *AcceptStreamResponse) XXX_Size() int {
	return xxx_messageInfo_AcceptStreamResponse.Size(m)
}
func (m *AcceptStreamResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_AcceptStreamResponse.DiscardUnknown(m)
}

var xxx_messageInfo_AcceptStreamResponse proto.InternalMessageInfo

func (m *AcceptStreamResponse) GetData() *rpc.Data {
	if m != nil {
		return m.Data
	}
	return nil
}

// DialStreamRequest is the request type for DialStream.
type DialStreamRequest struct {
	// Config is the configuration for the dial.
	// The first packet will contain this value.
	Config *dial.Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	// Data is a data packet.
	Data                 *rpc.Data `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *DialStreamRequest) Reset()         { *m = DialStreamRequest{} }
func (m *DialStreamRequest) String() string { return proto.CompactTextString(m) }
func (*DialStreamRequest) ProtoMessage()    {}
func (*DialStreamRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc527d05dedd29fb, []int{6}
}

func (m *DialStreamRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DialStreamRequest.Unmarshal(m, b)
}
func (m *DialStreamRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DialStreamRequest.Marshal(b, m, deterministic)
}
func (m *DialStreamRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DialStreamRequest.Merge(m, src)
}
func (m *DialStreamRequest) XXX_Size() int {
	return xxx_messageInfo_DialStreamRequest.Size(m)
}
func (m *DialStreamRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DialStreamRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DialStreamRequest proto.InternalMessageInfo

func (m *DialStreamRequest) GetConfig() *dial.Config {
	if m != nil {
		return m.Config
	}
	return nil
}

func (m *DialStreamRequest) GetData() *rpc.Data {
	if m != nil {
		return m.Data
	}
	return nil
}

// DialStreamResponse is the response type for DialStream.
type DialStreamResponse struct {
	// Data is a data packet.
	Data                 *rpc.Data `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *DialStreamResponse) Reset()         { *m = DialStreamResponse{} }
func (m *DialStreamResponse) String() string { return proto.CompactTextString(m) }
func (*DialStreamResponse) ProtoMessage()    {}
func (*DialStreamResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_bc527d05dedd29fb, []int{7}
}

func (m *DialStreamResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DialStreamResponse.Unmarshal(m, b)
}
func (m *DialStreamResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DialStreamResponse.Marshal(b, m, deterministic)
}
func (m *DialStreamResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DialStreamResponse.Merge(m, src)
}
func (m *DialStreamResponse) XXX_Size() int {
	return xxx_messageInfo_DialStreamResponse.Size(m)
}
func (m *DialStreamResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DialStreamResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DialStreamResponse proto.InternalMessageInfo

func (m *DialStreamResponse) GetData() *rpc.Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*ForwardStreamsRequest)(nil), "stream.api.ForwardStreamsRequest")
	proto.RegisterType((*ForwardStreamsResponse)(nil), "stream.api.ForwardStreamsResponse")
	proto.RegisterType((*ListenStreamsRequest)(nil), "stream.api.ListenStreamsRequest")
	proto.RegisterType((*ListenStreamsResponse)(nil), "stream.api.ListenStreamsResponse")
	proto.RegisterType((*AcceptStreamRequest)(nil), "stream.api.AcceptStreamRequest")
	proto.RegisterType((*AcceptStreamResponse)(nil), "stream.api.AcceptStreamResponse")
	proto.RegisterType((*DialStreamRequest)(nil), "stream.api.DialStreamRequest")
	proto.RegisterType((*DialStreamResponse)(nil), "stream.api.DialStreamResponse")
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/bifrost/stream/api/api.proto", fileDescriptor_bc527d05dedd29fb)
}

var fileDescriptor_bc527d05dedd29fb = []byte{
	// 491 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x54, 0xc1, 0x8e, 0xd3, 0x30,
	0x10, 0x25, 0x2b, 0xb4, 0x87, 0x81, 0x5d, 0x5a, 0xd3, 0x85, 0x12, 0x09, 0x68, 0x73, 0xea, 0x29,
	0x59, 0xca, 0x85, 0x03, 0x5a, 0x01, 0xad, 0x7a, 0x42, 0x48, 0xb4, 0x07, 0x90, 0xf6, 0x50, 0x39,
	0xae, 0xdb, 0xb5, 0x14, 0xe2, 0x60, 0x3b, 0xc0, 0x77, 0xf3, 0x05, 0x28, 0xb6, 0x93, 0x38, 0x21,
	0x5b, 0x14, 0xa4, 0x3d, 0x24, 0x4d, 0xc6, 0xf3, 0xde, 0xcc, 0x9b, 0x37, 0x29, 0xbc, 0x39, 0x30,
	0x75, 0x93, 0xc7, 0x21, 0xe1, 0xdf, 0x22, 0x9c, 0x51, 0xa1, 0x72, 0x41, 0x05, 0x8f, 0xb9, 0x62,
	0x44, 0x46, 0x31, 0xdb, 0x0b, 0x2e, 0x55, 0x24, 0x95, 0xa0, 0xb8, 0x38, 0x67, 0xc5, 0x15, 0x66,
	0x82, 0x2b, 0x8e, 0xc0, 0x44, 0x43, 0x9c, 0x31, 0x7f, 0x79, 0x8c, 0x85, 0xf0, 0x54, 0x09, 0x9e,
	0x24, 0x54, 0xc4, 0xb9, 0xfb, 0x16, 0xd1, 0x5f, 0x94, 0xe8, 0x9b, 0x61, 0xf4, 0x57, 0x3d, 0x7a,
	0xd9, 0x73, 0xf1, 0x13, 0x8b, 0x1d, 0x4b, 0x0f, 0xce, 0xa3, 0xe5, 0x59, 0xf6, 0xe0, 0x49, 0x98,
	0x54, 0x34, 0x2d, 0x68, 0xaa, 0x27, 0xcb, 0xf2, 0xb6, 0xe7, 0x64, 0x44, 0x46, 0x8a, 0xcb, 0xa2,
	0x3f, 0xf4, 0x9d, 0x2b, 0x21, 0x34, 0x53, 0xf6, 0xc7, 0x72, 0x5c, 0xf5, 0xe4, 0xd8, 0x31, 0x9c,
	0xe8, 0x9b, 0xc1, 0x07, 0x5b, 0xb8, 0x58, 0x99, 0xd9, 0x6c, 0x74, 0x92, 0x5c, 0xd3, 0xef, 0x39,
	0x95, 0x0a, 0xad, 0x60, 0x58, 0x0f, 0x6d, 0x4b, 0x78, 0xba, 0x67, 0x87, 0xb1, 0x37, 0xf1, 0x66,
	0x0f, 0xe6, 0xcf, 0x42, 0x6b, 0xab, 0x33, 0xd5, 0x85, 0x4e, 0x58, 0x0f, 0xea, 0x90, 0x89, 0x04,
	0x37, 0xf0, 0xa4, 0x5d, 0x40, 0x66, 0x3c, 0x95, 0x14, 0x7d, 0x82, 0x61, 0x6d, 0xf4, 0x56, 0x2a,
	0xac, 0x72, 0xa9, 0x2b, 0x9c, 0xcf, 0xa7, 0x61, 0x7d, 0x12, 0x6a, 0xf7, 0x17, 0xd5, 0xfb, 0x46,
	0x27, 0xae, 0x07, 0xa4, 0x15, 0x09, 0xae, 0x61, 0xf4, 0x51, 0xfb, 0xd3, 0x52, 0xb2, 0x80, 0x41,
	0xe5, 0x5b, 0x53, 0xc8, 0xb8, 0x14, 0x52, 0xfb, 0x6a, 0x75, 0x3c, 0xaa, 0x22, 0x56, 0xc6, 0x01,
	0x2e, 0x5a, 0xe4, 0x77, 0xa4, 0x42, 0xc0, 0xe3, 0xf7, 0xda, 0x60, 0x53, 0xa8, 0x14, 0xf1, 0x0a,
	0x4e, 0xbb, 0x3d, 0x28, 0x3e, 0x36, 0xbb, 0x11, 0xb6, 0x77, 0x9b, 0x88, 0x66, 0x70, 0x7f, 0x87,
	0x15, 0x1e, 0x9f, 0x68, 0xc0, 0xc8, 0x05, 0x14, 0x3b, 0xb8, 0xc4, 0x0a, 0xaf, 0x75, 0x46, 0xf0,
	0x0e, 0x46, 0xcd, 0x9a, 0x56, 0x5b, 0xc9, 0xe0, 0xfd, 0x93, 0x21, 0x85, 0xe1, 0x92, 0xe1, 0xa4,
	0xd9, 0x73, 0xd4, 0xea, 0xf9, 0xa9, 0x4b, 0xa0, 0x77, 0xf0, 0xbf, 0x3b, 0xbe, 0x02, 0xe4, 0xd6,
	0xeb, 0xdb, 0xef, 0xfc, 0xf7, 0x09, 0x9c, 0x19, 0xf0, 0x86, 0x8a, 0x1f, 0x8c, 0x50, 0x74, 0x0d,
	0xe7, 0xcd, 0x3d, 0x45, 0x53, 0x17, 0xdf, 0xf9, 0x91, 0xf8, 0xc1, 0xb1, 0x14, 0xd3, 0x54, 0x70,
	0xef, 0xd2, 0x43, 0x5f, 0xe1, 0xac, 0xb1, 0x3d, 0x68, 0xe2, 0x02, 0xbb, 0xb6, 0xd6, 0x9f, 0x1e,
	0xc9, 0x70, 0x98, 0xbf, 0xc0, 0x43, 0xd7, 0x3a, 0xf4, 0xd2, 0x85, 0x75, 0x2c, 0x92, 0x3f, 0xb9,
	0x3d, 0xa1, 0xa4, 0x9d, 0x79, 0x97, 0x1e, 0xfa, 0x0c, 0x50, 0x4f, 0x18, 0x3d, 0x77, 0x51, 0x7f,
	0x39, 0xed, 0xbf, 0xb8, 0xed, 0xd8, 0xa5, 0x8c, 0x4f, 0xf5, 0x5f, 0xce, 0xeb, 0x3f, 0x01, 0x00,
	0x00, 0xff, 0xff, 0xac, 0x19, 0xe0, 0xa9, 0x50, 0x06, 0x00, 0x00,
}
