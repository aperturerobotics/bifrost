// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0-devel
// 	protoc        v5.26.1
// source: github.com/aperturerobotics/bifrost/transport/webrtc/webrtc.proto

package webrtc

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/aperturerobotics/bifrost/stream/srpc/client"
	dialer "github.com/aperturerobotics/bifrost/transport/common/dialer"
	quic "github.com/aperturerobotics/bifrost/transport/common/quic"
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

// IceTransportPolicy contains the set of allowed ICE transport policies.
type IceTransportPolicy int32

const (
	// IceTransportPolicy_ALL allows any kind of ICE candidate.
	IceTransportPolicy_IceTransportPolicy_ALL IceTransportPolicy = 0
	// IceTransportPolicy_RELAY allows only media relay candidates (TURN).
	IceTransportPolicy_IceTransportPolicy_RELAY IceTransportPolicy = 1
)

// Enum value maps for IceTransportPolicy.
var (
	IceTransportPolicy_name = map[int32]string{
		0: "IceTransportPolicy_ALL",
		1: "IceTransportPolicy_RELAY",
	}
	IceTransportPolicy_value = map[string]int32{
		"IceTransportPolicy_ALL":   0,
		"IceTransportPolicy_RELAY": 1,
	}
)

func (x IceTransportPolicy) Enum() *IceTransportPolicy {
	p := new(IceTransportPolicy)
	*p = x
	return p
}

