// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0-devel
// 	protoc        v3.19.3
// source: github.com/aperturerobotics/bifrost/stream/srpc/client/client.proto

package stream_srpc_client

import (
	reflect "reflect"
	sync "sync"

	backoff "github.com/aperturerobotics/bifrost/util/backoff"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Config configures a client for a srpc service.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ServerPeerIds are the static list of peer IDs to contact.
	ServerPeerIds []string `protobuf:"bytes,1,rep,name=server_peer_ids,json=serverPeerIds,proto3" json:"server_peer_ids,omitempty"`
	// PerServerBackoff is the server peer error backoff configuration.
	// Can be empty.
	PerServerBackoff *backoff.Backoff `protobuf:"bytes,2,opt,name=per_server_backoff,json=perServerBackoff,proto3" json:"per_server_backoff,omitempty"`
	// SrcPeerId is the source peer id to contact from.
	// Can be empty.
	SrcPeerId string `protobuf:"bytes,3,opt,name=src_peer_id,json=srcPeerId,proto3" json:"src_peer_id,omitempty"`
	// TransportId restricts which transport we can dial out from.
	TransportId uint64 `protobuf:"varint,4,opt,name=transport_id,json=transportId,proto3" json:"transport_id,omitempty"`
	// TimeoutDur sets the per-server establish timeout.
	// If unset, no timeout.
	// Example: 15s
	TimeoutDur string `protobuf:"bytes,5,opt,name=timeout_dur,json=timeoutDur,proto3" json:"timeout_dur,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_msgTypes[0]
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
	return file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetServerPeerIds() []string {
	if x != nil {
		return x.ServerPeerIds
	}
	return nil
}

func (x *Config) GetPerServerBackoff() *backoff.Backoff {
	if x != nil {
		return x.PerServerBackoff
	}
	return nil
}

func (x *Config) GetSrcPeerId() string {
	if x != nil {
		return x.SrcPeerId
	}
	return ""
}

func (x *Config) GetTransportId() uint64 {
	if x != nil {
		return x.TransportId
	}
	return 0
}

func (x *Config) GetTimeoutDur() string {
	if x != nil {
		return x.TimeoutDur
	}
	return ""
}

var File_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDesc = []byte{
	0x0a, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2f, 0x73, 0x72, 0x70,
	0x63, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x12, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x73, 0x72,
	0x70, 0x63, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x1a, 0x3e, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f,
	0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69, 0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x75,
	0x74, 0x69, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x2f, 0x62, 0x61, 0x63, 0x6b,
	0x6f, 0x66, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd4, 0x01, 0x0a, 0x06, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x12, 0x26, 0x0a, 0x0f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x70,
	0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0d, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x73, 0x12, 0x3e, 0x0a, 0x12,
	0x70, 0x65, 0x72, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x6f,
	0x66, 0x66, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x62, 0x61, 0x63, 0x6b, 0x6f,
	0x66, 0x66, 0x2e, 0x42, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x52, 0x10, 0x70, 0x65, 0x72, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x42, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x12, 0x1e, 0x0a, 0x0b,
	0x73, 0x72, 0x63, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x73, 0x72, 0x63, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x0b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x49, 0x64, 0x12,
	0x1f, 0x0a, 0x0b, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x5f, 0x64, 0x75, 0x72, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x44, 0x75, 0x72,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDescData = file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_goTypes = []interface{}{
	(*Config)(nil),          // 0: stream.srpc.client.Config
	(*backoff.Backoff)(nil), // 1: backoff.Backoff
}
var file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_depIdxs = []int32{
	1, // 0: stream.srpc.client.Config.per_server_backoff:type_name -> backoff.Backoff
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_init() }
func file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_init() {
	if File_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto = out.File
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_stream_srpc_client_client_proto_depIdxs = nil
}