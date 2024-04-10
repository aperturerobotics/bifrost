// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0-devel
// 	protoc        v5.26.1
// source: github.com/aperturerobotics/bifrost/stream/srpc/client/controller/config.proto

package stream_srpc_client_controller

import (
	reflect "reflect"
	sync "sync"

	client "github.com/aperturerobotics/bifrost/stream/srpc/client"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Config configures mounting a bifrost srpc RPC client to a bus.
// Resolves the LookupRpcClient directive.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Client contains srpc.client configuration for the RPC client.
	Client *client.Config `protobuf:"bytes,1,opt,name=client,proto3" json:"client,omitempty"`
	// ProtocolId is the protocol ID to use to contact the remote RPC service.
	// Must be set.
	ProtocolId string `protobuf:"bytes,2,opt,name=protocol_id,json=protocolId,proto3" json:"protocol_id,omitempty"`
	// ServiceIdPrefixes are the service ID prefixes to match.
	// The prefix will be stripped from the service id before being passed to the client.
	// This is used like: LookupRpcClient<remote/my/service> -> my/service.
	//
	// If empty slice or empty string: matches all LookupRpcClient calls ignoring service ID.
	// Optional.
	ServiceIdPrefixes []string `protobuf:"bytes,3,rep,name=service_id_prefixes,json=serviceIdPrefixes,proto3" json:"service_id_prefixes,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_msgTypes[0]
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
	return file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetClient() *client.Config {
	if x != nil {
		return x.Client
	}
	return nil
}

func (x *Config) GetProtocolId() string {
	if x != nil {
		return x.ProtocolId
	}
	return ""
}

func (x *Config) GetServiceIdPrefixes() []string {
	if x != nil {
		return x.ServiceIdPrefixes
	}
	return nil
}

var File_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDesc = []byte{
	0x0a, 0x4e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2f, 0x73, 0x72, 0x70,
	0x63, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x1d, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x1a,
	0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72,
	0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69, 0x66,
	0x72, 0x6f, 0x73, 0x74, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2f, 0x73, 0x72, 0x70, 0x63,
	0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8d, 0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x32, 0x0a, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x49, 0x64, 0x12, 0x2e, 0x0a, 0x13, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f,
	0x69, 0x64, 0x5f, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x11, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x50, 0x72, 0x65, 0x66,
	0x69, 0x78, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDescData = file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_goTypes = []interface{}{
	(*Config)(nil),        // 0: stream.srpc.client.controller.Config
	(*client.Config)(nil), // 1: stream.srpc.client.Config
}
var file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_depIdxs = []int32{
	1, // 0: stream.srpc.client.controller.Config.client:type_name -> stream.srpc.client.Config
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() {
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_init()
}
func file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_init() {
	if File_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto = out.File
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_controller_config_proto_depIdxs = nil
}