func (x IceTransportPolicy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (IceTransportPolicy) Descriptor() protoreflect.EnumDescriptor {
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_enumTypes[0].Descriptor()
}

func (IceTransportPolicy) Type() protoreflect.EnumType {
	return &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_enumTypes[0]
}

func (x IceTransportPolicy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use IceTransportPolicy.Descriptor instead.
func (IceTransportPolicy) EnumDescriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP(), []int{0}
}

// Config is the configuration for the WebRTC Signal RPC transport.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// SignalingId is the signaling channel identifier.
	// Cannot be empty.
	SignalingId string `protobuf:"bytes,1,opt,name=signaling_id,json=signalingId,proto3" json:"signaling_id,omitempty"`
	// TransportPeerId sets the peer ID to attach the transport to.
	// If unset, attaches to any running peer with a private key.
	// Must match the transport peer ID of the signaling transport.
	TransportPeerId string `protobuf:"bytes,2,opt,name=transport_peer_id,json=transportPeerId,proto3" json:"transport_peer_id,omitempty"`
	// TransportType overrides the transport type id for dial addresses.
	// Defaults to "webrtc"
	// Configures the scheme for addr matching to this transport.
	// E.x.: webrtc://
	TransportType string `protobuf:"bytes,3,opt,name=transport_type,json=transportType,proto3" json:"transport_type,omitempty"`
	// Quic contains the quic protocol options.
	//
	// The WebRTC transport always disables FEC and several other UDP-centric
	// features which are unnecessary due to the "reliable" nature of WebRTC.
	Quic *quic.Opts `protobuf:"bytes,4,opt,name=quic,proto3" json:"quic,omitempty"`
	// WebRtc contains the WebRTC protocol options.
	WebRtc *WebRtcConfig `protobuf:"bytes,5,opt,name=web_rtc,json=webRtc,proto3" json:"web_rtc,omitempty"`
	// Backoff is the backoff config for connecting to a PeerConnection.
	// If unset, defaults to reasonable defaults.
	Backoff *backoff.Backoff `protobuf:"bytes,6,opt,name=backoff,proto3" json:"backoff,omitempty"`
	// Dialers maps peer IDs to dialers.
	// This allows mapping which peer ID should be dialed via this transport.
	Dialers map[string]*dialer.DialerOpts `protobuf:"bytes,7,rep,name=dialers,proto3" json:"dialers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// AllPeers tells the transport to attempt to negotiate a WebRTC session with
	// any peer, even those not listed in the Dialers map.
	AllPeers bool `protobuf:"varint,8,opt,name=all_peers,json=allPeers,proto3" json:"all_peers,omitempty"`
	// DisableListen disables listening for incoming Links.
	// If set, we will only dial out, not accept incoming links.
	DisableListen bool `protobuf:"varint,9,opt,name=disable_listen,json=disableListen,proto3" json:"disable_listen,omitempty"`
	// BlockPeers is a list of peer ids that will not be contacted via this transport.
	BlockPeers []string `protobuf:"bytes,10,rep,name=block_peers,json=blockPeers,proto3" json:"block_peers,omitempty"`
	// Verbose enables very verbose logging.
	Verbose bool `protobuf:"varint,11,opt,name=verbose,proto3" json:"verbose,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[0]
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
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetSignalingId() string {
	if x != nil {
		return x.SignalingId
	}
	return ""
}

func (x *Config) GetTransportPeerId() string {
	if x != nil {
		return x.TransportPeerId
	}
	return ""
}

func (x *Config) GetTransportType() string {
	if x != nil {
		return x.TransportType
	}
	return ""
}

func (x *Config) GetQuic() *quic.Opts {
	if x != nil {
		return x.Quic
	}
	return nil
}

func (x *Config) GetWebRtc() *WebRtcConfig {
	if x != nil {
		return x.WebRtc
	}
	return nil
}

func (x *Config) GetBackoff() *backoff.Backoff {
	if x != nil {
		return x.Backoff
	}
	return nil
}

func (x *Config) GetDialers() map[string]*dialer.DialerOpts {
	if x != nil {
		return x.Dialers
	}
	return nil
}

func (x *Config) GetAllPeers() bool {
	if x != nil {
		return x.AllPeers
	}
	return false
}

func (x *Config) GetDisableListen() bool {
	if x != nil {
		return x.DisableListen
	}
	return false
}

func (x *Config) GetBlockPeers() []string {
	if x != nil {
		return x.BlockPeers
	}
	return nil
}

func (x *Config) GetVerbose() bool {
	if x != nil {
		return x.Verbose
	}
	return false
}

// WebRtcConfig configures the WebRTC PeerConnection.
type WebRtcConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// IceServers contains the list of ICE servers to use.
	IceServers []*IceServerConfig `protobuf:"bytes,1,rep,name=ice_servers,json=iceServers,proto3" json:"ice_servers,omitempty"`
	// IceTransportPolicy defines the policy for permitted ICE candidates.
	// Optional.
	IceTransportPolicy IceTransportPolicy `protobuf:"varint,2,opt,name=ice_transport_policy,json=iceTransportPolicy,proto3,enum=webrtc.IceTransportPolicy" json:"ice_transport_policy,omitempty"`
	// IceCandidatePoolSize defines the size of the prefetched ICE pool.
	// Optional.
	IceCandidatePoolSize uint32 `protobuf:"varint,3,opt,name=ice_candidate_pool_size,json=iceCandidatePoolSize,proto3" json:"ice_candidate_pool_size,omitempty"`
}

func (x *WebRtcConfig) Reset() {
	*x = WebRtcConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebRtcConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebRtcConfig) ProtoMessage() {}

func (x *WebRtcConfig) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebRtcConfig.ProtoReflect.Descriptor instead.
func (*WebRtcConfig) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP(), []int{1}
}

func (x *WebRtcConfig) GetIceServers() []*IceServerConfig {
	if x != nil {
		return x.IceServers
	}
	return nil
}

func (x *WebRtcConfig) GetIceTransportPolicy() IceTransportPolicy {
	if x != nil {
		return x.IceTransportPolicy
	}
	return IceTransportPolicy_IceTransportPolicy_ALL
}

func (x *WebRtcConfig) GetIceCandidatePoolSize() uint32 {
	if x != nil {
		return x.IceCandidatePoolSize
	}
	return 0
}

