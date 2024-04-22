// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.5.0
// source: github.com/aperturerobotics/bifrost/transport/controller/controller.proto

package transport_controller

import (
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
	errors "github.com/pkg/errors"
)

// StreamEstablish is the first message sent by the initiator of a stream.
// Prefixed by a uint32 length.
// Max size: 100kb
type StreamEstablish struct {
	unknownFields []byte
	// ProtocolID is the protocol identifier string for the stream.
	ProtocolId string `protobuf:"bytes,1,opt,name=protocol_id,json=protocolId,proto3" json:"protocolId,omitempty"`
}

func (x *StreamEstablish) Reset() {
	*x = StreamEstablish{}
}

func (*StreamEstablish) ProtoMessage() {}

func (x *StreamEstablish) GetProtocolId() string {
	if x != nil {
		return x.ProtocolId
	}
	return ""
}

func (m *StreamEstablish) CloneVT() *StreamEstablish {
	if m == nil {
		return (*StreamEstablish)(nil)
	}
	r := new(StreamEstablish)
	r.ProtocolId = m.ProtocolId
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *StreamEstablish) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *StreamEstablish) EqualVT(that *StreamEstablish) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.ProtocolId != that.ProtocolId {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *StreamEstablish) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*StreamEstablish)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the StreamEstablish message to JSON.
func (x *StreamEstablish) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.ProtocolId != "" || s.HasField("protocolId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("protocolId")
		s.WriteString(x.ProtocolId)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the StreamEstablish to JSON.
func (x *StreamEstablish) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the StreamEstablish message from JSON.
func (x *StreamEstablish) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "protocol_id", "protocolId":
			s.AddField("protocol_id")
			x.ProtocolId = s.ReadString()
		}
	})
}

// UnmarshalJSON unmarshals the StreamEstablish from JSON.
func (x *StreamEstablish) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *StreamEstablish) MarshalVT() (dAtA []byte, err error) {
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

func (m *StreamEstablish) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *StreamEstablish) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if len(m.ProtocolId) > 0 {
		i -= len(m.ProtocolId)
		copy(dAtA[i:], m.ProtocolId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ProtocolId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *StreamEstablish) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ProtocolId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (x *StreamEstablish) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("StreamEstablish { ")
	if x.ProtocolId != "" {
		sb.WriteString(" protocol_id: ")
		sb.WriteString(strconv.Quote(x.ProtocolId))
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *StreamEstablish) String() string {
	return x.MarshalProtoText()
}
func (m *StreamEstablish) UnmarshalVT(dAtA []byte) error {
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
			return errors.Errorf("proto: StreamEstablish: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return errors.Errorf("proto: StreamEstablish: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return errors.Errorf("proto: wrong wireType = %d for field ProtocolId", wireType)
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
