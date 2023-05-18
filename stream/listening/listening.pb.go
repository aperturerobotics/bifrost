// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0-devel
// 	protoc        v3.21.9
// source: github.com/aperturerobotics/bifrost/stream/listening/listening.proto

package stream_listening

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

// Config configures the listening controller.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// LocalPeerId is the peer ID to forward incoming connections with.
	// Can be empty.
	LocalPeerId string `protobuf:"bytes,1,opt,name=local_peer_id,json=localPeerId,proto3" json:"local_peer_id,omitempty"`
	// RemotePeerId is the peer ID to forward incoming connections to.
	RemotePeerId string `protobuf:"bytes,2,opt,name=remote_peer_id,json=remotePeerId,proto3" json:"remote_peer_id,omitempty"`
	// ProtocolId is the protocol ID to assign to incoming connections.
	ProtocolId string `protobuf:"bytes,3,opt,name=protocol_id,json=protocolId,proto3" json:"protocol_id,omitempty"`
	// ListenMultiaddr is the listening multiaddress.
	ListenMultiaddr string `protobuf:"bytes,4,opt,name=listen_multiaddr,json=listenMultiaddr,proto3" json:"listen_multiaddr,omitempty"`
	// TransportId sets a transport ID constraint.
	// Can be empty.
	TransportId uint64 `protobuf:"varint,5,opt,name=transport_id,json=transportId,proto3" json:"transport_id,omitempty"`
	// Reliable indicates the stream should be reliable.
	Reliable bool `protobuf:"varint,6,opt,name=reliable,proto3" json:"reliable,omitempty"`
	// Encrypted indicates the stream should be encrypted.
	Encrypted bool `protobuf:"varint,7,opt,name=encrypted,proto3" json:"encrypted,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_msgTypes[0]
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
	return file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetLocalPeerId() string {
	if x != nil {
		return x.LocalPeerId
	}
	return ""
}

func (x *Config) GetRemotePeerId() string {
	if x != nil {
		return x.RemotePeerId
	}
	return ""
}

func (x *Config) GetProtocolId() string {
	if x != nil {
		return x.ProtocolId
	}
	return ""
}

func (x *Config) GetListenMultiaddr() string {
	if x != nil {
		return x.ListenMultiaddr
	}
	return ""
}

func (x *Config) GetTransportId() uint64 {
	if x != nil {
		return x.TransportId
	}
	return 0
}

func (x *Config) GetReliable() bool {
	if x != nil {
		return x.Reliable
	}
	return false
}

func (x *Config) GetEncrypted() bool {
	if x != nil {
		return x.Encrypted
	}
	return false
}

var File_github_com_aperturerobotics_bifrost_stream_listening_listening_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDesc = []byte{
	0x0a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2f, 0x6c, 0x69, 0x73,
	0x74, 0x65, 0x6e, 0x69, 0x6e, 0x67, 0x2f, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x69, 0x6e, 0x67,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x6c,
	0x69, 0x73, 0x74, 0x65, 0x6e, 0x69, 0x6e, 0x67, 0x22, 0xfb, 0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0x22, 0x0a, 0x0d, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x5f, 0x70, 0x65, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6c, 0x6f, 0x63, 0x61,
	0x6c, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x24, 0x0a, 0x0e, 0x72, 0x65, 0x6d, 0x6f, 0x74,
	0x65, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0c, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1f, 0x0a,
	0x0b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x49, 0x64, 0x12, 0x29,
	0x0a, 0x10, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x5f, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x61, 0x64,
	0x64, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e,
	0x4d, 0x75, 0x6c, 0x74, 0x69, 0x61, 0x64, 0x64, 0x72, 0x12, 0x21, 0x0a, 0x0c, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x0b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08,
	0x72, 0x65, 0x6c, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08,
	0x72, 0x65, 0x6c, 0x69, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x6e, 0x63, 0x72,
	0x79, 0x70, 0x74, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x65, 0x6e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDescData = file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_goTypes = []interface{}{
	(*Config)(nil), // 0: stream.listening.Config
}
var file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_init() }
func file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_init() {
	if File_github_com_aperturerobotics_bifrost_stream_listening_listening_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_stream_listening_listening_proto = out.File
	file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_stream_listening_listening_proto_depIdxs = nil
}