// IceServer is a WebRTC ICE server config.
type IceServerConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Urls is the list of URLs for the ICE server.
	//
	// Format: stun:{url} or turn:{url} or turns:{url}?transport=tcp
	// Examples:
	// - stun:stun.l.google.com:19302
	// - stun:stun.stunprotocol.org:3478
	// - turns:google.de?transport=tcp
	Urls []string `protobuf:"bytes,1,rep,name=urls,proto3" json:"urls,omitempty"`
	// Username is the username for the ICE server.
	Username string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	// Credential contains the ice server credential, if any.
	//
	// Types that are assignable to Credential:
	//
	//	*IceServerConfig_Password
	//	*IceServerConfig_Oauth
	Credential isIceServerConfig_Credential `protobuf_oneof:"credential"`
}

func (x *IceServerConfig) Reset() {
	*x = IceServerConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IceServerConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IceServerConfig) ProtoMessage() {}

func (x *IceServerConfig) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IceServerConfig.ProtoReflect.Descriptor instead.
func (*IceServerConfig) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP(), []int{2}
}

func (x *IceServerConfig) GetUrls() []string {
	if x != nil {
		return x.Urls
	}
	return nil
}

func (x *IceServerConfig) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (m *IceServerConfig) GetCredential() isIceServerConfig_Credential {
	if m != nil {
		return m.Credential
	}
	return nil
}

func (x *IceServerConfig) GetPassword() string {
	if x, ok := x.GetCredential().(*IceServerConfig_Password); ok {
		return x.Password
	}
	return ""
}

func (x *IceServerConfig) GetOauth() *IceServerConfig_OauthCredential {
	if x, ok := x.GetCredential().(*IceServerConfig_Oauth); ok {
		return x.Oauth
	}
	return nil
}

type isIceServerConfig_Credential interface {
	isIceServerConfig_Credential()
}

type IceServerConfig_Password struct {
	// Password contains the ICE server password.
	Password string `protobuf:"bytes,3,opt,name=password,proto3,oneof"`
}

type IceServerConfig_Oauth struct {
	// Oauth contains an OAuth credential.
	Oauth *IceServerConfig_OauthCredential `protobuf:"bytes,4,opt,name=oauth,proto3,oneof"`
}

func (*IceServerConfig_Password) isIceServerConfig_Credential() {}

func (*IceServerConfig_Oauth) isIceServerConfig_Credential() {}

// WebRtcSignal is a WebRTC Signaling message sent via the Signaling channel.
type WebRtcSignal struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Body is the body of the message.
	//
	// Types that are assignable to Body:
	//
	//	*WebRtcSignal_RequestOffer
	//	*WebRtcSignal_Sdp
	//	*WebRtcSignal_Ice
	Body isWebRtcSignal_Body `protobuf_oneof:"body"`
}

func (x *WebRtcSignal) Reset() {
	*x = WebRtcSignal{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebRtcSignal) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebRtcSignal) ProtoMessage() {}

func (x *WebRtcSignal) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebRtcSignal.ProtoReflect.Descriptor instead.
func (*WebRtcSignal) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP(), []int{3}
}

func (m *WebRtcSignal) GetBody() isWebRtcSignal_Body {
	if m != nil {
		return m.Body
	}
	return nil
}

func (x *WebRtcSignal) GetRequestOffer() uint64 {
	if x, ok := x.GetBody().(*WebRtcSignal_RequestOffer); ok {
		return x.RequestOffer
	}
	return 0
}

func (x *WebRtcSignal) GetSdp() *WebRtcSdp {
	if x, ok := x.GetBody().(*WebRtcSignal_Sdp); ok {
		return x.Sdp
	}
	return nil
}

func (x *WebRtcSignal) GetIce() *WebRtcIce {
	if x, ok := x.GetBody().(*WebRtcSignal_Ice); ok {
		return x.Ice
	}
	return nil
}

type isWebRtcSignal_Body interface {
	isWebRtcSignal_Body()
}

