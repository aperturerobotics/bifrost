// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0-devel
// 	protoc        v4.25.3
// source: github.com/aperturerobotics/bifrost/transport/common/kcp/coord_stream.proto

package kcp

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// CoordPacketType is the packet type of a coordination stream packet.
type CoordPacketType int32

const (
	CoordPacketType_CoordPacketType_UNKNOWN           CoordPacketType = 0
	CoordPacketType_CoordPacketType_RSTREAM_ESTABLISH CoordPacketType = 1
	CoordPacketType_CoordPacketType_RSTREAM_ACK       CoordPacketType = 2
	CoordPacketType_CoordPacketType_RSTREAM_CLOSE     CoordPacketType = 3
	CoordPacketType_CoordPacketType_RSTREAM_NOOP      CoordPacketType = 4
)

// Enum value maps for CoordPacketType.
var (
	CoordPacketType_name = map[int32]string{
		0: "CoordPacketType_UNKNOWN",
		1: "CoordPacketType_RSTREAM_ESTABLISH",
		2: "CoordPacketType_RSTREAM_ACK",
		3: "CoordPacketType_RSTREAM_CLOSE",
		4: "CoordPacketType_RSTREAM_NOOP",
	}
	CoordPacketType_value = map[string]int32{
		"CoordPacketType_UNKNOWN":           0,
		"CoordPacketType_RSTREAM_ESTABLISH": 1,
		"CoordPacketType_RSTREAM_ACK":       2,
		"CoordPacketType_RSTREAM_CLOSE":     3,
		"CoordPacketType_RSTREAM_NOOP":      4,
	}
)

func (x CoordPacketType) Enum() *CoordPacketType {
	p := new(CoordPacketType)
	*p = x
	return p
}

func (x CoordPacketType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CoordPacketType) Descriptor() protoreflect.EnumDescriptor {
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_enumTypes[0].Descriptor()
}

func (CoordPacketType) Type() protoreflect.EnumType {
	return &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_enumTypes[0]
}

func (x CoordPacketType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CoordPacketType.Descriptor instead.
func (CoordPacketType) EnumDescriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescGZIP(), []int{0}
}

// CoordinationStreamPacket is the packet wrapper for a coordination stream
// packet.
type CoordinationStreamPacket struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// PacketType is the coordination stream packet type.
	PacketType CoordPacketType `protobuf:"varint,1,opt,name=packet_type,json=packetType,proto3,enum=kcp.CoordPacketType" json:"packet_type,omitempty"`
	// RawStreamEstablish is the raw stream establish packet.
	RawStreamEstablish *RawStreamEstablish `protobuf:"bytes,2,opt,name=raw_stream_establish,json=rawStreamEstablish,proto3" json:"raw_stream_establish,omitempty"`
	// RawStreamAck is the raw stream ack packet.
	RawStreamAck *RawStreamAck `protobuf:"bytes,3,opt,name=raw_stream_ack,json=rawStreamAck,proto3" json:"raw_stream_ack,omitempty"`
	// RawStreamClose is the raw stream close packet.
	RawStreamClose *RawStreamClose `protobuf:"bytes,4,opt,name=raw_stream_close,json=rawStreamClose,proto3" json:"raw_stream_close,omitempty"`
}

func (x *CoordinationStreamPacket) Reset() {
	*x = CoordinationStreamPacket{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CoordinationStreamPacket) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CoordinationStreamPacket) ProtoMessage() {}

func (x *CoordinationStreamPacket) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CoordinationStreamPacket.ProtoReflect.Descriptor instead.
func (*CoordinationStreamPacket) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescGZIP(), []int{0}
}

func (x *CoordinationStreamPacket) GetPacketType() CoordPacketType {
	if x != nil {
		return x.PacketType
	}
	return CoordPacketType_CoordPacketType_UNKNOWN
}

func (x *CoordinationStreamPacket) GetRawStreamEstablish() *RawStreamEstablish {
	if x != nil {
		return x.RawStreamEstablish
	}
	return nil
}

func (x *CoordinationStreamPacket) GetRawStreamAck() *RawStreamAck {
	if x != nil {
		return x.RawStreamAck
	}
	return nil
}

func (x *CoordinationStreamPacket) GetRawStreamClose() *RawStreamClose {
	if x != nil {
		return x.RawStreamClose
	}
	return nil
}

