// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1-devel
// 	protoc        v3.21.9
// source: github.com/aperturerobotics/bifrost/pubsub/api/api.proto

package pubsub_api

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

// SubcribeRequest is a pubsub subscription request message.
type SubscribeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// ChannelId is the channel id to subscribe to.
	// Must be sent before / with publish.
	// Cannot change the channel ID after first transmission.
	ChannelId string `protobuf:"bytes,1,opt,name=channel_id,json=channelId,proto3" json:"channel_id,omitempty"`
	// PeerId is the peer identifier of the publisher/subscriber.
	// The peer ID will be used to acquire the peer private key.
	PeerId string `protobuf:"bytes,2,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	// PrivKeyPem is an alternate to PeerId, specify private key inline.
	// Overrides PeerId if set.
	PrivKeyPem string `protobuf:"bytes,3,opt,name=priv_key_pem,json=privKeyPem,proto3" json:"priv_key_pem,omitempty"`
	// PublishRequest contains a publish message request.
	PublishRequest *PublishRequest `protobuf:"bytes,4,opt,name=publish_request,json=publishRequest,proto3" json:"publish_request,omitempty"`
}

func (x *SubscribeRequest) Reset() {
	*x = SubscribeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeRequest) ProtoMessage() {}

func (x *SubscribeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeRequest.ProtoReflect.Descriptor instead.
func (*SubscribeRequest) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescGZIP(), []int{0}
}

func (x *SubscribeRequest) GetChannelId() string {
	if x != nil {
		return x.ChannelId
	}
	return ""
}

func (x *SubscribeRequest) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

func (x *SubscribeRequest) GetPrivKeyPem() string {
	if x != nil {
		return x.PrivKeyPem
	}
	return ""
}

func (x *SubscribeRequest) GetPublishRequest() *PublishRequest {
	if x != nil {
		return x.PublishRequest
	}
	return nil
}

// PublishRequest is a message published via the subscribe channel.
type PublishRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Data is the published data.
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	// Identifier is a uint32 identifier to use for outgoing status.
	// If zero, no outgoing status response will be sent.
	Identifier uint32 `protobuf:"varint,2,opt,name=identifier,proto3" json:"identifier,omitempty"`
}

func (x *PublishRequest) Reset() {
	*x = PublishRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PublishRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PublishRequest) ProtoMessage() {}

func (x *PublishRequest) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PublishRequest.ProtoReflect.Descriptor instead.
func (*PublishRequest) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescGZIP(), []int{1}
}

func (x *PublishRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *PublishRequest) GetIdentifier() uint32 {
	if x != nil {
		return x.Identifier
	}
	return 0
}

// SubcribeResponse is a pubsub subscription response message.
type SubscribeResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// IncomingMessage is an incoming message.
	IncomingMessage *IncomingMessage `protobuf:"bytes,1,opt,name=incoming_message,json=incomingMessage,proto3" json:"incoming_message,omitempty"`
	// OutgoingStatus is status of an outgoing message.
	// Sent when a Publish request finishes.
	OutgoingStatus *OutgoingStatus `protobuf:"bytes,2,opt,name=outgoing_status,json=outgoingStatus,proto3" json:"outgoing_status,omitempty"`
	// SubscriptionStatus is the status of the subscription
	SubscriptionStatus *SubscriptionStatus `protobuf:"bytes,3,opt,name=subscription_status,json=subscriptionStatus,proto3" json:"subscription_status,omitempty"`
}

func (x *SubscribeResponse) Reset() {
	*x = SubscribeResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscribeResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscribeResponse) ProtoMessage() {}

func (x *SubscribeResponse) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscribeResponse.ProtoReflect.Descriptor instead.
func (*SubscribeResponse) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescGZIP(), []int{2}
}

func (x *SubscribeResponse) GetIncomingMessage() *IncomingMessage {
	if x != nil {
		return x.IncomingMessage
	}
	return nil
}

func (x *SubscribeResponse) GetOutgoingStatus() *OutgoingStatus {
	if x != nil {
		return x.OutgoingStatus
	}
	return nil
}

func (x *SubscribeResponse) GetSubscriptionStatus() *SubscriptionStatus {
	if x != nil {
		return x.SubscriptionStatus
	}
	return nil
}

// SubscripionStatus is the status of the subscription handle.
type SubscriptionStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Subscribed indicates the subscription is established.
	Subscribed bool `protobuf:"varint,1,opt,name=subscribed,proto3" json:"subscribed,omitempty"`
}

func (x *SubscriptionStatus) Reset() {
	*x = SubscriptionStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SubscriptionStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubscriptionStatus) ProtoMessage() {}

func (x *SubscriptionStatus) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubscriptionStatus.ProtoReflect.Descriptor instead.
func (*SubscriptionStatus) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescGZIP(), []int{3}
}

func (x *SubscriptionStatus) GetSubscribed() bool {
	if x != nil {
		return x.Subscribed
	}
	return false
}

// OutgoingStatus is status of an outgoing message.
type OutgoingStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Identifier is the request-provided identifier for the message.
	Identifier uint32 `protobuf:"varint,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	// Sent indicates if the message was sent.
	Sent bool `protobuf:"varint,2,opt,name=sent,proto3" json:"sent,omitempty"`
}