type WebRtcSignal_RequestOffer struct {
	// RequestOffer requests a new offer from the offerer with the local session seqno.
	// Incremented when negotiation is needed (something changes about the session).
	RequestOffer uint64 `protobuf:"varint,1,opt,name=request_offer,json=requestOffer,proto3,oneof"`
}

type WebRtcSignal_Sdp struct {
	// Sdp contains the sdp offer or answer.
	Sdp *WebRtcSdp `protobuf:"bytes,2,opt,name=sdp,proto3,oneof"`
}

type WebRtcSignal_Ice struct {
	// Ice contains an ICE candidate.
	Ice *WebRtcIce `protobuf:"bytes,3,opt,name=ice,proto3,oneof"`
}

func (*WebRtcSignal_RequestOffer) isWebRtcSignal_Body() {}

func (*WebRtcSignal_Sdp) isWebRtcSignal_Body() {}

func (*WebRtcSignal_Ice) isWebRtcSignal_Body() {}

// WebRtcSdp contains the SDP offer or answer.
type WebRtcSdp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TxSeqno is the sequence number of the transmitting peer.
	// The receiver should update the local seqno to match.
	TxSeqno uint64 `protobuf:"varint,1,opt,name=tx_seqno,json=txSeqno,proto3" json:"tx_seqno,omitempty"`
	// SdpType is the string encoded type of the sdp.
	// Examples: "offer" "answer"
	SdpType string `protobuf:"bytes,2,opt,name=sdp_type,json=sdpType,proto3" json:"sdp_type,omitempty"`
	// Sdp contains the WebRTC session description.
	Sdp string `protobuf:"bytes,3,opt,name=sdp,proto3" json:"sdp,omitempty"`
}

func (x *WebRtcSdp) Reset() {
	*x = WebRtcSdp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebRtcSdp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebRtcSdp) ProtoMessage() {}

func (x *WebRtcSdp) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebRtcSdp.ProtoReflect.Descriptor instead.
func (*WebRtcSdp) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP(), []int{4}
}

func (x *WebRtcSdp) GetTxSeqno() uint64 {
	if x != nil {
		return x.TxSeqno
	}
	return 0
}

func (x *WebRtcSdp) GetSdpType() string {
	if x != nil {
		return x.SdpType
	}
	return ""
}

func (x *WebRtcSdp) GetSdp() string {
	if x != nil {
		return x.Sdp
	}
	return ""
}

// WebRtcIce contains an ICE candidate.
type WebRtcIce struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Candidate contains the JSON-encoded ICE candidate.
	Candidate string `protobuf:"bytes,1,opt,name=candidate,proto3" json:"candidate,omitempty"`
}

func (x *WebRtcIce) Reset() {
	*x = WebRtcIce{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebRtcIce) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebRtcIce) ProtoMessage() {}

func (x *WebRtcIce) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebRtcIce.ProtoReflect.Descriptor instead.
func (*WebRtcIce) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP(), []int{5}
}

func (x *WebRtcIce) GetCandidate() string {
	if x != nil {
		return x.Candidate
	}
	return ""
}

// OauthCredential is an OAuth credential information for the ICE server.
type IceServerConfig_OauthCredential struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// MacKey is a base64-url format.
	MacKey string `protobuf:"bytes,1,opt,name=mac_key,json=macKey,proto3" json:"mac_key,omitempty"`
	// AccessToken is the access token in base64-encoded format.
	AccessToken string `protobuf:"bytes,2,opt,name=access_token,json=accessToken,proto3" json:"access_token,omitempty"`
}

func (x *IceServerConfig_OauthCredential) Reset() {
	*x = IceServerConfig_OauthCredential{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IceServerConfig_OauthCredential) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IceServerConfig_OauthCredential) ProtoMessage() {}

