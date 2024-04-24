// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.6.0
// source: github.com/aperturerobotics/bifrost/peer/api/api.proto

package peer_api

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	controller "github.com/aperturerobotics/bifrost/peer/controller"
	exec "github.com/aperturerobotics/controllerbus/controller/exec"
	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// IdentifyRequest is a request to load an identity.
type IdentifyRequest struct {
	unknownFields []byte
	// Config is the request to configure the peer controller.
	Config *controller.Config `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *IdentifyRequest) Reset() {
	*x = IdentifyRequest{}
}

func (*IdentifyRequest) ProtoMessage() {}

func (x *IdentifyRequest) GetConfig() *controller.Config {
	if x != nil {
		return x.Config
	}
	return nil
}

// IdentifyResponse is a response to an identify request.
type IdentifyResponse struct {
	unknownFields []byte
	// ControllerStatus is the status of the peer controller.
	ControllerStatus exec.ControllerStatus `protobuf:"varint,1,opt,name=controller_status,json=controllerStatus,proto3" json:"controllerStatus,omitempty"`
}

func (x *IdentifyResponse) Reset() {
	*x = IdentifyResponse{}
}

func (*IdentifyResponse) ProtoMessage() {}

func (x *IdentifyResponse) GetControllerStatus() exec.ControllerStatus {
	if x != nil {
		return x.ControllerStatus
	}
	return exec.ControllerStatus(0)
}

// GetPeerInfoRequest is the request type for GetPeerInfo.
type GetPeerInfoRequest struct {
	unknownFields []byte
	// PeerId restricts the response to a specific peer ID.
	PeerId string `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peerId,omitempty"`
}

func (x *GetPeerInfoRequest) Reset() {
	*x = GetPeerInfoRequest{}
}

func (*GetPeerInfoRequest) ProtoMessage() {}

func (x *GetPeerInfoRequest) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

// PeerInfo is basic information about a peer.
type PeerInfo struct {
	unknownFields []byte
	// PeerId is the b58 peer ID.
	PeerId string `protobuf:"bytes,1,opt,name=peer_id,json=peerId,proto3" json:"peerId,omitempty"`
}

func (x *PeerInfo) Reset() {
	*x = PeerInfo{}
}

func (*PeerInfo) ProtoMessage() {}

func (x *PeerInfo) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

// GetPeerInfoResponse is the response type for GetPeerInfo.
type GetPeerInfoResponse struct {
	unknownFields []byte
	// LocalPeers is the set of peers loaded.
	LocalPeers []*PeerInfo `protobuf:"bytes,1,rep,name=local_peers,json=localPeers,proto3" json:"localPeers,omitempty"`
}

func (x *GetPeerInfoResponse) Reset() {
	*x = GetPeerInfoResponse{}
}

func (*GetPeerInfoResponse) ProtoMessage() {}

func (x *GetPeerInfoResponse) GetLocalPeers() []*PeerInfo {
	if x != nil {
		return x.LocalPeers
	}
	return nil
}

