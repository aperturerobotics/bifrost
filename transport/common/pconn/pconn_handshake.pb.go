// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/bifrost/transport/common/pconn/pconn_handshake.proto

package pconn

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// HandshakeExtraData contains the extra data field of the pconn handshake.
type HandshakeExtraData struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HandshakeExtraData) Reset()         { *m = HandshakeExtraData{} }
func (m *HandshakeExtraData) String() string { return proto.CompactTextString(m) }
func (*HandshakeExtraData) ProtoMessage()    {}
func (*HandshakeExtraData) Descriptor() ([]byte, []int) {
	return fileDescriptor_pconn_handshake_3a39680144b60428, []int{0}
}
func (m *HandshakeExtraData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HandshakeExtraData.Unmarshal(m, b)
}
func (m *HandshakeExtraData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HandshakeExtraData.Marshal(b, m, deterministic)
}
func (dst *HandshakeExtraData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HandshakeExtraData.Merge(dst, src)
}
func (m *HandshakeExtraData) XXX_Size() int {
	return xxx_messageInfo_HandshakeExtraData.Size(m)
}
func (m *HandshakeExtraData) XXX_DiscardUnknown() {
	xxx_messageInfo_HandshakeExtraData.DiscardUnknown(m)
}

var xxx_messageInfo_HandshakeExtraData proto.InternalMessageInfo

func init() {
	proto.RegisterType((*HandshakeExtraData)(nil), "pconn.HandshakeExtraData")
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/bifrost/transport/common/pconn/pconn_handshake.proto", fileDescriptor_pconn_handshake_3a39680144b60428)
}

var fileDescriptor_pconn_handshake_3a39680144b60428 = []byte{
	// 123 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0x8a, 0x31, 0x0a, 0x02, 0x31,
	0x14, 0x05, 0x2b, 0x2d, 0xb6, 0x5c, 0x3c, 0x81, 0x07, 0xd8, 0x5f, 0x78, 0x05, 0x05, 0x4b, 0x6f,
	0x20, 0x3f, 0x31, 0x9a, 0x20, 0xc9, 0x0b, 0x2f, 0x6f, 0xc1, 0xe3, 0x0b, 0x8b, 0x36, 0x53, 0xcc,
	0xcc, 0x74, 0x7b, 0x15, 0xe5, 0x35, 0x2c, 0x11, 0xd5, 0xbc, 0x27, 0x6a, 0x65, 0x22, 0x02, 0x54,
	0xe2, 0xb0, 0x50, 0x9e, 0xc4, 0x90, 0x89, 0xde, 0x46, 0x07, 0x65, 0x11, 0xb5, 0xa2, 0x59, 0x8f,
	0x68, 0x3f, 0xde, 0xb3, 0xb7, 0xc7, 0xc8, 0xfe, 0x4e, 0x4b, 0x27, 0x84, 0x79, 0xb7, 0xe9, 0xe3,
	0x61, 0x9a, 0xaf, 0xff, 0x72, 0xf9, 0x88, 0x7e, 0x76, 0x79, 0xd8, 0x6f, 0xcf, 0xe9, 0x1b, 0x00,
	0x00, 0xff, 0xff, 0x9d, 0x02, 0x2f, 0xf2, 0x77, 0x00, 0x00, 0x00,
}
