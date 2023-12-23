// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0-devel
// 	protoc        v4.25.1
// source: github.com/aperturerobotics/bifrost/transport/common/kcp/kcp_handshake.proto

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

// HandshakeExtraData contains the extra data field of the pconn handshake.
type HandshakeExtraData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// LocalTransportUuid is the transport uuid of the sender.
	// This is used for monitoring / analysis at a later time.
	// Coorelates the transport connections between two machines.
	LocalTransportUuid uint64 `protobuf:"varint,1,opt,name=local_transport_uuid,json=localTransportUuid,proto3" json:"local_transport_uuid,omitempty"`
}

func (x *HandshakeExtraData) Reset() {
	*x = HandshakeExtraData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HandshakeExtraData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HandshakeExtraData) ProtoMessage() {}

func (x *HandshakeExtraData) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HandshakeExtraData.ProtoReflect.Descriptor instead.
func (*HandshakeExtraData) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDescGZIP(), []int{0}
}

func (x *HandshakeExtraData) GetLocalTransportUuid() uint64 {
	if x != nil {
		return x.LocalTransportUuid
	}
	return 0
}

var File_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDesc = []byte{
	0x0a, 0x4c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x6b, 0x63, 0x70, 0x2f, 0x6b, 0x63, 0x70, 0x5f, 0x68,
	0x61, 0x6e, 0x64, 0x73, 0x68, 0x61, 0x6b, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03,
	0x6b, 0x63, 0x70, 0x22, 0x46, 0x0a, 0x12, 0x48, 0x61, 0x6e, 0x64, 0x73, 0x68, 0x61, 0x6b, 0x65,
	0x45, 0x78, 0x74, 0x72, 0x61, 0x44, 0x61, 0x74, 0x61, 0x12, 0x30, 0x0a, 0x14, 0x6c, 0x6f, 0x63,
	0x61, 0x6c, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x75, 0x75, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x12, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x55, 0x75, 0x69, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDescData = file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_goTypes = []interface{}{
	(*HandshakeExtraData)(nil), // 0: kcp.HandshakeExtraData
}
var file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_init() }
func file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_init() {
	if File_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HandshakeExtraData); i {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto = out.File
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_transport_common_kcp_kcp_handshake_proto_depIdxs = nil
}