func (x *OutgoingStatus) Reset() {
	*x = OutgoingStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OutgoingStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OutgoingStatus) ProtoMessage() {}

func (x *OutgoingStatus) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OutgoingStatus.ProtoReflect.Descriptor instead.
func (*OutgoingStatus) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescGZIP(), []int{4}
}

func (x *OutgoingStatus) GetIdentifier() uint32 {
	if x != nil {
		return x.Identifier
	}
	return 0
}

func (x *OutgoingStatus) GetSent() bool {
	if x != nil {
		return x.Sent
	}
	return false
}

// IncomingMessage implements Message with a proto object.
type IncomingMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// FromPeerId is the peer identifier of the sender.
	FromPeerId string `protobuf:"bytes,1,opt,name=from_peer_id,json=fromPeerId,proto3" json:"from_peer_id,omitempty"`
	// Authenticated indicates if the message is verified to be from the sender.
	Authenticated bool `protobuf:"varint,2,opt,name=authenticated,proto3" json:"authenticated,omitempty"`
	// Data is the inner data.
	Data []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *IncomingMessage) Reset() {
	*x = IncomingMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IncomingMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IncomingMessage) ProtoMessage() {}

func (x *IncomingMessage) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IncomingMessage.ProtoReflect.Descriptor instead.
func (*IncomingMessage) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescGZIP(), []int{5}
}

func (x *IncomingMessage) GetFromPeerId() string {
	if x != nil {
		return x.FromPeerId
	}
	return ""
}

func (x *IncomingMessage) GetAuthenticated() bool {
	if x != nil {
		return x.Authenticated
	}
	return false
}

