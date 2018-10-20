// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/bifrost/transport/udp/udp.proto

package udp

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

// Config is the configuration for the udp transport.
type Config struct {
	// NodePeerID constrains the node peer ID.
	// If empty, attaches to whatever node is running.
	NodePeerId string `protobuf:"bytes,1,opt,name=node_peer_id,json=nodePeerId" json:"node_peer_id,omitempty"`
	// ListenAddr contains the address to listen on.
	// Has no effect in the browser.
	ListenAddr           string   `protobuf:"bytes,2,opt,name=listen_addr,json=listenAddr" json:"listen_addr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_udp_74fffb96f44da465, []int{0}
}
func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (dst *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(dst, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetNodePeerId() string {
	if m != nil {
		return m.NodePeerId
	}
	return ""
}

func (m *Config) GetListenAddr() string {
	if m != nil {
		return m.ListenAddr
	}
	return ""
}

func init() {
	proto.RegisterType((*Config)(nil), "udp.Config")
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/bifrost/transport/udp/udp.proto", fileDescriptor_udp_74fffb96f44da465)
}

var fileDescriptor_udp_74fffb96f44da465 = []byte{
	// 155 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x24, 0xcc, 0xb1, 0xaa, 0xc2, 0x40,
	0x10, 0x85, 0x61, 0x72, 0x2f, 0x04, 0x5c, 0xad, 0x52, 0xa5, 0x33, 0x58, 0x59, 0x65, 0x0b, 0x4b,
	0x2b, 0xb1, 0x12, 0x1b, 0xf1, 0x05, 0xc2, 0x6e, 0x66, 0x12, 0x07, 0x74, 0x67, 0x98, 0x9d, 0x7d,
	0x7f, 0x89, 0x29, 0x4e, 0xf3, 0x71, 0xf8, 0xdd, 0x79, 0x26, 0x7b, 0x95, 0xd8, 0x8f, 0xfc, 0xf1,
	0x41, 0x50, 0xad, 0x28, 0x2a, 0x47, 0x36, 0x1a, 0xb3, 0x8f, 0x34, 0x29, 0x67, 0xf3, 0xa6, 0x21,
	0x65, 0x61, 0x35, 0x5f, 0x40, 0x96, 0xf5, 0xa2, 0x6c, 0xdc, 0xfc, 0x17, 0x90, 0xc3, 0xdd, 0xd5,
	0x57, 0x4e, 0x13, 0xcd, 0x4d, 0xe7, 0x76, 0x89, 0x01, 0x07, 0x41, 0xd4, 0x81, 0xa0, 0xad, 0xba,
	0xea, 0xb8, 0x79, 0xba, 0xc5, 0x1e, 0x88, 0x7a, 0x83, 0x66, 0xef, 0xb6, 0x6f, 0xca, 0x86, 0x69,
	0x08, 0x00, 0xda, 0xfe, 0xad, 0x87, 0x95, 0x2e, 0x00, 0x1a, 0xeb, 0x5f, 0xf8, 0xf4, 0x0d, 0x00,
	0x00, 0xff, 0xff, 0x16, 0xd9, 0x2e, 0x8b, 0x97, 0x00, 0x00, 0x00,
}
