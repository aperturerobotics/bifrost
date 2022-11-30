// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1-devel
// 	protoc        v3.21.9
// source: github.com/aperturerobotics/bifrost/peer/api/api.proto

package peer_api

import (
	reflect "reflect"
	sync "sync"

	controller "github.com/aperturerobotics/bifrost/peer/controller"
	exec "github.com/aperturerobotics/controllerbus/controller/exec"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// IdentifyRequest is a request to load an identity.
type IdentifyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Config is the request to configure the peer controller.
	Config *controller.Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *IdentifyRequest) Reset() {
	*x = IdentifyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IdentifyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IdentifyRequest) ProtoMessage() {}

func (x *IdentifyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IdentifyRequest.ProtoReflect.Descriptor instead.
func (*IdentifyRequest) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescGZIP(), []int{0}
}

func (x *IdentifyRequest) GetConfig() *controller.Config {
	if x != nil {
		return x.Config
	}
	return nil
}

// IdentifyResponse is a response to an identify request.
type IdentifyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ControllerStatus is the status of the peer controller.
	ControllerStatus exec.ControllerStatus `protobuf:"varint,1,opt,name=controller_status,json=controllerStatus,proto3,enum=controller.exec.ControllerStatus" json:"controller_status,omitempty"`
}

func (x *IdentifyResponse) Reset() {
	*x = IdentifyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IdentifyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IdentifyResponse) ProtoMessage() {}

func (x *IdentifyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IdentifyResponse.ProtoReflect.Descriptor instead.
func (*IdentifyResponse) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescGZIP(), []int{1}
}

func (x *IdentifyResponse) GetControllerStatus() exec.ControllerStatus {
	if x != nil {
		return x.ControllerStatus
	}
	return exec.ControllerStatus(0)
}

// GetPeerInfoRequest is the request type for GetPeerInfo.
type GetPeerInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// PeerId restricts the response to a specific peer ID.
	PeerId string `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
}

func (x *GetPeerInfoRequest) Reset() {
	*x = GetPeerInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPeerInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPeerInfoRequest) ProtoMessage() {}

func (x *GetPeerInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPeerInfoRequest.ProtoReflect.Descriptor instead.
func (*GetPeerInfoRequest) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescGZIP(), []int{2}
}

func (x *GetPeerInfoRequest) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

// PeerInfo is basic information about a peer.
type PeerInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// PeerId is the b58 peer ID.
	PeerId string `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
}

func (x *PeerInfo) Reset() {
	*x = PeerInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerInfo) ProtoMessage() {}

func (x *PeerInfo) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerInfo.ProtoReflect.Descriptor instead.
func (*PeerInfo) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescGZIP(), []int{3}
}

func (x *PeerInfo) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

// GetPeerInfoResponse is the response type for GetPeerInfo.
type GetPeerInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// LocalPeers is the set of peers loaded.
	LocalPeers []*PeerInfo `protobuf:"bytes,1,rep,name=local_peers,json=localPeers,proto3" json:"local_peers,omitempty"`
}

func (x *GetPeerInfoResponse) Reset() {
	*x = GetPeerInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPeerInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPeerInfoResponse) ProtoMessage() {}

func (x *GetPeerInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPeerInfoResponse.ProtoReflect.Descriptor instead.
func (*GetPeerInfoResponse) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescGZIP(), []int{4}
}

func (x *GetPeerInfoResponse) GetLocalPeers() []*PeerInfo {
	if x != nil {
		return x.LocalPeers
	}
	return nil
}

