// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/bifrost/transport/common/dialer/dialer.proto

package dialer

import (
	fmt "fmt"
	backoff "github.com/aperturerobotics/bifrost/util/backoff"
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

// DialerOpts contains options relating to dialing a statically configured peer.
type DialerOpts struct {
	// Address is the address of the peer, in the format expected by the transport.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// Backoff is the dialing backoff configuration.
	// Can be empty.
	Backoff              *backoff.Backoff `protobuf:"bytes,2,opt,name=backoff,proto3" json:"backoff,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *DialerOpts) Reset()         { *m = DialerOpts{} }
func (m *DialerOpts) String() string { return proto.CompactTextString(m) }
func (*DialerOpts) ProtoMessage()    {}
func (*DialerOpts) Descriptor() ([]byte, []int) {
	return fileDescriptor_df3d358221dde688, []int{0}
}

func (m *DialerOpts) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DialerOpts.Unmarshal(m, b)
}
func (m *DialerOpts) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DialerOpts.Marshal(b, m, deterministic)
}
func (m *DialerOpts) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DialerOpts.Merge(m, src)
}
func (m *DialerOpts) XXX_Size() int {
	return xxx_messageInfo_DialerOpts.Size(m)
}
func (m *DialerOpts) XXX_DiscardUnknown() {
	xxx_messageInfo_DialerOpts.DiscardUnknown(m)
}

var xxx_messageInfo_DialerOpts proto.InternalMessageInfo

func (m *DialerOpts) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *DialerOpts) GetBackoff() *backoff.Backoff {
	if m != nil {
		return m.Backoff
	}
	return nil
}

func init() {
	proto.RegisterType((*DialerOpts)(nil), "dialer.DialerOpts")
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/bifrost/transport/common/dialer/dialer.proto", fileDescriptor_df3d358221dde688)
}

var fileDescriptor_df3d358221dde688 = []byte{
	// 172 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0xcc, 0x31, 0xcb, 0xc2, 0x30,
	0x10, 0xc6, 0x71, 0xfa, 0x0e, 0x2d, 0x6f, 0x5c, 0xa4, 0x53, 0x71, 0x2a, 0x4e, 0xc5, 0xa1, 0x07,
	0xba, 0x3b, 0x88, 0x83, 0x9b, 0xd0, 0x6f, 0x90, 0xa4, 0xa9, 0x06, 0xdb, 0x5e, 0xb8, 0xbb, 0x7e,
	0x7f, 0xa1, 0x69, 0x76, 0xa7, 0x3f, 0x0f, 0xdc, 0xfd, 0xd4, 0xe3, 0xe5, 0xe5, 0xbd, 0x98, 0xd6,
	0xe2, 0x04, 0x3a, 0x38, 0x92, 0x85, 0x1c, 0xa1, 0x41, 0xf1, 0x96, 0xc1, 0xf8, 0x81, 0x90, 0x05,
	0x84, 0xf4, 0xcc, 0x01, 0x49, 0xc0, 0xe2, 0x34, 0xe1, 0x0c, 0xbd, 0xd7, 0xa3, 0xa3, 0x2d, 0x6d,
	0x20, 0x14, 0x2c, 0xf3, 0xb8, 0x0e, 0xd7, 0x5f, 0xc4, 0x45, 0xfc, 0x08, 0x46, 0xdb, 0x0f, 0x0e,
	0x43, 0x6a, 0x74, 0x8e, 0x9d, 0x52, 0xf7, 0x55, 0x7a, 0x06, 0xe1, 0xb2, 0x52, 0x85, 0xee, 0x7b,
	0x72, 0xcc, 0x55, 0x56, 0x67, 0xcd, 0x7f, 0x97, 0x66, 0x79, 0x52, 0xc5, 0xf6, 0x58, 0xfd, 0xd5,
	0x59, 0xb3, 0x3b, 0xef, 0xdb, 0x04, 0xdd, 0x62, 0xbb, 0x74, 0x60, 0xf2, 0x95, 0xbe, 0x7c, 0x03,
	0x00, 0x00, 0xff, 0xff, 0x02, 0xca, 0xf8, 0xaa, 0xee, 0x00, 0x00, 0x00,
}
