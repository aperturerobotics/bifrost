// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.8.0
// source: github.com/aperturerobotics/bifrost/stream/srpc/server/server.proto

package stream_srpc_server

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// Config configures the server for the srpc service.
type Config struct {
	unknownFields []byte
	// PeerIds are the list of peer IDs to listen on.
	// If empty, allows any incoming peer id w/ the protocol id(s).
	PeerIds []string `protobuf:"bytes,1,rep,name=peer_ids,json=peerIds,proto3" json:"peerIds,omitempty"`
	// ProtocolIds is the list of protocol ids to listen on.
	// If empty, no incoming streams will be accepted.
	ProtocolIds []string `protobuf:"bytes,2,rep,name=protocol_ids,json=protocolIds,proto3" json:"protocolIds,omitempty"`
	// DisableEstablishLink disables adding an EstablishLink directive for each incoming peer.
	DisableEstablishLink bool `protobuf:"varint,3,opt,name=disable_establish_link,json=disableEstablishLink,proto3" json:"disableEstablishLink,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
}

func (*Config) ProtoMessage() {}

func (x *Config) GetPeerIds() []string {
	if x != nil {
		return x.PeerIds
	}
	return nil
}

func (x *Config) GetProtocolIds() []string {
	if x != nil {
		return x.ProtocolIds
	}
	return nil
}

func (x *Config) GetDisableEstablishLink() bool {
	if x != nil {
		return x.DisableEstablishLink
	}
	return false
}

func (m *Config) CloneVT() *Config {
	if m == nil {
		return (*Config)(nil)
	}
	r := new(Config)
	r.DisableEstablishLink = m.DisableEstablishLink
	if rhs := m.PeerIds; rhs != nil {
		tmpContainer := make([]string, len(rhs))
		copy(tmpContainer, rhs)
		r.PeerIds = tmpContainer
	}
	if rhs := m.ProtocolIds; rhs != nil {
		tmpContainer := make([]string, len(rhs))
		copy(tmpContainer, rhs)
		r.ProtocolIds = tmpContainer
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Config) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *Config) EqualVT(that *Config) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if len(this.PeerIds) != len(that.PeerIds) {
		return false
	}
	for i, vx := range this.PeerIds {
		vy := that.PeerIds[i]
		if vx != vy {
			return false
		}
	}
	if len(this.ProtocolIds) != len(that.ProtocolIds) {
		return false
	}
	for i, vx := range this.ProtocolIds {
		vy := that.ProtocolIds[i]
		if vx != vy {
			return false
		}
	}
	if this.DisableEstablishLink != that.DisableEstablishLink {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Config) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*Config)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the Config message to JSON.
func (x *Config) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if len(x.PeerIds) > 0 || s.HasField("peerIds") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("peerIds")
		s.WriteStringArray(x.PeerIds)
	}
	if len(x.ProtocolIds) > 0 || s.HasField("protocolIds") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("protocolIds")
		s.WriteStringArray(x.ProtocolIds)
	}
	if x.DisableEstablishLink || s.HasField("disableEstablishLink") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("disableEstablishLink")
		s.WriteBool(x.DisableEstablishLink)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the Config to JSON.
func (x *Config) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the Config message from JSON.
func (x *Config) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "peer_ids", "peerIds":
			s.AddField("peer_ids")
			if s.ReadNil() {
				x.PeerIds = nil
				return
			}
			x.PeerIds = s.ReadStringArray()
		case "protocol_ids", "protocolIds":
			s.AddField("protocol_ids")
			if s.ReadNil() {
				x.ProtocolIds = nil
				return
			}
			x.ProtocolIds = s.ReadStringArray()
		case "disable_establish_link", "disableEstablishLink":
			s.AddField("disable_establish_link")
			x.DisableEstablishLink = s.ReadBool()
		}
	})
}

// UnmarshalJSON unmarshals the Config from JSON.
func (x *Config) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *Config) MarshalVT() (dAtA []byte, err error) {
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

func (m *Config) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Config) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if m.DisableEstablishLink {
		i--
		if m.DisableEstablishLink {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x18
	}
	if len(m.ProtocolIds) > 0 {
		for iNdEx := len(m.ProtocolIds) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.ProtocolIds[iNdEx])
			copy(dAtA[i:], m.ProtocolIds[iNdEx])
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ProtocolIds[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.PeerIds) > 0 {
		for iNdEx := len(m.PeerIds) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.PeerIds[iNdEx])
			copy(dAtA[i:], m.PeerIds[iNdEx])
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.PeerIds[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *Config) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.PeerIds) > 0 {
		for _, s := range m.PeerIds {
			l = len(s)
			n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
		}
	}
	if len(m.ProtocolIds) > 0 {
		for _, s := range m.ProtocolIds {
			l = len(s)
			n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
		}
	}
	if m.DisableEstablishLink {
		n += 2
	}
	n += len(m.unknownFields)
	return n
}

func (x *Config) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Config {")
	if len(x.PeerIds) > 0 {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("peer_ids: [")
		for i, v := range x.PeerIds {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(strconv.Quote(v))
		}
		sb.WriteString("]")
	}
	if len(x.ProtocolIds) > 0 {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("protocol_ids: [")
		for i, v := range x.ProtocolIds {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(strconv.Quote(v))
		}
		sb.WriteString("]")
	}
	if x.DisableEstablishLink != false {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("disable_establish_link: ")
		sb.WriteString(strconv.FormatBool(x.DisableEstablishLink))
	}
	sb.WriteString("}")
	return sb.String()
}

func (x *Config) String() string {
	return x.MarshalProtoText()
}
func (m *Config) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: Config: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Config: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PeerIds", wireType)
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
			m.PeerIds = append(m.PeerIds, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtocolIds", wireType)
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
			m.ProtocolIds = append(m.ProtocolIds, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DisableEstablishLink", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.DisableEstablishLink = bool(v != 0)
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