func (x *IncomingMessage) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_github_com_aperturerobotics_bifrost_pubsub_api_api_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDesc = []byte{
	0x0a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x70, 0x75, 0x62, 0x73,
	0x75, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x22, 0xb1, 0x01, 0x0a, 0x10, 0x53, 0x75, 0x62, 0x73, 0x63,
	0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x63,
	0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x65,
	0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x65, 0x65,
	0x72, 0x49, 0x64, 0x12, 0x20, 0x0a, 0x0c, 0x70, 0x72, 0x69, 0x76, 0x5f, 0x6b, 0x65, 0x79, 0x5f,
	0x70, 0x65, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x4b,
	0x65, 0x79, 0x50, 0x65, 0x6d, 0x12, 0x43, 0x0a, 0x0f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x73, 0x68,
	0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x50, 0x75, 0x62, 0x6c,
	0x69, 0x73, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x52, 0x0e, 0x70, 0x75, 0x62, 0x6c,
	0x69, 0x73, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x44, 0x0a, 0x0e, 0x50, 0x75,
	0x62, 0x6c, 0x69, 0x73, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x12, 0x1e, 0x0a, 0x0a, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72,
	0x22, 0xf1, 0x01, 0x0a, 0x11, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x46, 0x0a, 0x10, 0x69, 0x6e, 0x63, 0x6f, 0x6d, 0x69,
	0x6e, 0x67, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1b, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x49, 0x6e,
	0x63, 0x6f, 0x6d, 0x69, 0x6e, 0x67, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x0f, 0x69,
	0x6e, 0x63, 0x6f, 0x6d, 0x69, 0x6e, 0x67, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x43,
	0x0a, 0x0f, 0x6f, 0x75, 0x74, 0x67, 0x6f, 0x69, 0x6e, 0x67, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4f, 0x75, 0x74, 0x67, 0x6f, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x0e, 0x6f, 0x75, 0x74, 0x67, 0x6f, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x4f, 0x0a, 0x13, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1e, 0x2e, 0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x53, 0x75,
	0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x52, 0x12, 0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x22, 0x34, 0x0a, 0x12, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x75,
	0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a,
	0x73, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x64, 0x22, 0x44, 0x0a, 0x0e, 0x4f, 0x75,
	0x74, 0x67, 0x6f, 0x69, 0x6e, 0x67, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1e, 0x0a, 0x0a,
	0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x0a, 0x69, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04,
	0x73, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x73, 0x65, 0x6e, 0x74,
	0x22, 0x6d, 0x0a, 0x0f, 0x49, 0x6e, 0x63, 0x6f, 0x6d, 0x69, 0x6e, 0x67, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x20, 0x0a, 0x0c, 0x66, 0x72, 0x6f, 0x6d, 0x5f, 0x70, 0x65, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x66, 0x72, 0x6f, 0x6d, 0x50,
	0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x24, 0x0a, 0x0d, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74,
	0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x61, 0x75,
	0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x65, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32,
	0x5f, 0x0a, 0x0d, 0x50, 0x75, 0x62, 0x53, 0x75, 0x62, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x4e, 0x0a, 0x09, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x12, 0x1c, 0x2e,
	0x70, 0x75, 0x62, 0x73, 0x75, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63,
	0x72, 0x69, 0x62, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x70, 0x75,
	0x62, 0x73, 0x75, 0x62, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69,
	0x62, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescData = file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_goTypes = []interface{}{
	(*SubscribeRequest)(nil),   // 0: pubsub.api.SubscribeRequest
	(*PublishRequest)(nil),     // 1: pubsub.api.PublishRequest
	(*SubscribeResponse)(nil),  // 2: pubsub.api.SubscribeResponse
	(*SubscriptionStatus)(nil), // 3: pubsub.api.SubscriptionStatus
	(*OutgoingStatus)(nil),     // 4: pubsub.api.OutgoingStatus
	(*IncomingMessage)(nil),    // 5: pubsub.api.IncomingMessage
}
var file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_depIdxs = []int32{
	1, // 0: pubsub.api.SubscribeRequest.publish_request:type_name -> pubsub.api.PublishRequest
	5, // 1: pubsub.api.SubscribeResponse.incoming_message:type_name -> pubsub.api.IncomingMessage
	4, // 2: pubsub.api.SubscribeResponse.outgoing_status:type_name -> pubsub.api.OutgoingStatus
	3, // 3: pubsub.api.SubscribeResponse.subscription_status:type_name -> pubsub.api.SubscriptionStatus
	0, // 4: pubsub.api.PubSubService.Subscribe:input_type -> pubsub.api.SubscribeRequest
	2, // 5: pubsub.api.PubSubService.Subscribe:output_type -> pubsub.api.SubscribeResponse
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_init() }
func file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_init() {
	if File_github_com_aperturerobotics_bifrost_pubsub_api_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeRequest); i {
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
		file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PublishRequest); i {
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
		file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscribeResponse); i {
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
		file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SubscriptionStatus); i {
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
		file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OutgoingStatus); i {
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
		file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IncomingMessage); i {
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
			RawDescriptor: file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_depIdxs,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_pubsub_api_api_proto = out.File
	file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_pubsub_api_api_proto_depIdxs = nil
}
