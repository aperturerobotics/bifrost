// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.9.1
// source: github.com/aperturerobotics/bifrost/stream/listening/listening.proto

package stream_listening

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// Config configures the listening controller.
type Config struct {
	unknownFields []byte
	// LocalPeerId is the peer ID to forward incoming connections with.
	// Can be empty.
	LocalPeerId string `protobuf:"bytes,1,opt,name=local_peer_id,json=localPeerId,proto3" json:"localPeerId,omitempty"`
	// RemotePeerId is the peer ID to forward incoming connections to.
	RemotePeerId string `protobuf:"bytes,2,opt,name=remote_peer_id,json=remotePeerId,proto3" json:"remotePeerId,omitempty"`
	// ProtocolId is the protocol ID to assign to incoming connections.
	ProtocolId string `protobuf:"bytes,3,opt,name=protocol_id,json=protocolId,proto3" json:"protocolId,omitempty"`
	// ListenMultiaddr is the listening multiaddress.
	ListenMultiaddr string `protobuf:"bytes,4,opt,name=listen_multiaddr,json=listenMultiaddr,proto3" json:"listenMultiaddr,omitempty"`
	// TransportId sets a transport ID constraint.
	// Can be empty.
	TransportId uint64 `protobuf:"varint,5,opt,name=transport_id,json=transportId,proto3" json:"transportId,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
}

func (*Config) ProtoMessage() {}

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

func (m *Config) CloneVT() *Config {
	if m == nil {
		return (*Config)(nil)
	}
	r := new(Config)
	r.LocalPeerId = m.LocalPeerId
	r.RemotePeerId = m.RemotePeerId
	r.ProtocolId = m.ProtocolId
	r.ListenMultiaddr = m.ListenMultiaddr
	r.TransportId = m.TransportId
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
	if this.LocalPeerId != that.LocalPeerId {
		return false
	}
	if this.RemotePeerId != that.RemotePeerId {
		return false
	}
	if this.ProtocolId != that.ProtocolId {
		return false
	}
	if this.ListenMultiaddr != that.ListenMultiaddr {
		return false
	}
	if this.TransportId != that.TransportId {
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
	if x.LocalPeerId != "" || s.HasField("localPeerId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("localPeerId")
		s.WriteString(x.LocalPeerId)
	}
	if x.RemotePeerId != "" || s.HasField("remotePeerId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("remotePeerId")
		s.WriteString(x.RemotePeerId)
	}
	if x.ProtocolId != "" || s.HasField("protocolId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("protocolId")
		s.WriteString(x.ProtocolId)
	}
	if x.ListenMultiaddr != "" || s.HasField("listenMultiaddr") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("listenMultiaddr")
		s.WriteString(x.ListenMultiaddr)
	}
	if x.TransportId != 0 || s.HasField("transportId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("transportId")
		s.WriteUint64(x.TransportId)
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
		case "local_peer_id", "localPeerId":
			s.AddField("local_peer_id")
			x.LocalPeerId = s.ReadString()
		case "remote_peer_id", "remotePeerId":
			s.AddField("remote_peer_id")
			x.RemotePeerId = s.ReadString()
		case "protocol_id", "protocolId":
			s.AddField("protocol_id")
			x.ProtocolId = s.ReadString()
		case "listen_multiaddr", "listenMultiaddr":
			s.AddField("listen_multiaddr")
			x.ListenMultiaddr = s.ReadString()
		case "transport_id", "transportId":
			s.AddField("transport_id")
			x.TransportId = s.ReadUint64()
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
	if m.TransportId != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.TransportId))
		i--
		dAtA[i] = 0x28
	}
	if len(m.ListenMultiaddr) > 0 {
		i -= len(m.ListenMultiaddr)
		copy(dAtA[i:], m.ListenMultiaddr)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ListenMultiaddr)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.ProtocolId) > 0 {
		i -= len(m.ProtocolId)
		copy(dAtA[i:], m.ProtocolId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ProtocolId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.RemotePeerId) > 0 {
		i -= len(m.RemotePeerId)
		copy(dAtA[i:], m.RemotePeerId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.RemotePeerId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.LocalPeerId) > 0 {
		i -= len(m.LocalPeerId)
		copy(dAtA[i:], m.LocalPeerId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.LocalPeerId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Config) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.LocalPeerId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.RemotePeerId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.ProtocolId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.ListenMultiaddr)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.TransportId != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.TransportId))
	}
	n += len(m.unknownFields)
	return n
}

func (x *Config) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Config {")
	if x.LocalPeerId != "" {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("local_peer_id: ")
		sb.WriteString(strconv.Quote(x.LocalPeerId))
	}
	if x.RemotePeerId != "" {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("remote_peer_id: ")
		sb.WriteString(strconv.Quote(x.RemotePeerId))
	}
	if x.ProtocolId != "" {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("protocol_id: ")
		sb.WriteString(strconv.Quote(x.ProtocolId))
	}
	if x.ListenMultiaddr != "" {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("listen_multiaddr: ")
		sb.WriteString(strconv.Quote(x.ListenMultiaddr))
	}
	if x.TransportId != 0 {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("transport_id: ")
		sb.WriteString(strconv.FormatUint(uint64(x.TransportId), 10))
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
				return fmt.Errorf("proto: wrong wireType = %d for field LocalPeerId", wireType)
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
			m.LocalPeerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RemotePeerId", wireType)
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
			m.RemotePeerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProtocolId", wireType)
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
			m.ProtocolId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ListenMultiaddr", wireType)
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
			m.ListenMultiaddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TransportId", wireType)
			}
			m.TransportId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TransportId |= uint64(b&0x7F) << shift
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
