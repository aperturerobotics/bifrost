// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0-devel
// 	protoc        v4.25.1
// source: github.com/aperturerobotics/bifrost/handshake/identity/s2s/s2s.proto

package s2s

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

type PacketType int32

const (
	// INIT initializes the handshake.
	PacketType_PacketType_INIT PacketType = 0
	// INIT_ACK is the reply to the init.
	PacketType_PacketType_INIT_ACK PacketType = 1
	// COMPLETE is the completion of the handshake.
	PacketType_PacketType_COMPLETE PacketType = 2
)

// Enum value maps for PacketType.
var (
	PacketType_name = map[int32]string{
		0: "PacketType_INIT",
		1: "PacketType_INIT_ACK",
		2: "PacketType_COMPLETE",
	}
	PacketType_value = map[string]int32{
		"PacketType_INIT":     0,
		"PacketType_INIT_ACK": 1,
		"PacketType_COMPLETE": 2,
	}
)

func (x PacketType) Enum() *PacketType {
	p := new(PacketType)
	*p = x
	return p
}

func (x PacketType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PacketType) Descriptor() protoreflect.EnumDescriptor {
	return file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_enumTypes[0].Descriptor()
}

func (PacketType) Type() protoreflect.EnumType {
	return &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_enumTypes[0]
}

func (x PacketType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PacketType.Descriptor instead.
func (PacketType) EnumDescriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescGZIP(), []int{0}
}

// Packet is a handshake packet.
type Packet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// PacketType is the packet type.
	PacketType PacketType `protobuf:"varint,1,opt,name=packet_type,json=packetType,proto3,enum=s2s.PacketType" json:"packet_type,omitempty"`
	// InitPkt is the init packet.
	InitPkt *Packet_Init `protobuf:"bytes,2,opt,name=init_pkt,json=initPkt,proto3" json:"init_pkt,omitempty"`
	// InitAck is the init-ack packet.
	InitAckPkt *Packet_InitAck `protobuf:"bytes,3,opt,name=init_ack_pkt,json=initAckPkt,proto3" json:"init_ack_pkt,omitempty"`
	// Complete is the complete packet.
	CompletePkt *Packet_Complete `protobuf:"bytes,4,opt,name=complete_pkt,json=completePkt,proto3" json:"complete_pkt,omitempty"`
}

func (x *Packet) Reset() {
	*x = Packet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Packet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet) ProtoMessage() {}

func (x *Packet) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet.ProtoReflect.Descriptor instead.
func (*Packet) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescGZIP(), []int{0}
}

func (x *Packet) GetPacketType() PacketType {
	if x != nil {
		return x.PacketType
	}
	return PacketType_PacketType_INIT
}

func (x *Packet) GetInitPkt() *Packet_Init {
	if x != nil {
		return x.InitPkt
	}
	return nil
}

func (x *Packet) GetInitAckPkt() *Packet_InitAck {
	if x != nil {
		return x.InitAckPkt
	}
	return nil
}

func (x *Packet) GetCompletePkt() *Packet_Complete {
	if x != nil {
		return x.CompletePkt
	}
	return nil
}

type Packet_Init struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// SenderPeerID is the peer ID of the sender.
	SenderPeerId []byte `protobuf:"bytes,1,opt,name=sender_peer_id,json=senderPeerId,proto3" json:"sender_peer_id,omitempty"`
	// ReceiverPeerID is the receiver peer ID, if known.
	// If this does not match, the public key is included in the next message.
	ReceiverPeerId []byte `protobuf:"bytes,2,opt,name=receiver_peer_id,json=receiverPeerId,proto3" json:"receiver_peer_id,omitempty"`
	// SenderEphPub is the ephemeral public key of the sender.
	SenderEphPub []byte `protobuf:"bytes,3,opt,name=sender_eph_pub,json=senderEphPub,proto3" json:"sender_eph_pub,omitempty"`
}

func (x *Packet_Init) Reset() {
	*x = Packet_Init{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Packet_Init) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet_Init) ProtoMessage() {}

func (x *Packet_Init) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet_Init.ProtoReflect.Descriptor instead.
func (*Packet_Init) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Packet_Init) GetSenderPeerId() []byte {
	if x != nil {
		return x.SenderPeerId
	}
	return nil
}

func (x *Packet_Init) GetReceiverPeerId() []byte {
	if x != nil {
		return x.ReceiverPeerId
	}
	return nil
}

func (x *Packet_Init) GetSenderEphPub() []byte {
	if x != nil {
		return x.SenderEphPub
	}
	return nil
}

