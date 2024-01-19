// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0-devel
// 	protoc        v4.25.2
// source: github.com/aperturerobotics/bifrost/transport/websocket/websocket.proto

package websocket

import (
	reflect "reflect"
	sync "sync"

	dialer "github.com/aperturerobotics/bifrost/transport/common/dialer"
	quic "github.com/aperturerobotics/bifrost/transport/common/quic"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Config is the configuration for the Websocket transport.
//
// Bifrost speaks Quic over the websocket. While this is not always necessary,
// especially when using wss transports, we still need to ensure end-to-end
// encryption to the peer that we handshake with on the other end, and to manage
// stream congestion control, multiplexing,
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TransportPeerID sets the peer ID to attach the transport to.
	// If unset, attaches to any running peer with a private key.
	TransportPeerId string `protobuf:"bytes,1,opt,name=transport_peer_id,json=transportPeerId,proto3" json:"transport_peer_id,omitempty"`
	// ListenAddr contains the address to listen on.
	// Has no effect in the browser.
	ListenAddr string `protobuf:"bytes,2,opt,name=listen_addr,json=listenAddr,proto3" json:"listen_addr,omitempty"`
	// Quic contains the quic protocol options.
	//
	// The WebSocket transport always disables FEC and several other UDP-centric
	// features which are unnecessary due to the "reliable" nature of WebSockets.
	Quic *quic.Opts `protobuf:"bytes,3,opt,name=quic,proto3" json:"quic,omitempty"`
	// Dialers maps peer IDs to dialers.
	Dialers map[string]*dialer.DialerOpts `protobuf:"bytes,4,rep,name=dialers,proto3" json:"dialers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// RestrictPeerId restricts all incoming peer IDs to the given ID.
	// Any other peers trying to connect will be disconneted at handshake time.
	RestrictPeerId string `protobuf:"bytes,5,opt,name=restrict_peer_id,json=restrictPeerId,proto3" json:"restrict_peer_id,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_msgTypes[0]
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
	return file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetTransportPeerId() string {
	if x != nil {
		return x.TransportPeerId
	}
	return ""
}

func (x *Config) GetListenAddr() string {
	if x != nil {
		return x.ListenAddr
	}
	return ""
}

func (x *Config) GetQuic() *quic.Opts {
	if x != nil {
		return x.Quic
	}
	return nil
}

func (x *Config) GetDialers() map[string]*dialer.DialerOpts {
	if x != nil {
		return x.Dialers
	}
	return nil
}

func (x *Config) GetRestrictPeerId() string {
	if x != nil {
		return x.RestrictPeerId
	}
	return ""
}

var File_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDesc = []byte{
	0x0a, 0x47, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f,
	0x77, 0x65, 0x62, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2f, 0x77, 0x65, 0x62, 0x73, 0x6f, 0x63,
	0x6b, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x77, 0x65, 0x62, 0x73, 0x6f,
	0x63, 0x6b, 0x65, 0x74, 0x1a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x61, 0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63,
	0x73, 0x2f, 0x62, 0x69, 0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x71, 0x75, 0x69, 0x63, 0x2f,
	0x71, 0x75, 0x69, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x48, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72,
	0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69, 0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e,
	0x2f, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x2f, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0xb3, 0x02, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x2a, 0x0a, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x70, 0x65, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x74, 0x72, 0x61, 0x6e,
	0x73, 0x70, 0x6f, 0x72, 0x74, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x6c,
	0x69, 0x73, 0x74, 0x65, 0x6e, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x41, 0x64, 0x64, 0x72, 0x12, 0x28, 0x0a, 0x04,
	0x71, 0x75, 0x69, 0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2e, 0x71, 0x75, 0x69, 0x63, 0x2e, 0x4f, 0x70, 0x74, 0x73,
	0x52, 0x04, 0x71, 0x75, 0x69, 0x63, 0x12, 0x38, 0x0a, 0x07, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x72,
	0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x77, 0x65, 0x62, 0x73, 0x6f, 0x63,
	0x6b, 0x65, 0x74, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x44, 0x69, 0x61, 0x6c, 0x65,
	0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x73,
	0x12, 0x28, 0x0a, 0x10, 0x72, 0x65, 0x73, 0x74, 0x72, 0x69, 0x63, 0x74, 0x5f, 0x70, 0x65, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x72, 0x65, 0x73, 0x74,
	0x72, 0x69, 0x63, 0x74, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x1a, 0x4e, 0x0a, 0x0c, 0x44, 0x69,
	0x61, 0x6c, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x64, 0x69,
	0x61, 0x6c, 0x65, 0x72, 0x2e, 0x44, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x4f, 0x70, 0x74, 0x73, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDescData = file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_goTypes = []interface{}{
	(*Config)(nil),            // 0: websocket.Config
	nil,                       // 1: websocket.Config.DialersEntry
	(*quic.Opts)(nil),         // 2: transport.quic.Opts
	(*dialer.DialerOpts)(nil), // 3: dialer.DialerOpts
}
var file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_depIdxs = []int32{
	2, // 0: websocket.Config.quic:type_name -> transport.quic.Opts
	1, // 1: websocket.Config.dialers:type_name -> websocket.Config.DialersEntry
	3, // 2: websocket.Config.DialersEntry.value:type_name -> dialer.DialerOpts
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_init() }
func file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_init() {
	if File_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto = out.File
	file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_transport_websocket_websocket_proto_depIdxs = nil
}
