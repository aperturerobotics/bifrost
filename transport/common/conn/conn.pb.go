// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.6.1
// source: github.com/aperturerobotics/bifrost/transport/common/conn/conn.proto

package conn

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	quic "github.com/aperturerobotics/bifrost/transport/common/quic"
	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// Opts are extra options for the reliable conn.
type Opts struct {
	unknownFields []byte
	// Quic are the quic protocol options.
	Quic *quic.Opts `protobuf:"bytes,1,opt,name=quic,proto3" json:"quic,omitempty"`
	// Verbose turns on verbose debug logging.
	Verbose bool `protobuf:"varint,2,opt,name=verbose,proto3" json:"verbose,omitempty"`
	// Mtu sets the maximum size for a single packet.
	// Defaults to 65000.
	Mtu uint32 `protobuf:"varint,3,opt,name=mtu,proto3" json:"mtu,omitempty"`
	// BufSize is the number of packets to buffer.
	//
	// Total memory cap is mtu * bufSize.
	// Defaults to 10.
	BufSize uint32 `protobuf:"varint,4,opt,name=buf_size,json=bufSize,proto3" json:"bufSize,omitempty"`
}

func (x *Opts) Reset() {
	*x = Opts{}
}

func (*Opts) ProtoMessage() {}

func (x *Opts) GetQuic() *quic.Opts {
	if x != nil {
		return x.Quic
	}
	return nil
}

func (x *Opts) GetVerbose() bool {
	if x != nil {
		return x.Verbose
	}
	return false
}

func (x *Opts) GetMtu() uint32 {
	if x != nil {
		return x.Mtu
	}
	return 0
}

func (x *Opts) GetBufSize() uint32 {
	if x != nil {
		return x.BufSize
	}
	return 0
}

func (m *Opts) CloneVT() *Opts {
	if m == nil {
		return (*Opts)(nil)
	}
	r := new(Opts)
	r.Verbose = m.Verbose
	r.Mtu = m.Mtu
	r.BufSize = m.BufSize
	if rhs := m.Quic; rhs != nil {
		r.Quic = rhs.CloneVT()
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Opts) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *Opts) EqualVT(that *Opts) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if !this.Quic.EqualVT(that.Quic) {
		return false
	}
	if this.Verbose != that.Verbose {
		return false
	}
	if this.Mtu != that.Mtu {
		return false
	}
	if this.BufSize != that.BufSize {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Opts) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*Opts)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the Opts message to JSON.
func (x *Opts) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Quic != nil || s.HasField("quic") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("quic")
		x.Quic.MarshalProtoJSON(s.WithField("quic"))
	}
	if x.Verbose || s.HasField("verbose") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("verbose")
		s.WriteBool(x.Verbose)
	}
	if x.Mtu != 0 || s.HasField("mtu") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("mtu")
		s.WriteUint32(x.Mtu)
	}
	if x.BufSize != 0 || s.HasField("bufSize") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("bufSize")
		s.WriteUint32(x.BufSize)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the Opts to JSON.
func (x *Opts) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the Opts message from JSON.
func (x *Opts) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "quic":
			if s.ReadNil() {
				x.Quic = nil
				return
			}
			x.Quic = &quic.Opts{}
			x.Quic.UnmarshalProtoJSON(s.WithField("quic", true))
		case "verbose":
			s.AddField("verbose")
			x.Verbose = s.ReadBool()
		case "mtu":
			s.AddField("mtu")
			x.Mtu = s.ReadUint32()
		case "buf_size", "bufSize":
			s.AddField("buf_size")
			x.BufSize = s.ReadUint32()
		}
	})
}

// UnmarshalJSON unmarshals the Opts from JSON.
func (x *Opts) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *Opts) MarshalVT() (dAtA []byte, err error) {
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

func (m *Opts) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Opts) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if m.BufSize != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.BufSize))
		i--
		dAtA[i] = 0x20
	}
	if m.Mtu != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.Mtu))
		i--
		dAtA[i] = 0x18
	}
	if m.Verbose {
		i--
		if m.Verbose {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x10
	}
	if m.Quic != nil {
		size, err := m.Quic.MarshalToSizedBufferVT(dAtA[:i])
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

func (m *Opts) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Quic != nil {
		l = m.Quic.SizeVT()
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.Verbose {
		n += 2
	}
	if m.Mtu != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.Mtu))
	}
	if m.BufSize != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.BufSize))
	}
	n += len(m.unknownFields)
	return n
}

func (x *Opts) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Opts { ")
	if x.Quic != nil {
		sb.WriteString(" quic: ")
		sb.WriteString(x.Quic.MarshalProtoText())
	}
	if x.Verbose {
		sb.WriteString(" verbose: ")
		sb.WriteString(strconv.FormatBool(x.Verbose))
	}
	if x.Mtu != 0 {
		sb.WriteString(" mtu: ")
		sb.WriteString(strconv.FormatUint(uint64(x.Mtu), 10))
	}
	if x.BufSize != 0 {
		sb.WriteString(" buf_size: ")
		sb.WriteString(strconv.FormatUint(uint64(x.BufSize), 10))
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *Opts) String() string {
	return x.MarshalProtoText()
}
func (m *Opts) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: Opts: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Opts: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Quic", wireType)
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
			if m.Quic == nil {
				m.Quic = &quic.Opts{}
			}
			if err := m.Quic.UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Verbose", wireType)
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
			m.Verbose = bool(v != 0)
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Mtu", wireType)
			}
			m.Mtu = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Mtu |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BufSize", wireType)
			}
			m.BufSize = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BufSize |= uint32(b&0x7F) << shift
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
