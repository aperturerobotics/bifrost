// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.6.1
// source: github.com/aperturerobotics/bifrost/pubsub/util/pubmessage/pubmessage.proto

package pubmessage

import (
	base64 "encoding/base64"
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
	timestamppb "github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
)

// PubMessageInner is the signed inner portion of the message.
type PubMessageInner struct {
	unknownFields []byte
	// Data is the message data.
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	// Channel is the channel.
	Channel string `protobuf:"bytes,2,opt,name=channel,proto3" json:"channel,omitempty"`
	// Timestamp is the message timestamp.
	Timestamp *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (x *PubMessageInner) Reset() {
	*x = PubMessageInner{}
}

func (*PubMessageInner) ProtoMessage() {}

func (x *PubMessageInner) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *PubMessageInner) GetChannel() string {
	if x != nil {
		return x.Channel
	}
	return ""
}

func (x *PubMessageInner) GetTimestamp() *timestamppb.Timestamp {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (m *PubMessageInner) CloneVT() *PubMessageInner {
	if m == nil {
		return (*PubMessageInner)(nil)
	}
	r := new(PubMessageInner)
	r.Channel = m.Channel
	if rhs := m.Data; rhs != nil {
		tmpBytes := make([]byte, len(rhs))
		copy(tmpBytes, rhs)
		r.Data = tmpBytes
	}
	if rhs := m.Timestamp; rhs != nil {
		r.Timestamp = rhs.CloneVT()
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *PubMessageInner) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *PubMessageInner) EqualVT(that *PubMessageInner) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if string(this.Data) != string(that.Data) {
		return false
	}
	if this.Channel != that.Channel {
		return false
	}
	if !this.Timestamp.EqualVT(that.Timestamp) {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *PubMessageInner) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*PubMessageInner)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the PubMessageInner message to JSON.
func (x *PubMessageInner) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if len(x.Data) > 0 || s.HasField("data") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("data")
		s.WriteBytes(x.Data)
	}
	if x.Channel != "" || s.HasField("channel") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("channel")
		s.WriteString(x.Channel)
	}
	if x.Timestamp != nil || s.HasField("timestamp") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("timestamp")
		x.Timestamp.MarshalProtoJSON(s.WithField("timestamp"))
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the PubMessageInner to JSON.
func (x *PubMessageInner) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the PubMessageInner message from JSON.
func (x *PubMessageInner) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "data":
			s.AddField("data")
			x.Data = s.ReadBytes()
		case "channel":
			s.AddField("channel")
			x.Channel = s.ReadString()
		case "timestamp":
			if s.ReadNil() {
				x.Timestamp = nil
				return
			}
			x.Timestamp = &timestamppb.Timestamp{}
			x.Timestamp.UnmarshalProtoJSON(s.WithField("timestamp", true))
		}
	})
}

// UnmarshalJSON unmarshals the PubMessageInner from JSON.
func (x *PubMessageInner) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *PubMessageInner) MarshalVT() (dAtA []byte, err error) {
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

func (m *PubMessageInner) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *PubMessageInner) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if m.Timestamp != nil {
		size, err := m.Timestamp.MarshalToSizedBufferVT(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Channel) > 0 {
		i -= len(m.Channel)
		copy(dAtA[i:], m.Channel)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.Channel)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *PubMessageInner) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.Channel)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.Timestamp != nil {
		l = m.Timestamp.SizeVT()
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (x *PubMessageInner) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("PubMessageInner { ")
	if len(x.Data) > 0 {
		sb.WriteString(" data: ")
		sb.WriteString("\"")
		sb.WriteString(base64.StdEncoding.EncodeToString(x.Data))
		sb.WriteString("\"")
	}
	if x.Channel != "" {
		sb.WriteString(" channel: ")
		sb.WriteString(strconv.Quote(x.Channel))
	}
	if x.Timestamp != nil {
		sb.WriteString(" timestamp: ")
		sb.WriteString(x.Timestamp.MarshalProtoText())
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *PubMessageInner) String() string {
	return x.MarshalProtoText()
}
func (m *PubMessageInner) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: PubMessageInner: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PubMessageInner: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return protobuf_go_lite.ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Channel", wireType)
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
			m.Channel = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timestamp", wireType)
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
			if m.Timestamp == nil {
				m.Timestamp = &timestamppb.Timestamp{}
			}
			if err := m.Timestamp.UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
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
