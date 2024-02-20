// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0-devel
// 	protoc        v4.25.3
// source: github.com/aperturerobotics/bifrost/link/establish/config.proto

package link_establish_controller

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

// Config is the link establish controller config.
// The establish controller attempts to establish links with configured peers.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// PeerIds is the list of peer IDs to attempt to establish links to.
	PeerIds []string `protobuf:"bytes,1,rep,name=peer_ids,json=peerIds,proto3" json:"peer_ids,omitempty"`
	// SrcPeerId is the source peer id to establish links from.
	// Can be empty.
	SrcPeerId string `protobuf:"bytes,2,opt,name=src_peer_id,json=srcPeerId,proto3" json:"src_peer_id,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_link_establish_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_link_establish_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetPeerIds() []string {
	if x != nil {
		return x.PeerIds
	}
	return nil
}

func (x *Config) GetSrcPeerId() string {
	if x != nil {
		return x.SrcPeerId
	}
	return ""
}

var File_github_com_aperturerobotics_bifrost_link_establish_config_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDesc = []byte{
	0x0a, 0x3f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x6c, 0x69, 0x6e, 0x6b, 0x2f, 0x65, 0x73, 0x74, 0x61, 0x62,
	0x6c, 0x69, 0x73, 0x68, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x19, 0x6c, 0x69, 0x6e, 0x6b, 0x2e, 0x65, 0x73, 0x74, 0x61, 0x62, 0x6c, 0x69, 0x73,
	0x68, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x22, 0x43, 0x0a, 0x06,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69,
	0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x70, 0x65, 0x65, 0x72, 0x49, 0x64,
	0x73, 0x12, 0x1e, 0x0a, 0x0b, 0x73, 0x72, 0x63, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x72, 0x63, 0x50, 0x65, 0x65, 0x72, 0x49,
	0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDescData = file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_link_establish_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_aperturerobotics_bifrost_link_establish_config_proto_goTypes = []interface{}{
	(*Config)(nil), // 0: link.establish.controller.Config
}
var file_github_com_aperturerobotics_bifrost_link_establish_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_link_establish_config_proto_init() }
func file_github_com_aperturerobotics_bifrost_link_establish_config_proto_init() {
	if File_github_com_aperturerobotics_bifrost_link_establish_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_link_establish_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_link_establish_config_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_link_establish_config_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_link_establish_config_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_link_establish_config_proto = out.File
	file_github_com_aperturerobotics_bifrost_link_establish_config_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_link_establish_config_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_link_establish_config_proto_depIdxs = nil
}