func (x *IceServerConfig_OauthCredential) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IceServerConfig_OauthCredential.ProtoReflect.Descriptor instead.
func (*IceServerConfig_OauthCredential) Descriptor() ([]byte, []int) {
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP(), []int{2, 0}
}

func (x *IceServerConfig_OauthCredential) GetMacKey() string {
	if x != nil {
		return x.MacKey
	}
	return ""
}

func (x *IceServerConfig_OauthCredential) GetAccessToken() string {
	if x != nil {
		return x.AccessToken
	}
	return ""
}

var File_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto protoreflect.FileDescriptor

var file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDesc = []byte{
	0x0a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69,
	0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f,
	0x77, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2f, 0x77, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x06, 0x77, 0x65, 0x62, 0x72, 0x74, 0x63, 0x1a, 0x44, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65,
	0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69, 0x66, 0x72, 0x6f, 0x73, 0x74,
	0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2f, 0x71, 0x75, 0x69, 0x63, 0x2f, 0x71, 0x75, 0x69, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x48, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70,
	0x65, 0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62,
	0x69, 0x66, 0x72, 0x6f, 0x73, 0x74, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74,
	0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x2f, 0x64,
	0x69, 0x61, 0x6c, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x43, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65, 0x72, 0x74, 0x75, 0x72, 0x65,
	0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x62, 0x69, 0x66, 0x72, 0x6f, 0x73, 0x74,
	0x2f, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x2f, 0x73, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x70, 0x65,
	0x72, 0x74, 0x75, 0x72, 0x65, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x69, 0x63, 0x73, 0x2f, 0x75, 0x74,
	0x69, 0x6c, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x2f, 0x62, 0x61, 0x63, 0x6b, 0x6f,
	0x66, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x89, 0x04, 0x0a, 0x06, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e, 0x67,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x69, 0x67, 0x6e, 0x61,
	0x6c, 0x69, 0x6e, 0x67, 0x49, 0x64, 0x12, 0x2a, 0x0a, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x50, 0x65, 0x65, 0x72,
	0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x5f,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x72, 0x61, 0x6e,
	0x73, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x28, 0x0a, 0x04, 0x71, 0x75, 0x69,
	0x63, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x2e, 0x71, 0x75, 0x69, 0x63, 0x2e, 0x4f, 0x70, 0x74, 0x73, 0x52, 0x04, 0x71,
	0x75, 0x69, 0x63, 0x12, 0x2d, 0x0a, 0x07, 0x77, 0x65, 0x62, 0x5f, 0x72, 0x74, 0x63, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x77, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2e, 0x57, 0x65,
	0x62, 0x52, 0x74, 0x63, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x06, 0x77, 0x65, 0x62, 0x52,
	0x74, 0x63, 0x12, 0x2a, 0x0a, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x2e, 0x42, 0x61,
	0x63, 0x6b, 0x6f, 0x66, 0x66, 0x52, 0x07, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x12, 0x35,
	0x0a, 0x07, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1b, 0x2e, 0x77, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e,
	0x44, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x64, 0x69,
	0x61, 0x6c, 0x65, 0x72, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x61, 0x6c, 0x6c, 0x5f, 0x70, 0x65, 0x65,
	0x72, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x61, 0x6c, 0x6c, 0x50, 0x65, 0x65,
	0x72, 0x73, 0x12, 0x25, 0x0a, 0x0e, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x6c, 0x69,
	0x73, 0x74, 0x65, 0x6e, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x64, 0x69, 0x73, 0x61,
	0x62, 0x6c, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x5f, 0x70, 0x65, 0x65, 0x72, 0x73, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x50, 0x65, 0x65, 0x72, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65,
	0x72, 0x62, 0x6f, 0x73, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x76, 0x65, 0x72,
	0x62, 0x6f, 0x73, 0x65, 0x1a, 0x4e, 0x0a, 0x0c, 0x44, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x72, 0x2e, 0x44,
	0x69, 0x61, 0x6c, 0x65, 0x72, 0x4f, 0x70, 0x74, 0x73, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x22, 0xcd, 0x01, 0x0a, 0x0c, 0x57, 0x65, 0x62, 0x52, 0x74, 0x63, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x38, 0x0a, 0x0b, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x77, 0x65, 0x62,
	0x72, 0x74, 0x63, 0x2e, 0x49, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x52, 0x0a, 0x69, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x12,
	0x4c, 0x0a, 0x14, 0x69, 0x63, 0x65, 0x5f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74,
	0x5f, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e,
	0x77, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2e, 0x49, 0x63, 0x65, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x6f, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x12, 0x69, 0x63, 0x65, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x35, 0x0a,
	0x17, 0x69, 0x63, 0x65, 0x5f, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x70,
	0x6f, 0x6f, 0x6c, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x14,
	0x69, 0x63, 0x65, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x50, 0x6f, 0x6f, 0x6c,
	0x53, 0x69, 0x7a, 0x65, 0x22, 0xfd, 0x01, 0x0a, 0x0f, 0x49, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x72, 0x6c, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x75, 0x72, 0x6c, 0x73, 0x12, 0x1a, 0x0a, 0x08,
	0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73,
	0x77, 0x6f, 0x72, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x08, 0x70, 0x61,
	0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x3f, 0x0a, 0x05, 0x6f, 0x61, 0x75, 0x74, 0x68, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x77, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2e, 0x49,
	0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x4f,
	0x61, 0x75, 0x74, 0x68, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x48, 0x00,
	0x52, 0x05, 0x6f, 0x61, 0x75, 0x74, 0x68, 0x1a, 0x4d, 0x0a, 0x0f, 0x4f, 0x61, 0x75, 0x74, 0x68,
	0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x12, 0x17, 0x0a, 0x07, 0x6d, 0x61,
	0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x61, 0x63,
	0x4b, 0x65, 0x79, 0x12, 0x21, 0x0a, 0x0c, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x74, 0x6f,
	0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x42, 0x0c, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x61, 0x6c, 0x22, 0x8b, 0x01, 0x0a, 0x0c, 0x57, 0x65, 0x62, 0x52, 0x74, 0x63, 0x53,
	0x69, 0x67, 0x6e, 0x61, 0x6c, 0x12, 0x25, 0x0a, 0x0d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x5f, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x48, 0x00, 0x52, 0x0c,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4f, 0x66, 0x66, 0x65, 0x72, 0x12, 0x25, 0x0a, 0x03,
	0x73, 0x64, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x77, 0x65, 0x62, 0x72,
	0x74, 0x63, 0x2e, 0x57, 0x65, 0x62, 0x52, 0x74, 0x63, 0x53, 0x64, 0x70, 0x48, 0x00, 0x52, 0x03,
	0x73, 0x64, 0x70, 0x12, 0x25, 0x0a, 0x03, 0x69, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x11, 0x2e, 0x77, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2e, 0x57, 0x65, 0x62, 0x52, 0x74, 0x63,
	0x49, 0x63, 0x65, 0x48, 0x00, 0x52, 0x03, 0x69, 0x63, 0x65, 0x42, 0x06, 0x0a, 0x04, 0x62, 0x6f,
	0x64, 0x79, 0x22, 0x53, 0x0a, 0x09, 0x57, 0x65, 0x62, 0x52, 0x74, 0x63, 0x53, 0x64, 0x70, 0x12,
	0x19, 0x0a, 0x08, 0x74, 0x78, 0x5f, 0x73, 0x65, 0x71, 0x6e, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x07, 0x74, 0x78, 0x53, 0x65, 0x71, 0x6e, 0x6f, 0x12, 0x19, 0x0a, 0x08, 0x73, 0x64,
	0x70, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x64,
	0x70, 0x54, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x64, 0x70, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x73, 0x64, 0x70, 0x22, 0x29, 0x0a, 0x09, 0x57, 0x65, 0x62, 0x52, 0x74,
	0x63, 0x49, 0x63, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2a, 0x4e, 0x0a, 0x12, 0x49, 0x63, 0x65, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f,
	0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12, 0x1a, 0x0a, 0x16, 0x49, 0x63, 0x65, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x70, 0x6f, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x5f, 0x41,
	0x4c, 0x4c, 0x10, 0x00, 0x12, 0x1c, 0x0a, 0x18, 0x49, 0x63, 0x65, 0x54, 0x72, 0x61, 0x6e, 0x73,
	0x70, 0x6f, 0x72, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x5f, 0x52, 0x45, 0x4c, 0x41, 0x59,
	0x10, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescOnce sync.Once
	file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescData = file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDesc
)