type Packet_InitAck struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// SenderEphPub is the ephemeral public key of the sender.
	// This is used to compute the shared secret and decode AckInner.
	SenderEphPub []byte `protobuf:"bytes,1,opt,name=sender_eph_pub,json=senderEphPub,proto3" json:"sender_eph_pub,omitempty"`
	// Ciphertext is a Ciphertext message encoded and encrypted with the shared key.
	Ciphertext []byte `protobuf:"bytes,2,opt,name=ciphertext,proto3" json:"ciphertext,omitempty"`
}

func (x *Packet_InitAck) Reset() {
	*x = Packet_InitAck{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Packet_InitAck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet_InitAck) ProtoMessage() {}

func (x *Packet_InitAck) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet_InitAck.ProtoReflect.Descriptor instead.
func (*Packet_InitAck) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Packet_InitAck) GetSenderEphPub() []byte {
	if x != nil {
		return x.SenderEphPub
	}
	return nil
}

func (x *Packet_InitAck) GetCiphertext() []byte {
	if x != nil {
		return x.Ciphertext
	}
	return nil
}

type Packet_Complete struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Ciphertext is a Ciphertext message encoded and encrypted with the shared key.
	Ciphertext []byte `protobuf:"bytes,1,opt,name=ciphertext,proto3" json:"ciphertext,omitempty"`
}

func (x *Packet_Complete) Reset() {
	*x = Packet_Complete{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Packet_Complete) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet_Complete) ProtoMessage() {}

func (x *Packet_Complete) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet_Complete.ProtoReflect.Descriptor instead.
func (*Packet_Complete) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescGZIP(), []int{0, 2}
}

func (x *Packet_Complete) GetCiphertext() []byte {
	if x != nil {
		return x.Ciphertext
	}
	return nil
}

type Packet_Ciphertext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TupleSignature is the signature of the two ephemeral pub keys.
	// The signature is made using the sender's public key.
	// The keys are concatinated as AB
	TupleSignature []byte `protobuf:"bytes,1,opt,name=tuple_signature,json=tupleSignature,proto3" json:"tuple_signature,omitempty"`
	// SenderPubKey contains B's public key if necessary.
	SenderPubKey []byte `protobuf:"bytes,2,opt,name=sender_pub_key,json=senderPubKey,proto3" json:"sender_pub_key,omitempty"`
	// ReceiverKeyKnown indicates that A's public key is known.
	ReceiverKeyKnown bool `protobuf:"varint,3,opt,name=receiver_key_known,json=receiverKeyKnown,proto3" json:"receiver_key_known,omitempty"`
	// ExtraInfo contains extra information supplied by the transport.
	// Example: in UDP this is information about what port to dial KCP on.
	ExtraInfo []byte `protobuf:"bytes,4,opt,name=extra_info,json=extraInfo,proto3" json:"extra_info,omitempty"`
}

func (x *Packet_Ciphertext) Reset() {
	*x = Packet_Ciphertext{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Packet_Ciphertext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet_Ciphertext) ProtoMessage() {}

func (x *Packet_Ciphertext) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet_Ciphertext.ProtoReflect.Descriptor instead.
func (*Packet_Ciphertext) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescGZIP(), []int{0, 3}
}

func (x *Packet_Ciphertext) GetTupleSignature() []byte {
	if x != nil {
		return x.TupleSignature
	}
	return nil
}

func (x *Packet_Ciphertext) GetSenderPubKey() []byte {
	if x != nil {
		return x.SenderPubKey
	}
	return nil
}

func (x *Packet_Ciphertext) GetReceiverKeyKnown() bool {
	if x != nil {
		return x.ReceiverKeyKnown
	}
	return false
}

func (x *Packet_Ciphertext) GetExtraInfo() []byte {
	if x != nil {
		return x.ExtraInfo
	}
	return nil
}