var File_github_com_aperturerobotics_bifrost_peer_api_api_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDesc = []byte{
	0x0a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x70, 0x65, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61,
	0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x70, 0x65, 0x65, 0x72, 0x2e, 0x61,
	0x70, 0x69, 0x1a, 0x40, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61,
	0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f,
	0x62, 0x69, 0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x70, 0x65, 0x65, 0x72, 0x2f, 0x63, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x61, 0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63,
	0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x62, 0x75, 0x73, 0x2f,
	0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x65, 0x78, 0x65, 0x63, 0x2f,
	0x65, 0x78, 0x65, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x42, 0x0a, 0x0f, 0x49, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2f, 0x0a,
	0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e,
	0x70, 0x65, 0x65, 0x72, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x62,
	0x0a, 0x10, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x4e, 0x0a, 0x11, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72,
	0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e,
	0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x65, 0x78, 0x65, 0x63, 0x2e,
	0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x10, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x22, 0x2d, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x50, 0x65, 0x65, 0x72, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x65, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49,
	0x64, 0x22, 0x23, 0x0a, 0x08, 0x50, 0x65, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x17, 0x0a,
	0x07, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x70, 0x65, 0x65, 0x72, 0x49, 0x64, 0x22, 0x4a, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x50, 0x65, 0x65,
	0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a,
	0x0b, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x65, 0x65, 0x72, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x65,
	0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0a, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x50, 0x65, 0x65,
	0x72, 0x73, 0x32, 0xa2, 0x01, 0x0a, 0x0b, 0x50, 0x65, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x45, 0x0a, 0x08, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x79, 0x12, 0x19,
	0x2e, 0x70, 0x65, 0x65, 0x72, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x66, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x70, 0x65, 0x65, 0x72,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x4c, 0x0a, 0x0b, 0x47, 0x65, 0x74,
	0x50, 0x65, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1c, 0x2e, 0x70, 0x65, 0x65, 0x72, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x65, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x70, 0x65, 0x65, 0x72, 0x2e, 0x61, 0x70,
	0x69, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x65, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescData = file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_github_com_aperturerobotics_bifrost_peer_api_api_proto_goTypes = []interface{}{
	(*IdentifyRequest)(nil),     // 0: peer.api.IdentifyRequest
	(*IdentifyResponse)(nil),    // 1: peer.api.IdentifyResponse
	(*GetPeerInfoRequest)(nil),  // 2: peer.api.GetPeerInfoRequest
	(*PeerInfo)(nil),            // 3: peer.api.PeerInfo
	(*GetPeerInfoResponse)(nil), // 4: peer.api.GetPeerInfoResponse
	(*controller.Config)(nil),   // 5: peer.controller.Config
	(exec.ControllerStatus)(0),  // 6: controller.exec.ControllerStatus
}
var file_github_com_aperturerobotics_bifrost_peer_api_api_proto_depIdxs = []int32{
	5, // 0: peer.api.IdentifyRequest.config:type_name -> peer.controller.Config
	6, // 1: peer.api.IdentifyResponse.controller_status:type_name -> controller.exec.ControllerStatus
	3, // 2: peer.api.GetPeerInfoResponse.local_peers:type_name -> peer.api.PeerInfo
	0, // 3: peer.api.PeerService.Identify:input_type -> peer.api.IdentifyRequest
	2, // 4: peer.api.PeerService.GetPeerInfo:input_type -> peer.api.GetPeerInfoRequest
	1, // 5: peer.api.PeerService.Identify:output_type -> peer.api.IdentifyResponse
	4, // 6: peer.api.PeerService.GetPeerInfo:output_type -> peer.api.GetPeerInfoResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_peer_api_api_proto_init() }
func file_github_com_aperturerobotics_bifrost_peer_api_api_proto_init() {
	if File_github_com_aperturerobotics_bifrost_peer_api_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IdentifyRequest); i {
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
		file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IdentifyResponse); i {
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
		file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPeerInfoRequest); i {
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
		file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerInfo); i {
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
		file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPeerInfoResponse); i {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_peer_api_api_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_peer_api_api_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_peer_api_api_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_peer_api_api_proto = out.File
	file_github_com_aperturerobotics_bifrost_peer_api_api_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_peer_api_api_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_peer_api_api_proto_depIdxs = nil
}
