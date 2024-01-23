// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0-devel
// 	protoc        v4.25.2
// source: github.com/aperturerobotics/bifrost/signaling/rpc/client/config.proto

package signaling_rpc_client

import (
	reflect "reflect"
	sync "sync"

	client "github.com/aperturerobotics/bifrost/stream/srpc/client"
	backoff "github.com/aperturerobotics/util/backoff"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Config configures a client for the Signaling SRPC service.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// SignalingId is the signaling channel ID.
	// Filters which SignalPeer directives will be handled.
	SignalingId string `protobuf:"bytes,1,opt,name=signaling_id,json=signalingId,proto3" json:"signaling_id,omitempty"`
	// PeerId is the local peer id to use for the client.
	// Can be empty to use any local peer.
	PeerId string `protobuf:"bytes,2,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	// Client contains srpc.client configuration for the signaling RPC client.
	// The local peer ID is overridden with the peer ID of the looked-up peer.
	Client *client.Config `protobuf:"bytes,3,opt,name=client,proto3" json:"client,omitempty"`
	// ProtocolId overrides the default protocol id for the signaling client.
	// Default: bifrost/signaling
	ProtocolId string `protobuf:"bytes,4,opt,name=protocol_id,json=protocolId,proto3" json:"protocol_id,omitempty"`
	// ServiceId overrides the default service id for the signaling client.
	// Default: signaling.rpc.Signaling
	ServiceId string `protobuf:"bytes,5,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	// Backoff is the backoff config for connecting to the service.
	// If unset, defaults to reasonable defaults.
	Backoff *backoff.Backoff `protobuf:"bytes,6,opt,name=backoff,proto3" json:"backoff,omitempty"`
	// DisableListen disables listening for incoming sessions.
	// If set, we will only call out, not accept incoming sessions.
	// If false, client will emit HandleSignalPeer directives for incoming sessions.
	DisableListen bool `protobuf:"varint,7,opt,name=disable_listen,json=disableListen,proto3" json:"disable_listen,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_msgTypes[0]
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
	return file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetSignalingId() string {
	if x != nil {
		return x.SignalingId
	}
	return ""
}

func (x *Config) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
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

func (x *Config) GetServiceId() string {
	if x != nil {
		return x.ServiceId
	}
	return ""
}

func (x *Config) GetBackoff() *backoff.Backoff {
	if x != nil {
		return x.Backoff
	}
	return nil
}

func (x *Config) GetDisableListen() bool {
	if x != nil {
		return x.DisableListen
	}
	return false
}

var File_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDesc = []byte{
	0x0a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x2f,
	0x72, 0x70, 0x63, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69,
	0x6e, 0x67, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x1a, 0x43, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72, 0x74, 0x75,
	0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69, 0x66, 0x72, 0x6f,
	0x73, 0x74, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2f, 0x73, 0x72, 0x70, 0x63, 0x2f, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61,
	0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f,
	0x75, 0x74, 0x69, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x2f, 0x62, 0x61, 0x63,
	0x6b, 0x6f, 0x66, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8b, 0x02, 0x0a, 0x06, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69,
	0x6e, 0x67, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x69, 0x67,
	0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x65, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x65, 0x65, 0x72, 0x49,
	0x64, 0x12, 0x32, 0x0a, 0x06, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x73, 0x72, 0x70, 0x63, 0x2e,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x63,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x2a, 0x0a, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66,
	0x2e, 0x42, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x52, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66,
	0x66, 0x12, 0x25, 0x0a, 0x0e, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x6c, 0x69, 0x73,
	0x74, 0x65, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x64, 0x69, 0x73, 0x61, 0x62,
	0x6c, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDescData = file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_goTypes = []interface{}{
	(*Config)(nil),          // 0: signaling.rpc.client.Config
	(*client.Config)(nil),   // 1: stream.srpc.client.Config
	(*backoff.Backoff)(nil), // 2: backoff.Backoff
}
var file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_depIdxs = []int32{
	1, // 0: signaling.rpc.client.Config.client:type_name -> stream.srpc.client.Config
	2, // 1: signaling.rpc.client.Config.backoff:type_name -> backoff.Backoff
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_init() }
func file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_init() {
	if File_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto = out.File
	file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_signaling_rpc_client_config_proto_depIdxs = nil
}