// RawStreamEstablish is a coordination stream raw-stream establish message.
type RawStreamEstablish struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// InitiatorStreamId is the stream ID the initiator wants to use.
	InitiatorStreamId uint32 `protobuf:"varint,1,opt,name=initiator_stream_id,json=initiatorStreamId,proto3" json:"initiator_stream_id,omitempty"`
}

func (x *RawStreamEstablish) Reset() {
	*x = RawStreamEstablish{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RawStreamEstablish) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RawStreamEstablish) ProtoMessage() {}

func (x *RawStreamEstablish) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RawStreamEstablish.ProtoReflect.Descriptor instead.
func (*RawStreamEstablish) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescGZIP(), []int{1}
}

func (x *RawStreamEstablish) GetInitiatorStreamId() uint32 {
	if x != nil {
		return x.InitiatorStreamId
	}
	return 0
}

// RawStreamAck is a coordination stream raw-stream acknowledge message.
type RawStreamAck struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// InitiatorStreamId is the stream ID the initiator wanted to use.
	InitiatorStreamId uint32 `protobuf:"varint,1,opt,name=initiator_stream_id,json=initiatorStreamId,proto3" json:"initiator_stream_id,omitempty"`
	// AckStreamId is the stream ID the responder wants to use.
	// Zero if the stream was rejected.
	AckStreamId uint32 `protobuf:"varint,2,opt,name=ack_stream_id,json=ackStreamId,proto3" json:"ack_stream_id,omitempty"`
	// AckError indicates an error establishing the stream, rejecting the stream.
	AckError string `protobuf:"bytes,3,opt,name=ack_error,json=ackError,proto3" json:"ack_error,omitempty"`
}

func (x *RawStreamAck) Reset() {
	*x = RawStreamAck{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RawStreamAck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RawStreamAck) ProtoMessage() {}

func (x *RawStreamAck) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RawStreamAck.ProtoReflect.Descriptor instead.
func (*RawStreamAck) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescGZIP(), []int{2}
}

func (x *RawStreamAck) GetInitiatorStreamId() uint32 {
	if x != nil {
		return x.InitiatorStreamId
	}
	return 0
}

func (x *RawStreamAck) GetAckStreamId() uint32 {
	if x != nil {
		return x.AckStreamId
	}
	return 0
}

func (x *RawStreamAck) GetAckError() string {
	if x != nil {
		return x.AckError
	}
	return ""
}

// RawStreamClose indicates an intent to close a raw stream.
type RawStreamClose struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// StreamId is the stream ID the reciever indicated to use.
	StreamId uint32 `protobuf:"varint,1,opt,name=stream_id,json=streamId,proto3" json:"stream_id,omitempty"`
	// CloseError indicates an error included with the stream close.
	CloseError string `protobuf:"bytes,2,opt,name=close_error,json=closeError,proto3" json:"close_error,omitempty"`
}

func (x *RawStreamClose) Reset() {
	*x = RawStreamClose{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RawStreamClose) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RawStreamClose) ProtoMessage() {}

func (x *RawStreamClose) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RawStreamClose.ProtoReflect.Descriptor instead.
func (*RawStreamClose) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescGZIP(), []int{3}
}

func (x *RawStreamClose) GetStreamId() uint32 {
	if x != nil {
		return x.StreamId
	}
	return 0
}

func (x *RawStreamClose) GetCloseError() string {
	if x != nil {
		return x.CloseError
	}
	return ""
}

