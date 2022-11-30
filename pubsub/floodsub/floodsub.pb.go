// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1-devel
// 	protoc        v3.21.9
// source: github.com/aperturerobotics/bifrost/pubsub/floodsub/floodsub.proto

package floodsub

import (
	reflect "reflect"
	sync "sync"

	hash "github.com/aperturerobotics/bifrost/hash"
	peer "github.com/aperturerobotics/bifrost/peer"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Config configures the floodsub router.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// PublishHashType is the hash type to use when signing published messages.
	// Defaults to sha256
	PublishHashType hash.HashType `protobuf:"varint,1,opt,name=publish_hash_type,json=publishHashType,proto3,enum=hash.HashType" json:"publish_hash_type,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[0]
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
	return file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetPublishHashType() hash.HashType {
	if x != nil {
		return x.PublishHashType
	}
	return hash.HashType(0)
}

// Packet is the floodsub packet.
type Packet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Subscriptions contains any new subscription changes.
	Subscriptions []*SubscriptionOpts `protobuf:"bytes,1,rep,name=subscriptions,proto3" json:"subscriptions,omitempty"`
	// Publish contains messages we are publishing.
	Publish []*peer.SignedMsg `protobuf:"bytes,2,rep,name=publish,proto3" json:"publish,omitempty"`
}

func (x *Packet) Reset() {
	*x = Packet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Packet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet) ProtoMessage() {}

func (x *Packet) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet.ProtoReflect.Descriptor instead.
func (*Packet) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescGZIP(), []int{1}
}

func (x *Packet) GetSubscriptions() []*SubscriptionOpts {
	if x != nil {
		return x.Subscriptions
	}
	return nil
}

func (x *Packet) GetPublish() []*peer.SignedMsg {
	if x != nil {
		return x.Publish
	}
	return nil
}

// SubscriptionOpts are subscription options.
type SubscriptionOpts struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Subscribe indicates if we are subscribing to this channel ID.
	Subscribe bool `protobuf:"varint,1,opt,name=subscribe,proto3" json:"subscribe,omitempty"`
	// ChannelId is the channel to subscribe to.
	ChannelId string `protobuf:"bytes,2,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
}

func (x *SubscriptionOpts) Reset() {
	*x = SubscriptionOpts{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscriptionOpts) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscriptionOpts) ProtoMessage() {}

func (x *SubscriptionOpts) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscriptionOpts.ProtoReflect.Descriptor instead.
func (*SubscriptionOpts) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescGZIP(), []int{2}
}

func (x *SubscriptionOpts) GetSubscribe() bool {
	if x != nil {
		return x.Subscribe
	}
	return false
}

func (x *SubscriptionOpts) GetChannelId() string {
	if x != nil {
		return x.ChannelId
	}
	return ""
}

var File_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDesc = []byte{
	0x0a, 0x42, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2f, 0x66, 0x6c, 0x6f,
	0x6f, 0x64, 0x73, 0x75, 0x62, 0x2f, 0x66, 0x6c, 0x6f, 0x6f, 0x64, 0x73, 0x75, 0x62, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x66, 0x6c, 0x6f, 0x6f, 0x64, 0x73, 0x75, 0x62, 0x1a, 0x33,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72, 0x74,
	0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69, 0x66, 0x72,
	0x6f, 0x73, 0x74, 0x2f, 0x70, 0x65, 0x65, 0x72, 0x2f, 0x70, 0x65, 0x65, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x61, 0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73,
	0x2f, 0x62, 0x69, 0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x68, 0x61, 0x73, 0x68, 0x2f, 0x68, 0x61,
	0x73, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x44, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x3a, 0x0a, 0x11, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x5f, 0x68, 0x61,
	0x73, 0x68, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e,
	0x68, 0x61, 0x73, 0x68, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0f, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x73, 0x68, 0x48, 0x61, 0x73, 0x68, 0x54, 0x79, 0x70, 0x65, 0x22, 0x75,
	0x0a, 0x06, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x12, 0x40, 0x0a, 0x0d, 0x73, 0x75, 0x62, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x66, 0x6c, 0x6f, 0x6f, 0x64, 0x73, 0x75, 0x62, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x70, 0x74, 0x73, 0x52, 0x0d, 0x73, 0x75, 0x62,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x29, 0x0a, 0x07, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x73, 0x68, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x70, 0x65,
	0x65, 0x72, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x65, 0x64, 0x4d, 0x73, 0x67, 0x52, 0x07, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x73, 0x68, 0x22, 0x4f, 0x0a, 0x10, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x4f, 0x70, 0x74, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x73, 0x75, 0x62,
	0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x73, 0x75,
	0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x68, 0x61, 0x6e, 0x6e,
	0x65, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x68, 0x61,
	0x6e, 0x6e, 0x65, 0x6c, 0x49, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescData = file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_goTypes = []interface{}{
	(*Config)(nil),           // 0: floodsub.Config
	(*Packet)(nil),           // 1: floodsub.Packet
	(*SubscriptionOpts)(nil), // 2: floodsub.SubscriptionOpts
	(hash.HashType)(0),       // 3: hash.HashType
	(*peer.SignedMsg)(nil),   // 4: peer.SignedMsg
}
var file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_depIdxs = []int32{
	3, // 0: floodsub.Config.publish_hash_type:type_name -> hash.HashType
	2, // 1: floodsub.Packet.subscriptions:type_name -> floodsub.SubscriptionOpts
	4, // 2: floodsub.Packet.publish:type_name -> peer.SignedMsg
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_init() }
func file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_init() {
	if File_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Packet); i {
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
		file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscriptionOpts); i {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto = out.File
	file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_pubsub_floodsub_floodsub_proto_depIdxs = nil
}
