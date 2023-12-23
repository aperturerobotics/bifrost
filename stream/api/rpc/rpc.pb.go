// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0-devel
// 	protoc        v4.25.1
// source: github.com/aperturerobotics/bifrost/stream/api/rpc/rpc.proto

package stream_api_rpc

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

// StreamState is state for the stream related calls.
type StreamState int32

const (
	// StreamState_NONE indicates nothing about the state
	StreamState_StreamState_NONE StreamState = 0
	// StreamState_ESTABLISHING indicates the stream is connecting.
	StreamState_StreamState_ESTABLISHING StreamState = 1
	// StreamState_ESTABLISHED indicates the stream is established.
	StreamState_StreamState_ESTABLISHED StreamState = 2
)

// Enum value maps for StreamState.
var (
	StreamState_name = map[int32]string{
		0: "StreamState_NONE",
		1: "StreamState_ESTABLISHING",
		2: "StreamState_ESTABLISHED",
	}
	StreamState_value = map[string]int32{
		"StreamState_NONE":         0,
		"StreamState_ESTABLISHING": 1,
		"StreamState_ESTABLISHED":  2,
	}
)

func (x StreamState) Enum() *StreamState {
	p := new(StreamState)
	*p = x
	return p
}

func (x StreamState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (StreamState) Descriptor() protoreflect.EnumDescriptor {
	return file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_enumTypes[0].Descriptor()
}

func (StreamState) Type() protoreflect.EnumType {
	return &file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_enumTypes[0]
}

func (x StreamState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use StreamState.Descriptor instead.
func (StreamState) EnumDescriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescGZIP(), []int{0}
}

// Data is a data packet.
type Data struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// State indicates stream state in-band.
	// Data is packet data from the remote.
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	// State indicates the stream state.
	State StreamState `protobuf:"varint,2,opt,name=state,proto3,enum=stream.api.rpc.StreamState" json:"state,omitempty"`
}

func (x *Data) Reset() {
	*x = Data{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Data) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Data) ProtoMessage() {}

func (x *Data) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Data.ProtoReflect.Descriptor instead.
func (*Data) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescGZIP(), []int{0}
}

func (x *Data) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *Data) GetState() StreamState {
	if x != nil {
		return x.State
	}
	return StreamState_StreamState_NONE
}

var File_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDesc = []byte{
	0x0a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x72, 0x70, 0x63, 0x2f, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x72, 0x70, 0x63, 0x22, 0x4d,
	0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x31, 0x0a, 0x05, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x73, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2a, 0x5e, 0x0a,
	0x0b, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x14, 0x0a, 0x10,
	0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x5f, 0x4e, 0x4f, 0x4e, 0x45,
	0x10, 0x00, 0x12, 0x1c, 0x0a, 0x18, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x5f, 0x45, 0x53, 0x54, 0x41, 0x42, 0x4c, 0x49, 0x53, 0x48, 0x49, 0x4e, 0x47, 0x10, 0x01,
	0x12, 0x1b, 0x0a, 0x17, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x5f,
	0x45, 0x53, 0x54, 0x41, 0x42, 0x4c, 0x49, 0x53, 0x48, 0x45, 0x44, 0x10, 0x02, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescData = file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_goTypes = []interface{}{
	(StreamState)(0), // 0: stream.api.rpc.StreamState
	(*Data)(nil),     // 1: stream.api.rpc.Data
}
var file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_depIdxs = []int32{
	0, // 0: stream.api.rpc.Data.state:type_name -> stream.api.rpc.StreamState
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_init() }
func file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_init() {
	if File_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Data); i {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_depIdxs,
		EnumInfos:         file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_enumTypes,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto = out.File
	file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_stream_api_rpc_rpc_proto_depIdxs = nil
}
