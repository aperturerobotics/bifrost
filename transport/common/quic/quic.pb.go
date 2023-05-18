// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0-devel
// 	protoc        v3.21.9
// source: github.com/aperturerobotics/bifrost/transport/common/quic/quic.proto

package transport_quic

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

type Opts struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// MaxIdleTimeoutDur is the duration of idle after which conn is closed.
	//
	// If unset, uses a default value of 30 seconds.
	MaxIdleTimeoutDur string `protobuf:"bytes,1,opt,name=max_idle_timeout_dur,json=maxIdleTimeoutDur,proto3" json:"max_idle_timeout_dur,omitempty"`
	// MaxIncomingStreams is the maximum number of concurrent bidirectional
	// streams that a peer is allowed to open.
	//
	// If unset or negative, defaults to 100000.
	MaxIncomingStreams int32 `protobuf:"varint,2,opt,name=max_incoming_streams,json=maxIncomingStreams,proto3" json:"max_incoming_streams,omitempty"`
	// DisableKeepAlive disables the keep alive packets.
	DisableKeepAlive bool `protobuf:"varint,3,opt,name=disable_keep_alive,json=disableKeepAlive,proto3" json:"disable_keep_alive,omitempty"`
	// KeepAliveDur is the duration between keep-alive pings.
	//
	// If disable_keep_alive is set, this value is ignored.
	// If unset, sets keep-alive to half of MaxIdleTimeout.
	KeepAliveDur string `protobuf:"bytes,7,opt,name=keep_alive_dur,json=keepAliveDur,proto3" json:"keep_alive_dur,omitempty"`
	// DisableDatagrams disables the unreliable datagrams feature.
	// Both peers must support it for it to be enabled, regardless of this flag.
	DisableDatagrams bool `protobuf:"varint,4,opt,name=disable_datagrams,json=disableDatagrams,proto3" json:"disable_datagrams,omitempty"`
	// DisablePathMtuDiscovery disables sending packets to discover max packet size.
	DisablePathMtuDiscovery bool `protobuf:"varint,5,opt,name=disable_path_mtu_discovery,json=disablePathMtuDiscovery,proto3" json:"disable_path_mtu_discovery,omitempty"`
	// Verbose indicates to use verbose logging.
	// Note: this is VERY verbose, logs every packet sent.
	Verbose bool `protobuf:"varint,6,opt,name=verbose,proto3" json:"verbose,omitempty"`
}

func (x *Opts) Reset() {
	*x = Opts{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Opts) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Opts) ProtoMessage() {}

func (x *Opts) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Opts.ProtoReflect.Descriptor instead.
func (*Opts) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDescGZIP(), []int{0}
}

func (x *Opts) GetMaxIdleTimeoutDur() string {
	if x != nil {
		return x.MaxIdleTimeoutDur
	}
	return ""
}

func (x *Opts) GetMaxIncomingStreams() int32 {
	if x != nil {
		return x.MaxIncomingStreams
	}
	return 0
}

func (x *Opts) GetDisableKeepAlive() bool {
	if x != nil {
		return x.DisableKeepAlive
	}
	return false
}

func (x *Opts) GetKeepAliveDur() string {
	if x != nil {
		return x.KeepAliveDur
	}
	return ""
}

func (x *Opts) GetDisableDatagrams() bool {
	if x != nil {
		return x.DisableDatagrams
	}
	return false
}

func (x *Opts) GetDisablePathMtuDiscovery() bool {
	if x != nil {
		return x.DisablePathMtuDiscovery
	}
	return false
}

func (x *Opts) GetVerbose() bool {
	if x != nil {
		return x.Verbose
	}
	return false
}

var File_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDesc = []byte{
	0x0a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x71, 0x75, 0x69, 0x63, 0x2f, 0x71, 0x75, 0x69, 0x63,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72,
	0x74, 0x2e, 0x71, 0x75, 0x69, 0x63, 0x22, 0xc1, 0x02, 0x0a, 0x04, 0x4f, 0x70, 0x74, 0x73, 0x12,
	0x2f, 0x0a, 0x14, 0x6d, 0x61, 0x78, 0x5f, 0x69, 0x64, 0x6c, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x6f, 0x75, 0x74, 0x5f, 0x64, 0x75, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x6d,
	0x61, 0x78, 0x49, 0x64, 0x6c, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x44, 0x75, 0x72,
	0x12, 0x30, 0x0a, 0x14, 0x6d, 0x61, 0x78, 0x5f, 0x69, 0x6e, 0x63, 0x6f, 0x6d, 0x69, 0x6e, 0x67,
	0x5f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x12,
	0x6d, 0x61, 0x78, 0x49, 0x6e, 0x63, 0x6f, 0x6d, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x73, 0x12, 0x2c, 0x0a, 0x12, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x6b, 0x65,
	0x65, 0x70, 0x5f, 0x61, 0x6c, 0x69, 0x76, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x10,
	0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x4b, 0x65, 0x65, 0x70, 0x41, 0x6c, 0x69, 0x76, 0x65,
	0x12, 0x24, 0x0a, 0x0e, 0x6b, 0x65, 0x65, 0x70, 0x5f, 0x61, 0x6c, 0x69, 0x76, 0x65, 0x5f, 0x64,
	0x75, 0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6b, 0x65, 0x65, 0x70, 0x41, 0x6c,
	0x69, 0x76, 0x65, 0x44, 0x75, 0x72, 0x12, 0x2b, 0x0a, 0x11, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c,
	0x65, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x67, 0x72, 0x61, 0x6d, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x10, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x44, 0x61, 0x74, 0x61, 0x67, 0x72,
	0x61, 0x6d, 0x73, 0x12, 0x3b, 0x0a, 0x1a, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x70,
	0x61, 0x74, 0x68, 0x5f, 0x6d, 0x74, 0x75, 0x5f, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72,
	0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x17, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65,
	0x50, 0x61, 0x74, 0x68, 0x4d, 0x74, 0x75, 0x44, 0x69, 0x73, 0x63, 0x6f, 0x76, 0x65, 0x72, 0x79,
	0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x62, 0x6f, 0x73, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x07, 0x76, 0x65, 0x72, 0x62, 0x6f, 0x73, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDescData = file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_goTypes = []interface{}{
	(*Opts)(nil), // 0: transport.quic.Opts
}
var file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_init() }
func file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_init() {
	if File_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Opts); i {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto = out.File
	file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_transport_common_quic_quic_proto_depIdxs = nil
}
