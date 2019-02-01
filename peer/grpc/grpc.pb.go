// Code generated by protoc-gen-go. DO NOT EDIT.
// source: github.com/aperturerobotics/bifrost/peer/grpc/grpc.proto

package peer_grpc

import (
	context "context"
	fmt "fmt"
	controller "github.com/aperturerobotics/bifrost/peer/controller"
	exec "github.com/aperturerobotics/controllerbus/controller/exec"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
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

// IdentifyRequest is a request to load an identity.
type IdentifyRequest struct {
	// Config is the request to configure the peer controller.
	Config               *controller.Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *IdentifyRequest) Reset()         { *m = IdentifyRequest{} }
func (m *IdentifyRequest) String() string { return proto.CompactTextString(m) }
func (*IdentifyRequest) ProtoMessage()    {}
func (*IdentifyRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8843ac85db7ceb86, []int{0}
}

func (m *IdentifyRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IdentifyRequest.Unmarshal(m, b)
}
func (m *IdentifyRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IdentifyRequest.Marshal(b, m, deterministic)
}
func (m *IdentifyRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IdentifyRequest.Merge(m, src)
}
func (m *IdentifyRequest) XXX_Size() int {
	return xxx_messageInfo_IdentifyRequest.Size(m)
}
func (m *IdentifyRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_IdentifyRequest.DiscardUnknown(m)
}

var xxx_messageInfo_IdentifyRequest proto.InternalMessageInfo

func (m *IdentifyRequest) GetConfig() *controller.Config {
	if m != nil {
		return m.Config
	}
	return nil
}

// IdentifyResponse is a response to an identify request.
type IdentifyResponse struct {
	// ControllerStatus is the status of the peer controller.
	ControllerStatus     exec.ControllerStatus `protobuf:"varint,1,opt,name=controller_status,json=controllerStatus,proto3,enum=controller.exec.ControllerStatus" json:"controller_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *IdentifyResponse) Reset()         { *m = IdentifyResponse{} }
func (m *IdentifyResponse) String() string { return proto.CompactTextString(m) }
func (*IdentifyResponse) ProtoMessage()    {}
func (*IdentifyResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8843ac85db7ceb86, []int{1}
}

func (m *IdentifyResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IdentifyResponse.Unmarshal(m, b)
}
func (m *IdentifyResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IdentifyResponse.Marshal(b, m, deterministic)
}
func (m *IdentifyResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IdentifyResponse.Merge(m, src)
}
func (m *IdentifyResponse) XXX_Size() int {
	return xxx_messageInfo_IdentifyResponse.Size(m)
}
func (m *IdentifyResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_IdentifyResponse.DiscardUnknown(m)
}

var xxx_messageInfo_IdentifyResponse proto.InternalMessageInfo

func (m *IdentifyResponse) GetControllerStatus() exec.ControllerStatus {
	if m != nil {
		return m.ControllerStatus
	}
	return exec.ControllerStatus_ControllerStatus_UNKNOWN
}

// GetPeerInfoRequest is the request type for GetPeerInfo.
type GetPeerInfoRequest struct {
	// PeerId restricts the response to a specific peer ID.
	PeerId               string   `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetPeerInfoRequest) Reset()         { *m = GetPeerInfoRequest{} }
func (m *GetPeerInfoRequest) String() string { return proto.CompactTextString(m) }
func (*GetPeerInfoRequest) ProtoMessage()    {}
func (*GetPeerInfoRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_8843ac85db7ceb86, []int{2}
}

func (m *GetPeerInfoRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPeerInfoRequest.Unmarshal(m, b)
}
func (m *GetPeerInfoRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPeerInfoRequest.Marshal(b, m, deterministic)
}
func (m *GetPeerInfoRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPeerInfoRequest.Merge(m, src)
}
func (m *GetPeerInfoRequest) XXX_Size() int {
	return xxx_messageInfo_GetPeerInfoRequest.Size(m)
}
func (m *GetPeerInfoRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPeerInfoRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetPeerInfoRequest proto.InternalMessageInfo

func (m *GetPeerInfoRequest) GetPeerId() string {
	if m != nil {
		return m.PeerId
	}
	return ""
}

// PeerInfo is basic information about a peer.
type PeerInfo struct {
	// PeerId is the b58 peer ID.
	PeerId               string   `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeerInfo) Reset()         { *m = PeerInfo{} }
func (m *PeerInfo) String() string { return proto.CompactTextString(m) }
func (*PeerInfo) ProtoMessage()    {}
func (*PeerInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_8843ac85db7ceb86, []int{3}
}

func (m *PeerInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeerInfo.Unmarshal(m, b)
}
func (m *PeerInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeerInfo.Marshal(b, m, deterministic)
}
func (m *PeerInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeerInfo.Merge(m, src)
}
func (m *PeerInfo) XXX_Size() int {
	return xxx_messageInfo_PeerInfo.Size(m)
}
func (m *PeerInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_PeerInfo.DiscardUnknown(m)
}

var xxx_messageInfo_PeerInfo proto.InternalMessageInfo

func (m *PeerInfo) GetPeerId() string {
	if m != nil {
		return m.PeerId
	}
	return ""
}

// GetPeerInfoResponse is the response type for GetPeerInfo.
type GetPeerInfoResponse struct {
	// LocalPeers is the set of peers loaded.
	LocalPeers           []*PeerInfo `protobuf:"bytes,1,rep,name=local_peers,json=localPeers,proto3" json:"local_peers,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetPeerInfoResponse) Reset()         { *m = GetPeerInfoResponse{} }
func (m *GetPeerInfoResponse) String() string { return proto.CompactTextString(m) }
func (*GetPeerInfoResponse) ProtoMessage()    {}
func (*GetPeerInfoResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_8843ac85db7ceb86, []int{4}
}

func (m *GetPeerInfoResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPeerInfoResponse.Unmarshal(m, b)
}
func (m *GetPeerInfoResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPeerInfoResponse.Marshal(b, m, deterministic)
}
func (m *GetPeerInfoResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPeerInfoResponse.Merge(m, src)
}
func (m *GetPeerInfoResponse) XXX_Size() int {
	return xxx_messageInfo_GetPeerInfoResponse.Size(m)
}
func (m *GetPeerInfoResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPeerInfoResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetPeerInfoResponse proto.InternalMessageInfo

func (m *GetPeerInfoResponse) GetLocalPeers() []*PeerInfo {
	if m != nil {
		return m.LocalPeers
	}
	return nil
}

func init() {
	proto.RegisterType((*IdentifyRequest)(nil), "peer.grpc.IdentifyRequest")
	proto.RegisterType((*IdentifyResponse)(nil), "peer.grpc.IdentifyResponse")
	proto.RegisterType((*GetPeerInfoRequest)(nil), "peer.grpc.GetPeerInfoRequest")
	proto.RegisterType((*PeerInfo)(nil), "peer.grpc.PeerInfo")
	proto.RegisterType((*GetPeerInfoResponse)(nil), "peer.grpc.GetPeerInfoResponse")
}

func init() {
	proto.RegisterFile("github.com/aperturerobotics/bifrost/peer/grpc/grpc.proto", fileDescriptor_8843ac85db7ceb86)
}

var fileDescriptor_8843ac85db7ceb86 = []byte{
	// 339 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x91, 0xd1, 0x4a, 0xc3, 0x30,
	0x14, 0x86, 0x2d, 0xc2, 0xdc, 0x4e, 0x41, 0x67, 0x76, 0x31, 0xa9, 0x28, 0x5a, 0x6f, 0x76, 0x63,
	0x2a, 0xd5, 0x0b, 0x2f, 0xc5, 0x09, 0x63, 0x08, 0x43, 0xba, 0x07, 0x18, 0x6b, 0x76, 0x3a, 0x03,
	0xb5, 0xa9, 0x49, 0x2a, 0xfa, 0x42, 0x3e, 0xa7, 0x24, 0x5d, 0xd7, 0x3a, 0x9d, 0x78, 0x13, 0x72,
	0xce, 0xf9, 0xff, 0x2f, 0x3f, 0x27, 0x70, 0xbb, 0xe4, 0xfa, 0xb9, 0x88, 0x29, 0x13, 0x2f, 0xc1,
	0x3c, 0x47, 0xa9, 0x0b, 0x89, 0x52, 0xc4, 0x42, 0x73, 0xa6, 0x82, 0x98, 0x27, 0x52, 0x28, 0x1d,
	0xe4, 0x88, 0x32, 0x58, 0xca, 0x9c, 0xd9, 0x83, 0xe6, 0x52, 0x68, 0x41, 0x3a, 0xa6, 0x4b, 0x4d,
	0xc3, 0xbb, 0xfb, 0x37, 0x84, 0x89, 0x4c, 0x4b, 0x91, 0xa6, 0xe5, 0x35, 0xe1, 0xcb, 0x12, 0xe6,
	0x3d, 0xfc, 0x45, 0xa8, 0x4d, 0x71, 0xd1, 0xac, 0x02, 0x7c, 0x47, 0x66, 0x8f, 0x92, 0xe2, 0xdf,
	0xc3, 0xc1, 0x78, 0x81, 0x99, 0xe6, 0xc9, 0x47, 0x84, 0xaf, 0x05, 0x2a, 0x4d, 0x02, 0x68, 0x95,
	0x0f, 0x1d, 0x39, 0x67, 0xce, 0xc0, 0x0d, 0xfb, 0xd4, 0xc6, 0xae, 0x21, 0x74, 0x68, 0xc7, 0xd1,
	0x4a, 0xe6, 0xc7, 0xd0, 0xad, 0x19, 0x2a, 0x17, 0x99, 0x42, 0x32, 0x81, 0xc3, 0xda, 0x30, 0x53,
	0x7a, 0xae, 0x0b, 0x65, 0x79, 0xfb, 0xe1, 0x79, 0x13, 0x65, 0xa3, 0x0c, 0xd7, 0xf5, 0xd4, 0x0a,
	0xa3, 0x2e, 0xdb, 0xe8, 0xf8, 0x97, 0x40, 0x46, 0xa8, 0x9f, 0x10, 0xe5, 0x38, 0x4b, 0x44, 0x15,
	0xb5, 0x0f, 0x7b, 0x26, 0xdb, 0x8c, 0x2f, 0x2c, 0xbb, 0x13, 0xb5, 0x4c, 0x39, 0x5e, 0xf8, 0x17,
	0xd0, 0xae, 0xb4, 0xdb, 0x45, 0x8f, 0xd0, 0xfb, 0xc6, 0x5c, 0x45, 0xbf, 0x01, 0x37, 0x15, 0x6c,
	0x9e, 0xce, 0x8c, 0xcc, 0x84, 0xde, 0x1d, 0xb8, 0x61, 0x8f, 0xae, 0xff, 0x8e, 0xae, 0x1d, 0x60,
	0x75, 0xa6, 0x54, 0xe1, 0xa7, 0x03, 0xae, 0xb9, 0x4d, 0x51, 0xbe, 0x71, 0x86, 0x64, 0x04, 0xed,
	0x6a, 0x29, 0xc4, 0x6b, 0x98, 0x37, 0xb6, 0xed, 0x1d, 0xff, 0x3a, 0x2b, 0xa3, 0xf8, 0x3b, 0x57,
	0x0e, 0x99, 0x80, 0xdb, 0x48, 0x49, 0x4e, 0x1a, 0xfa, 0x9f, 0x1b, 0xf1, 0x4e, 0xb7, 0x8d, 0x2b,
	0x62, 0xdc, 0xb2, 0x1f, 0x7f, 0xfd, 0x15, 0x00, 0x00, 0xff, 0xff, 0x88, 0x15, 0x29, 0x98, 0xc7,
	0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PeerServiceClient is the client API for PeerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PeerServiceClient interface {
	// Identify loads and manages a private key identity.
	Identify(ctx context.Context, in *IdentifyRequest, opts ...grpc.CallOption) (PeerService_IdentifyClient, error)
	// GetPeerInfo returns information about attached peers.
	GetPeerInfo(ctx context.Context, in *GetPeerInfoRequest, opts ...grpc.CallOption) (*GetPeerInfoResponse, error)
}

type peerServiceClient struct {
	cc *grpc.ClientConn
}

func NewPeerServiceClient(cc *grpc.ClientConn) PeerServiceClient {
	return &peerServiceClient{cc}
}

func (c *peerServiceClient) Identify(ctx context.Context, in *IdentifyRequest, opts ...grpc.CallOption) (PeerService_IdentifyClient, error) {
	stream, err := c.cc.NewStream(ctx, &_PeerService_serviceDesc.Streams[0], "/peer.grpc.PeerService/Identify", opts...)
	if err != nil {
		return nil, err
	}
	x := &peerServiceIdentifyClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PeerService_IdentifyClient interface {
	Recv() (*IdentifyResponse, error)
	grpc.ClientStream
}

type peerServiceIdentifyClient struct {
	grpc.ClientStream
}

func (x *peerServiceIdentifyClient) Recv() (*IdentifyResponse, error) {
	m := new(IdentifyResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *peerServiceClient) GetPeerInfo(ctx context.Context, in *GetPeerInfoRequest, opts ...grpc.CallOption) (*GetPeerInfoResponse, error) {
	out := new(GetPeerInfoResponse)
	err := c.cc.Invoke(ctx, "/peer.grpc.PeerService/GetPeerInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PeerServiceServer is the server API for PeerService service.
type PeerServiceServer interface {
	// Identify loads and manages a private key identity.
	Identify(*IdentifyRequest, PeerService_IdentifyServer) error
	// GetPeerInfo returns information about attached peers.
	GetPeerInfo(context.Context, *GetPeerInfoRequest) (*GetPeerInfoResponse, error)
}

func RegisterPeerServiceServer(s *grpc.Server, srv PeerServiceServer) {
	s.RegisterService(&_PeerService_serviceDesc, srv)
}

func _PeerService_Identify_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(IdentifyRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PeerServiceServer).Identify(m, &peerServiceIdentifyServer{stream})
}

type PeerService_IdentifyServer interface {
	Send(*IdentifyResponse) error
	grpc.ServerStream
}

type peerServiceIdentifyServer struct {
	grpc.ServerStream
}

func (x *peerServiceIdentifyServer) Send(m *IdentifyResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _PeerService_GetPeerInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPeerInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PeerServiceServer).GetPeerInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peer.grpc.PeerService/GetPeerInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PeerServiceServer).GetPeerInfo(ctx, req.(*GetPeerInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PeerService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "peer.grpc.PeerService",
	HandlerType: (*PeerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetPeerInfo",
			Handler:    _PeerService_GetPeerInfo_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Identify",
			Handler:       _PeerService_Identify_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "github.com/aperturerobotics/bifrost/peer/grpc/grpc.proto",
}
