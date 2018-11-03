// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/bifrost/transport/common/pconn/packet_type.proto

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

// PacketType is a one-byte trailer indicating the type of packet.
type PacketType int32

const (
	PacketType_PacketType_HANDSHAKE PacketType = 0
	PacketType_PacketType_RAW       PacketType = 1
	PacketType_PacketType_KCP_SMUX  PacketType = 2
)

var PacketType_name = map[int32]string{
	0: "PacketType_HANDSHAKE",
	1: "PacketType_RAW",
	2: "PacketType_KCP_SMUX",
}
var PacketType_value = map[string]int32{
	"PacketType_HANDSHAKE": 0,
	"PacketType_RAW":       1,
	"PacketType_KCP_SMUX":  2,
}

func (x PacketType) String() string {
	return proto.EnumName(PacketType_name, int32(x))
}
func (PacketType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_packet_type_4165e57a2fb78752, []int{0}
}

func init() {
	proto.RegisterEnum("pconn.PacketType", PacketType_name, PacketType_value)
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/bifrost/transport/common/pconn/packet_type.proto", fileDescriptor_packet_type_4165e57a2fb78752)
}

var fileDescriptor_packet_type_4165e57a2fb78752 = []byte{
	// 167 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xf2, 0x49, 0xcf, 0x2c, 0xc9,
	0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5, 0x4f, 0x2c, 0x48, 0x2d, 0x2a, 0x29, 0x2d, 0x4a, 0x2d,
	0xca, 0x4f, 0xca, 0x2f, 0xc9, 0x4c, 0x2e, 0xd6, 0x4f, 0xca, 0x4c, 0x2b, 0xca, 0x2f, 0x2e, 0xd1,
	0x2f, 0x29, 0x4a, 0xcc, 0x2b, 0x2e, 0xc8, 0x2f, 0x2a, 0xd1, 0x4f, 0xce, 0xcf, 0xcd, 0xcd, 0xcf,
	0xd3, 0x2f, 0x48, 0xce, 0xcf, 0xcb, 0xd3, 0x2f, 0x48, 0x4c, 0xce, 0x4e, 0x2d, 0x89, 0x2f, 0xa9,
	0x2c, 0x48, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x4b, 0x68, 0x05, 0x73, 0x71,
	0x05, 0x80, 0xe5, 0x42, 0x2a, 0x0b, 0x52, 0x85, 0x24, 0xb8, 0x44, 0x10, 0xbc, 0x78, 0x0f, 0x47,
	0x3f, 0x97, 0x60, 0x0f, 0x47, 0x6f, 0x57, 0x01, 0x06, 0x21, 0x21, 0x2e, 0x3e, 0x24, 0x99, 0x20,
	0xc7, 0x70, 0x01, 0x46, 0x21, 0x71, 0x2e, 0x61, 0x24, 0x31, 0x6f, 0xe7, 0x80, 0xf8, 0x60, 0xdf,
	0xd0, 0x08, 0x01, 0xa6, 0x24, 0x36, 0xb0, 0x15, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x4a,
	0x29, 0x21, 0x64, 0xb2, 0x00, 0x00, 0x00,
}
