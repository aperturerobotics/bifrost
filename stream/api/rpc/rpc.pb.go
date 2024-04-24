// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.6.0
// source: github.com/aperturerobotics/bifrost/stream/api/rpc/rpc.proto

package stream_api_rpc

import (
	base64 "encoding/base64"
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
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
	name, valid := StreamState_name[int32(x)]
	if valid {
		return name
	}
	return strconv.Itoa(int(x))
}

// Data is a data packet.
type Data struct {
	unknownFields []byte
	// State indicates stream state in-band.
	// Data is packet data from the remote.
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	// State indicates the stream state.
	State StreamState `protobuf:"varint,2,opt,name=state,proto3" json:"state,omitempty"`
}

func (x *Data) Reset() {
	*x = Data{}
}

func (*Data) ProtoMessage() {}

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

func (m *Data) CloneVT() *Data {
	if m == nil {
		return (*Data)(nil)
	}
	r := new(Data)
	r.State = m.State
	if rhs := m.Data; rhs != nil {
		tmpBytes := make([]byte, len(rhs))
		copy(tmpBytes, rhs)
		r.Data = tmpBytes
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Data) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *Data) EqualVT(that *Data) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if string(this.Data) != string(that.Data) {
		return false
	}
	if this.State != that.State {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Data) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*Data)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the StreamState to JSON.
func (x StreamState) MarshalProtoJSON(s *json.MarshalState) {
	s.WriteEnumString(int32(x), StreamState_name)
}

// MarshalText marshals the StreamState to text.
func (x StreamState) MarshalText() ([]byte, error) {
	return []byte(json.GetEnumString(int32(x), StreamState_name)), nil
}

// MarshalJSON marshals the StreamState to JSON.
func (x StreamState) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the StreamState from JSON.
func (x *StreamState) UnmarshalProtoJSON(s *json.UnmarshalState) {
	v := s.ReadEnum(StreamState_value)
	if err := s.Err(); err != nil {
		s.SetErrorf("could not read StreamState enum: %v", err)
		return
	}
	*x = StreamState(v)
}

// UnmarshalText unmarshals the StreamState from text.
func (x *StreamState) UnmarshalText(b []byte) error {
	i, err := json.ParseEnumString(string(b), StreamState_value)
	if err != nil {
		return err
	}
	*x = StreamState(i)
	return nil
}

// UnmarshalJSON unmarshals the StreamState from JSON.
func (x *StreamState) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the Data message to JSON.
func (x *Data) MarshalProtoJSON(s *json.MarshalState) {
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
	if x.State != 0 || s.HasField("state") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("state")
		x.State.MarshalProtoJSON(s)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the Data to JSON.
func (x *Data) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the Data message from JSON.
func (x *Data) UnmarshalProtoJSON(s *json.UnmarshalState) {
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
		case "state":
			s.AddField("state")
			x.State.UnmarshalProtoJSON(s)
		}
	})
}

// UnmarshalJSON unmarshals the Data from JSON.
func (x *Data) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *Data) MarshalVT() (dAtA []byte, err error) {
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

func (m *Data) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Data) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if m.State != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.State))
		i--
		dAtA[i] = 0x10
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

func (m *Data) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.State != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.State))
	}
	n += len(m.unknownFields)
	return n
}

func (x StreamState) MarshalProtoText() string {
	return x.String()
}
func (x *Data) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Data { ")
	if len(x.Data) > 0 {
		sb.WriteString(" data: ")
		sb.WriteString("\"")
		sb.WriteString(base64.StdEncoding.EncodeToString(x.Data))
		sb.WriteString("\"")
	}
	if x.State != 0 {
		sb.WriteString(" state: ")
		sb.WriteString(StreamState(x.State).String())
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *Data) String() string {
	return x.MarshalProtoText()
}
func (m *Data) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: Data: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Data: illegal tag %d (wire type %d)", fieldNum, wire)
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field State", wireType)
			}
			m.State = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.State |= StreamState(b&0x7F) << shift
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
