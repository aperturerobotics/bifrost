// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/bifrost/link/establish/config.proto

package link_establish_controller

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
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

// Config is the link establish controller config.
// The establish controller attempts to establish links with configured peers.
type Config struct {
	// PeerIds is the list of peer IDs to attempt to establish links to.
	PeerIds              []string `protobuf:"bytes,1,rep,name=peer_ids,json=peerIds,proto3" json:"peer_ids,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_10c51e3ba4c13894, []int{0}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetPeerIds() []string {
	if m != nil {
		return m.PeerIds
	}
	return nil
}

func init() {
	proto.RegisterType((*Config)(nil), "link.establish.controller.Config")
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/bifrost/link/establish/config.proto", fileDescriptor_10c51e3ba4c13894)
}

var fileDescriptor_10c51e3ba4c13894 = []byte{
	// 143 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0xcc, 0xb1, 0x0e, 0x82, 0x40,
	0x0c, 0xc6, 0xf1, 0x18, 0x13, 0x54, 0x46, 0x26, 0xd9, 0x8c, 0x2e, 0x4e, 0xd7, 0xc1, 0x07, 0x70,
	0x70, 0x72, 0xf5, 0x05, 0x0c, 0x3d, 0x0a, 0x34, 0x9e, 0xd7, 0x4b, 0x5b, 0xde, 0xdf, 0xc0, 0xe0,
	0xf8, 0xe5, 0xf7, 0xe5, 0x5f, 0xdf, 0x47, 0xf6, 0x69, 0xc6, 0x10, 0xe5, 0x0b, 0x5d, 0x21, 0xf5,
	0x59, 0x49, 0x05, 0xc5, 0x39, 0x1a, 0x20, 0x0f, 0x2a, 0xe6, 0x90, 0x38, 0x7f, 0x80, 0xcc, 0x3b,
	0x4c, 0x6c, 0x13, 0x44, 0xc9, 0x03, 0x8f, 0xa1, 0xa8, 0xb8, 0x34, 0xed, 0x82, 0xe1, 0x8f, 0x21,
	0x4a, 0x76, 0x95, 0x94, 0x48, 0xcf, 0x97, 0xba, 0x7a, 0xac, 0xd7, 0xa6, 0xad, 0xf7, 0x85, 0x48,
	0xdf, 0xdc, 0xdb, 0x71, 0x73, 0xda, 0x5e, 0x0f, 0xaf, 0xdd, 0xb2, 0x9f, 0xbd, 0x61, 0xb5, 0x66,
	0x6e, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb7, 0x42, 0x57, 0xba, 0x89, 0x00, 0x00, 0x00,
}