var File_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDesc = []byte{
	0x0a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x68, 0x61, 0x6e, 0x64, 0x73, 0x68, 0x61, 0x6b, 0x65, 0x2f,
	0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x2f, 0x73, 0x32, 0x73, 0x2f, 0x73, 0x32, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x73, 0x32, 0x73, 0x22, 0xfd, 0x04, 0x0a, 0x06,
	0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x30, 0x0a, 0x0b, 0x70, 0x61, 0x63, 0x6b, 0x65, 0x74,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0f, 0x2e, 0x73, 0x32,
	0x73, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x70, 0x61,
	0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x2b, 0x0a, 0x08, 0x69, 0x6e, 0x69, 0x74,
	0x5f, 0x70, 0x6b, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x73, 0x32, 0x73,
	0x2e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x49, 0x6e, 0x69, 0x74, 0x52, 0x07, 0x69, 0x6e,
	0x69, 0x74, 0x50, 0x6b, 0x74, 0x12, 0x35, 0x0a, 0x0c, 0x69, 0x6e, 0x69, 0x74, 0x5f, 0x61, 0x63,
	0x6b, 0x5f, 0x70, 0x6b, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x73, 0x32,
	0x73, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x49, 0x6e, 0x69, 0x74, 0x41, 0x63, 0x6b,
	0x52, 0x0a, 0x69, 0x6e, 0x69, 0x74, 0x41, 0x63, 0x6b, 0x50, 0x6b, 0x74, 0x12, 0x37, 0x0a, 0x0c,
	0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x5f, 0x70, 0x6b, 0x74, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x73, 0x32, 0x73, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x2e,
	0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x0b, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65,
	0x74, 0x65, 0x50, 0x6b, 0x74, 0x1a, 0x7c, 0x0a, 0x04, 0x49, 0x6e, 0x69, 0x74, 0x12, 0x24, 0x0a,
	0x0e, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x50, 0x65, 0x65,
	0x72, 0x49, 0x64, 0x12, 0x28, 0x0a, 0x10, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x5f,
	0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x72,
	0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x24, 0x0a,
	0x0e, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x65, 0x70, 0x68, 0x5f, 0x70, 0x75, 0x62, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x45, 0x70, 0x68,
	0x50, 0x75, 0x62, 0x1a, 0x4f, 0x0a, 0x07, 0x49, 0x6e, 0x69, 0x74, 0x41, 0x63, 0x6b, 0x12, 0x24,
	0x0a, 0x0e, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x65, 0x70, 0x68, 0x5f, 0x70, 0x75, 0x62,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0c, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x45, 0x70,
	0x68, 0x50, 0x75, 0x62, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65,
	0x78, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72,
	0x74, 0x65, 0x78, 0x74, 0x1a, 0x2a, 0x0a, 0x08, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x65,
	0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x63, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74,
	0x1a, 0xa8, 0x01, 0x0a, 0x0a, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72, 0x74, 0x65, 0x78, 0x74, 0x12,
	0x27, 0x0a, 0x0f, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x5f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0e, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x53,
	0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x73, 0x65, 0x6e, 0x64,
	0x65, 0x72, 0x5f, 0x70, 0x75, 0x62, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x0c, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x12, 0x2c,
	0x0a, 0x12, 0x72, 0x65, 0x63, 0x65, 0x69, 0x76, 0x65, 0x72, 0x5f, 0x6b, 0x65, 0x79, 0x5f, 0x6b,
	0x6e, 0x6f, 0x77, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10, 0x72, 0x65, 0x63, 0x65,
	0x69, 0x76, 0x65, 0x72, 0x4b, 0x65, 0x79, 0x4b, 0x6e, 0x6f, 0x77, 0x6e, 0x12, 0x1d, 0x0a, 0x0a,
	0x65, 0x78, 0x74, 0x72, 0x61, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x09, 0x65, 0x78, 0x74, 0x72, 0x61, 0x49, 0x6e, 0x66, 0x6f, 0x2a, 0x53, 0x0a, 0x0a, 0x50,
	0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x13, 0x0a, 0x0f, 0x50, 0x61, 0x63,
	0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x49, 0x4e, 0x49, 0x54, 0x10, 0x00, 0x12, 0x17,
	0x0a, 0x13, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x49, 0x4e, 0x49,
	0x54, 0x5f, 0x41, 0x43, 0x4b, 0x10, 0x01, 0x12, 0x17, 0x0a, 0x13, 0x50, 0x61, 0x63, 0x6b, 0x65,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x43, 0x4f, 0x4d, 0x50, 0x4c, 0x45, 0x54, 0x45, 0x10, 0x02,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescData = file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_goTypes = []interface{}{
	(PacketType)(0),           // 0: s2s.PacketType
	(*Packet)(nil),            // 1: s2s.Packet
	(*Packet_Init)(nil),       // 2: s2s.Packet.Init
	(*Packet_InitAck)(nil),    // 3: s2s.Packet.InitAck
	(*Packet_Complete)(nil),   // 4: s2s.Packet.Complete
	(*Packet_Ciphertext)(nil), // 5: s2s.Packet.Ciphertext
}
var file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_depIdxs = []int32{
	0, // 0: s2s.Packet.packet_type:type_name -> s2s.PacketType
	2, // 1: s2s.Packet.init_pkt:type_name -> s2s.Packet.Init
	3, // 2: s2s.Packet.init_ack_pkt:type_name -> s2s.Packet.InitAck
	4, // 3: s2s.Packet.complete_pkt:type_name -> s2s.Packet.Complete
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_init() }
func file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_init() {
	if File_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Packet); i {
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
		file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Packet_Init); i {
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
		file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Packet_InitAck); i {
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
		file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Packet_Complete); i {
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
		file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Packet_Ciphertext); i {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_depIdxs,
		EnumInfos:         file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_enumTypes,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto = out.File
	file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_handshake_identity_s2s_s2s_proto_depIdxs = nil
}