func (m *IdentifyRequest) CloneVT() *IdentifyRequest {
	if m == nil {
		return (*IdentifyRequest)(nil)
	}
	r := new(IdentifyRequest)
	if rhs := m.Config; rhs != nil {
		r.Config = rhs.CloneVT()
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *IdentifyRequest) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *IdentifyResponse) CloneVT() *IdentifyResponse {
	if m == nil {
		return (*IdentifyResponse)(nil)
	}
	r := new(IdentifyResponse)
	r.ControllerStatus = m.ControllerStatus
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *IdentifyResponse) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *GetPeerInfoRequest) CloneVT() *GetPeerInfoRequest {
	if m == nil {
		return (*GetPeerInfoRequest)(nil)
	}
	r := new(GetPeerInfoRequest)
	r.PeerId = m.PeerId
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *GetPeerInfoRequest) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *PeerInfo) CloneVT() *PeerInfo {
	if m == nil {
		return (*PeerInfo)(nil)
	}
	r := new(PeerInfo)
	r.PeerId = m.PeerId
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *PeerInfo) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *GetPeerInfoResponse) CloneVT() *GetPeerInfoResponse {
	if m == nil {
		return (*GetPeerInfoResponse)(nil)
	}
	r := new(GetPeerInfoResponse)
	if rhs := m.LocalPeers; rhs != nil {
		tmpContainer := make([]*PeerInfo, len(rhs))
		for k, v := range rhs {
			tmpContainer[k] = v.CloneVT()
		}
		r.LocalPeers = tmpContainer
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *GetPeerInfoResponse) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *IdentifyRequest) EqualVT(that *IdentifyRequest) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if !this.Config.EqualVT(that.Config) {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *IdentifyRequest) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*IdentifyRequest)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *IdentifyResponse) EqualVT(that *IdentifyResponse) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.ControllerStatus != that.ControllerStatus {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *IdentifyResponse) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*IdentifyResponse)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *GetPeerInfoRequest) EqualVT(that *GetPeerInfoRequest) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.PeerId != that.PeerId {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *GetPeerInfoRequest) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*GetPeerInfoRequest)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *PeerInfo) EqualVT(that *PeerInfo) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.PeerId != that.PeerId {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *PeerInfo) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*PeerInfo)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *GetPeerInfoResponse) EqualVT(that *GetPeerInfoResponse) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if len(this.LocalPeers) != len(that.LocalPeers) {
		return false
	}
	for i, vx := range this.LocalPeers {
		vy := that.LocalPeers[i]
		if p, q := vx, vy; p != q {
			if p == nil {
				p = &PeerInfo{}
			}
			if q == nil {
				q = &PeerInfo{}
			}
			if !p.EqualVT(q) {
				return false
			}
		}
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *GetPeerInfoResponse) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*GetPeerInfoResponse)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the IdentifyRequest message to JSON.
func (x *IdentifyRequest) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Config != nil || s.HasField("config") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("config")
		x.Config.MarshalProtoJSON(s.WithField("config"))
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the IdentifyRequest to JSON.
func (x *IdentifyRequest) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the IdentifyRequest message from JSON.
func (x *IdentifyRequest) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "config":
			if s.ReadNil() {
				x.Config = nil
				return
			}
			x.Config = &controller.Config{}
			x.Config.UnmarshalProtoJSON(s.WithField("config", true))
		}
	})
}

// UnmarshalJSON unmarshals the IdentifyRequest from JSON.
func (x *IdentifyRequest) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the IdentifyResponse message to JSON.
func (x *IdentifyResponse) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.ControllerStatus != 0 || s.HasField("controllerStatus") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("controllerStatus")
		x.ControllerStatus.MarshalProtoJSON(s)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the IdentifyResponse to JSON.
func (x *IdentifyResponse) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the IdentifyResponse message from JSON.
func (x *IdentifyResponse) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "controller_status", "controllerStatus":
			s.AddField("controller_status")
			x.ControllerStatus.UnmarshalProtoJSON(s)
		}
	})
}

// UnmarshalJSON unmarshals the IdentifyResponse from JSON.
func (x *IdentifyResponse) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the GetPeerInfoRequest message to JSON.
func (x *GetPeerInfoRequest) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.PeerId != "" || s.HasField("peerId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("peerId")
		s.WriteString(x.PeerId)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the GetPeerInfoRequest to JSON.
func (x *GetPeerInfoRequest) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the GetPeerInfoRequest message from JSON.
func (x *GetPeerInfoRequest) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "peer_id", "peerId":
			s.AddField("peer_id")
			x.PeerId = s.ReadString()
		}
	})
}

// UnmarshalJSON unmarshals the GetPeerInfoRequest from JSON.
func (x *GetPeerInfoRequest) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the PeerInfo message to JSON.
func (x *PeerInfo) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.PeerId != "" || s.HasField("peerId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("peerId")
		s.WriteString(x.PeerId)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the PeerInfo to JSON.
func (x *PeerInfo) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the PeerInfo message from JSON.
func (x *PeerInfo) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "peer_id", "peerId":
			s.AddField("peer_id")
			x.PeerId = s.ReadString()
		}
	})
}

// UnmarshalJSON unmarshals the PeerInfo from JSON.
func (x *PeerInfo) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the GetPeerInfoResponse message to JSON.
func (x *GetPeerInfoResponse) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if len(x.LocalPeers) > 0 || s.HasField("localPeers") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("localPeers")
		s.WriteArrayStart()
		var wroteElement bool
		for _, element := range x.LocalPeers {
			s.WriteMoreIf(&wroteElement)
			element.MarshalProtoJSON(s.WithField("localPeers"))
		}
		s.WriteArrayEnd()
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the GetPeerInfoResponse to JSON.
func (x *GetPeerInfoResponse) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the GetPeerInfoResponse message from JSON.
func (x *GetPeerInfoResponse) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "local_peers", "localPeers":
			s.AddField("local_peers")
			if s.ReadNil() {
				x.LocalPeers = nil
				return
			}
			s.ReadArray(func() {
				if s.ReadNil() {
					x.LocalPeers = append(x.LocalPeers, nil)
					return
				}
				v := &PeerInfo{}
				v.UnmarshalProtoJSON(s.WithField("local_peers", false))
				if s.Err() != nil {
					return
				}
				x.LocalPeers = append(x.LocalPeers, v)
			})
		}
	})
}