var File_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDesc = []byte{
	0x0a, 0x4b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x6b, 0x63, 0x70, 0x2f, 0x63, 0x6f, 0x6f, 0x72, 0x64,
	0x5f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x6b,
	0x63, 0x70, 0x22, 0x94, 0x02, 0x0a, 0x18, 0x43, 0x6f, 0x6f, 0x72, 0x64, 0x69, 0x6e, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x12,
	0x35, 0x0a, 0x0b, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x6b, 0x63, 0x70, 0x2e, 0x43, 0x6f, 0x6f, 0x72, 0x64,
	0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x70, 0x61, 0x63, 0x6b,
	0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x49, 0x0a, 0x14, 0x72, 0x61, 0x77, 0x5f, 0x73, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x5f, 0x65, 0x73, 0x74, 0x61, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x6b, 0x63, 0x70, 0x2e, 0x52, 0x61, 0x77, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x45, 0x73, 0x74, 0x61, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x52, 0x12, 0x72,
	0x61, 0x77, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x45, 0x73, 0x74, 0x61, 0x62, 0x6c, 0x69, 0x73,
	0x68, 0x12, 0x37, 0x0a, 0x0e, 0x72, 0x61, 0x77, 0x5f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f,
	0x61, 0x63, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x6b, 0x63, 0x70, 0x2e,
	0x52, 0x61, 0x77, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x41, 0x63, 0x6b, 0x52, 0x0c, 0x72, 0x61,
	0x77, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x41, 0x63, 0x6b, 0x12, 0x3d, 0x0a, 0x10, 0x72, 0x61,
	0x77, 0x5f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x6b, 0x63, 0x70, 0x2e, 0x52, 0x61, 0x77, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x0e, 0x72, 0x61, 0x77, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x22, 0x44, 0x0a, 0x12, 0x52, 0x61, 0x77,
	0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x45, 0x73, 0x74, 0x61, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x12,
	0x2e, 0x0a, 0x13, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x11, 0x69, 0x6e,
	0x69, 0x74, 0x69, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x22,
	0x7f, 0x0a, 0x0c, 0x52, 0x61, 0x77, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x41, 0x63, 0x6b, 0x12,
	0x2e, 0x0a, 0x13, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x74, 0x6f, 0x72, 0x5f, 0x73, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x11, 0x69, 0x6e,
	0x69, 0x74, 0x69, 0x61, 0x74, 0x6f, 0x72, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12,
	0x22, 0x0a, 0x0d, 0x61, 0x63, 0x6b, 0x5f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0b, 0x61, 0x63, 0x6b, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x63, 0x6b, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x61, 0x63, 0x6b, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x22, 0x4e, 0x0a, 0x0e, 0x52, 0x61, 0x77, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x43, 0x6c, 0x6f,
	0x73, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12,
	0x1f, 0x0a, 0x0b, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x2a, 0xbb, 0x01, 0x0a, 0x0f, 0x43, 0x6f, 0x6f, 0x72, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x1b, 0x0a, 0x17, 0x43, 0x6f, 0x6f, 0x72, 0x64, 0x50, 0x61, 0x63,
	0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10,
	0x00, 0x12, 0x25, 0x0a, 0x21, 0x43, 0x6f, 0x6f, 0x72, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x5f, 0x52, 0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x5f, 0x45, 0x53, 0x54,
	0x41, 0x42, 0x4c, 0x49, 0x53, 0x48, 0x10, 0x01, 0x12, 0x1f, 0x0a, 0x1b, 0x43, 0x6f, 0x6f, 0x72,
	0x64, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x52, 0x53, 0x54, 0x52,
	0x45, 0x41, 0x4d, 0x5f, 0x41, 0x43, 0x4b, 0x10, 0x02, 0x12, 0x21, 0x0a, 0x1d, 0x43, 0x6f, 0x6f,
	0x72, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x52, 0x53, 0x54,
	0x52, 0x45, 0x41, 0x4d, 0x5f, 0x43, 0x4c, 0x4f, 0x53, 0x45, 0x10, 0x03, 0x12, 0x20, 0x0a, 0x1c,
	0x43, 0x6f, 0x6f, 0x72, 0x64, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x5f,
	0x52, 0x53, 0x54, 0x52, 0x45, 0x41, 0x4d, 0x5f, 0x4e, 0x4f, 0x4f, 0x50, 0x10, 0x04, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescData = file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_goTypes = []interface{}{
	(CoordPacketType)(0),             // 0: kcp.CoordPacketType
	(*CoordinationStreamPacket)(nil), // 1: kcp.CoordinationStreamPacket
	(*RawStreamEstablish)(nil),       // 2: kcp.RawStreamEstablish
	(*RawStreamAck)(nil),             // 3: kcp.RawStreamAck
	(*RawStreamClose)(nil),           // 4: kcp.RawStreamClose
}
var file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_depIdxs = []int32{
	0, // 0: kcp.CoordinationStreamPacket.packet_type:type_name -> kcp.CoordPacketType
	2, // 1: kcp.CoordinationStreamPacket.raw_stream_establish:type_name -> kcp.RawStreamEstablish
	3, // 2: kcp.CoordinationStreamPacket.raw_stream_ack:type_name -> kcp.RawStreamAck
	4, // 3: kcp.CoordinationStreamPacket.raw_stream_close:type_name -> kcp.RawStreamClose
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_init() }
func file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_init() {
	if File_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CoordinationStreamPacket); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RawStreamEstablish); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RawStreamAck); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RawStreamClose); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_depIdxs,
		EnumInfos:         file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_enumTypes,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto = out.File
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_coord_stream_proto_depIdxs = nil
}