func file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescGZIP() []byte {
	file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescOnce.Do(func() {
		file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescData)
	})
	return file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDescData
}

var file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_goTypes = []interface{}{
	(IceTransportPolicy)(0),                 // 0: webrtc.IceTransportPolicy
	(*Config)(nil),                          // 1: webrtc.Config
	(*WebRtcConfig)(nil),                    // 2: webrtc.WebRtcConfig
	(*IceServerConfig)(nil),                 // 3: webrtc.IceServerConfig
	(*WebRtcSignal)(nil),                    // 4: webrtc.WebRtcSignal
	(*WebRtcSdp)(nil),                       // 5: webrtc.WebRtcSdp
	(*WebRtcIce)(nil),                       // 6: webrtc.WebRtcIce
	nil,                                     // 7: webrtc.Config.DialersEntry
	(*IceServerConfig_OauthCredential)(nil), // 8: webrtc.IceServerConfig.OauthCredential
	(*quic.Opts)(nil),                       // 9: transport.quic.Opts
	(*backoff.Backoff)(nil),                 // 10: backoff.Backoff
	(*dialer.DialerOpts)(nil),               // 11: dialer.DialerOpts
}
var file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_depIdxs = []int32{
	9,  // 0: webrtc.Config.quic:type_name -> transport.quic.Opts
	2,  // 1: webrtc.Config.web_rtc:type_name -> webrtc.WebRtcConfig
	10, // 2: webrtc.Config.backoff:type_name -> backoff.Backoff
	7,  // 3: webrtc.Config.dialers:type_name -> webrtc.Config.DialersEntry
	3,  // 4: webrtc.WebRtcConfig.ice_servers:type_name -> webrtc.IceServerConfig
	0,  // 5: webrtc.WebRtcConfig.ice_transport_policy:type_name -> webrtc.IceTransportPolicy
	8,  // 6: webrtc.IceServerConfig.oauth:type_name -> webrtc.IceServerConfig.OauthCredential
	5,  // 7: webrtc.WebRtcSignal.sdp:type_name -> webrtc.WebRtcSdp
	6,  // 8: webrtc.WebRtcSignal.ice:type_name -> webrtc.WebRtcIce
	11, // 9: webrtc.Config.DialersEntry.value:type_name -> dialer.DialerOpts
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_init() }
func file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_init() {
	if File_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebRtcConfig); i {
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
		file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IceServerConfig); i {
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
		file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebRtcSignal); i {
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
		file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebRtcSdp); i {
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
		file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebRtcIce); i {
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
		file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IceServerConfig_OauthCredential); i {
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
	file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*IceServerConfig_Password)(nil),
		(*IceServerConfig_Oauth)(nil),
	}
	file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*WebRtcSignal_RequestOffer)(nil),
		(*WebRtcSignal_Sdp)(nil),
		(*WebRtcSignal_Ice)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_goTypes,
		DependencyIndexes: file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_depIdxs,
		EnumInfos:         file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_enumTypes,
		MessageInfos:      file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_msgTypes,
	}.Build()
	File_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto = out.File
	file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_rawDesc = nil
	file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_goTypes = nil
	file_github_com_aperturerobotics_bifrost_transport_webrtc_webrtc_proto_depIdxs = nil
}