// UnmarshalJSON unmarshals the GetPeerInfoResponse from JSON.
func (x *GetPeerInfoResponse) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *IdentifyRequest) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IdentifyRequest) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *IdentifyRequest) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if m.Config != nil {
		size, err := m.Config.MarshalToSizedBufferVT(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *IdentifyResponse) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *IdentifyResponse) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *IdentifyResponse) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if m.ControllerStatus != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.ControllerStatus))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *GetPeerInfoRequest) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetPeerInfoRequest) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *GetPeerInfoRequest) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.PeerId) > 0 {
		i -= len(m.PeerId)
		copy(dAtA[i:], m.PeerId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.PeerId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *PeerInfo) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PeerInfo) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *PeerInfo) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.PeerId) > 0 {
		i -= len(m.PeerId)
		copy(dAtA[i:], m.PeerId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.PeerId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *GetPeerInfoResponse) MarshalVT() (dAtA []byte, err error) {
	if m == nil {
		return nil, nil
	}
	size := m.SizeVT()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBufferVT(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetPeerInfoResponse) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *GetPeerInfoResponse) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
	if m == nil {
		return 0, nil
	}
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.unknownFields != nil {
		i -= len(m.unknownFields)
		copy(dAtA[i:], m.unknownFields)
	}
	if len(m.LocalPeers) > 0 {
		for iNdEx := len(m.LocalPeers) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.LocalPeers[iNdEx].MarshalToSizedBufferVT(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *IdentifyRequest) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Config != nil {
		l = m.Config.SizeVT()
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (m *IdentifyResponse) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ControllerStatus != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.ControllerStatus))
	}
	n += len(m.unknownFields)
	return n
}

func (m *GetPeerInfoRequest) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (m *PeerInfo) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PeerId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (m *GetPeerInfoResponse) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.LocalPeers) > 0 {
		for _, e := range m.LocalPeers {
			l = e.SizeVT()
			n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
		}
	}
	n += len(m.unknownFields)
	return n
}

func (x *IdentifyRequest) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("IdentifyRequest { ")
	if x.Config != nil {
		sb.WriteString(" config: ")
		sb.WriteString(x.Config.MarshalProtoText())
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *IdentifyRequest) String() string {
	return x.MarshalProtoText()
}
func (x *IdentifyResponse) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("IdentifyResponse { ")
	if x.ControllerStatus != 0 {
		sb.WriteString(" controller_status: ")
		sb.WriteString(exec.ControllerStatus(x.ControllerStatus).String())
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *IdentifyResponse) String() string {
	return x.MarshalProtoText()
}
func (x *GetPeerInfoRequest) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("GetPeerInfoRequest { ")
	if x.PeerId != "" {
		sb.WriteString(" peer_id: ")
		sb.WriteString(strconv.Quote(x.PeerId))
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *GetPeerInfoRequest) String() string {
	return x.MarshalProtoText()
}
func (x *PeerInfo) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("PeerInfo { ")
	if x.PeerId != "" {
		sb.WriteString(" peer_id: ")
		sb.WriteString(strconv.Quote(x.PeerId))
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *PeerInfo) String() string {
	return x.MarshalProtoText()
}
func (x *GetPeerInfoResponse) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("GetPeerInfoResponse { ")
	if len(x.LocalPeers) > 0 {
		sb.WriteString(" local_peers: [")
		for i, v := range x.LocalPeers {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(v.MarshalProtoText())
		}
		sb.WriteString("]")
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *GetPeerInfoResponse) String() string {
	return x.MarshalProtoText()
}
func (m *IdentifyRequest) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protobuf_go_lite.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: IdentifyRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IdentifyRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Config", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Config == nil {
				m.Config = &controller.Config{}
			}
			if err := m.Config.UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *IdentifyResponse) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protobuf_go_lite.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: IdentifyResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: IdentifyResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ControllerStatus", wireType)
			}
			m.ControllerStatus = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ControllerStatus |= exec.ControllerStatus(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GetPeerInfoRequest) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protobuf_go_lite.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GetPeerInfoRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetPeerInfoRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PeerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *PeerInfo) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protobuf_go_lite.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PeerInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PeerInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PeerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GetPeerInfoResponse) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return protobuf_go_lite.ErrIntOverflow
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GetPeerInfoResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetPeerInfoResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LocalPeers", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LocalPeers = append(m.LocalPeers, &PeerInfo{})
			if err := m.LocalPeers[len(m.LocalPeers)-1].UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.unknownFields = append(m.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
