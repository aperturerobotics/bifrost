// Code generated by protoc-gen-go-vtproto. DO NOT EDIT.
// protoc-gen-go-vtproto version: v0.3.1-0.20220817155510-0ae748fd2007
// source: github.com/aperturerobotics/bifrost/stream/drpc/client/client.proto

package stream_drpc_client

import (
	fmt "fmt"
	io "io"
	bits "math/bits"

	drpc "github.com/aperturerobotics/bifrost/stream/drpc"
	backoff "github.com/aperturerobotics/util/backoff"
	proto "google.golang.org/protobuf/proto"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

func (m *Config) CloneVT() *Config {
	if m == nil {
		return (*Config)(nil)
	}
	r := &Config{
		SrcPeerId:   m.SrcPeerId,
		TransportId: m.TransportId,
		TimeoutDur:  m.TimeoutDur,
	}
	if rhs := m.ServerPeerIds; rhs != nil {
		tmpContainer := make([]string, len(rhs))
		copy(tmpContainer, rhs)
		r.ServerPeerIds = tmpContainer
	}
	if rhs := m.PerServerBackoff; rhs != nil {
		if vtpb, ok := interface{}(rhs).(interface{ CloneVT() *backoff.Backoff }); ok {
			r.PerServerBackoff = vtpb.CloneVT()
		} else {
			r.PerServerBackoff = proto.Clone(rhs).(*backoff.Backoff)
		}
	}
	if rhs := m.DrpcOpts; rhs != nil {
		if vtpb, ok := interface{}(rhs).(interface{ CloneVT() *drpc.DrpcOpts }); ok {
			r.DrpcOpts = vtpb.CloneVT()
		} else {
			r.DrpcOpts = proto.Clone(rhs).(*drpc.DrpcOpts)
		}
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Config) CloneGenericVT() proto.Message {
	return m.CloneVT()
}

func (this *Config) EqualVT(that *Config) bool {
	if this == nil {
		return that == nil
	} else if that == nil {
		return false
	}
	if len(this.ServerPeerIds) != len(that.ServerPeerIds) {
		return false
	}
	for i, vx := range this.ServerPeerIds {
		vy := that.ServerPeerIds[i]
		if vx != vy {
			return false
		}
	}
	if equal, ok := interface{}(this.PerServerBackoff).(interface{ EqualVT(*backoff.Backoff) bool }); ok {
		if !equal.EqualVT(that.PerServerBackoff) {
			return false
		}
	} else if !proto.Equal(this.PerServerBackoff, that.PerServerBackoff) {
		return false
	}
	if this.SrcPeerId != that.SrcPeerId {
		return false
	}
	if this.TransportId != that.TransportId {
		return false
	}
	if equal, ok := interface{}(this.DrpcOpts).(interface{ EqualVT(*drpc.DrpcOpts) bool }); ok {
		if !equal.EqualVT(that.DrpcOpts) {
			return false
		}
	} else if !proto.Equal(this.DrpcOpts, that.DrpcOpts) {
		return false
	}
	if this.TimeoutDur != that.TimeoutDur {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
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
	if len(m.TimeoutDur) > 0 {
		i -= len(m.TimeoutDur)
		copy(dAtA[i:], m.TimeoutDur)
		i = encodeVarint(dAtA, i, uint64(len(m.TimeoutDur)))
		i--
		dAtA[i] = 0x32
	}
	if m.DrpcOpts != nil {
		if vtmsg, ok := interface{}(m.DrpcOpts).(interface {
			MarshalToSizedBufferVT([]byte) (int, error)
		}); ok {
			size, err := vtmsg.MarshalToSizedBufferVT(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarint(dAtA, i, uint64(size))
		} else {
			encoded, err := proto.Marshal(m.DrpcOpts)
			if err != nil {
				return 0, err
			}
			i -= len(encoded)
			copy(dAtA[i:], encoded)
			i = encodeVarint(dAtA, i, uint64(len(encoded)))
		}
		i--
		dAtA[i] = 0x2a
	}
	if m.TransportId != 0 {
		i = encodeVarint(dAtA, i, uint64(m.TransportId))
		i--
		dAtA[i] = 0x20
	}
	if len(m.SrcPeerId) > 0 {
		i -= len(m.SrcPeerId)
		copy(dAtA[i:], m.SrcPeerId)
		i = encodeVarint(dAtA, i, uint64(len(m.SrcPeerId)))
		i--
		dAtA[i] = 0x1a
	}
	if m.PerServerBackoff != nil {
		if vtmsg, ok := interface{}(m.PerServerBackoff).(interface {
			MarshalToSizedBufferVT([]byte) (int, error)
		}); ok {
			size, err := vtmsg.MarshalToSizedBufferVT(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarint(dAtA, i, uint64(size))
		} else {
			encoded, err := proto.Marshal(m.PerServerBackoff)
			if err != nil {
				return 0, err
			}
			i -= len(encoded)
			copy(dAtA[i:], encoded)
			i = encodeVarint(dAtA, i, uint64(len(encoded)))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.ServerPeerIds) > 0 {
		for iNdEx := len(m.ServerPeerIds) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.ServerPeerIds[iNdEx])
			copy(dAtA[i:], m.ServerPeerIds[iNdEx])
			i = encodeVarint(dAtA, i, uint64(len(m.ServerPeerIds[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarint(dAtA []byte, offset int, v uint64) int {
	offset -= sov(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Config) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ServerPeerIds) > 0 {
		for _, s := range m.ServerPeerIds {
			l = len(s)
			n += 1 + l + sov(uint64(l))
		}
	}
	if m.PerServerBackoff != nil {
		if size, ok := interface{}(m.PerServerBackoff).(interface {
			SizeVT() int
		}); ok {
			l = size.SizeVT()
		} else {
			l = proto.Size(m.PerServerBackoff)
		}
		n += 1 + l + sov(uint64(l))
	}
	l = len(m.SrcPeerId)
	if l > 0 {
		n += 1 + l + sov(uint64(l))
	}
	if m.TransportId != 0 {
		n += 1 + sov(uint64(m.TransportId))
	}
	if m.DrpcOpts != nil {
		if size, ok := interface{}(m.DrpcOpts).(interface {
			SizeVT() int
		}); ok {
			l = size.SizeVT()
		} else {
			l = proto.Size(m.DrpcOpts)
		}
		n += 1 + l + sov(uint64(l))
	}
	l = len(m.TimeoutDur)
	if l > 0 {
		n += 1 + l + sov(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func sov(x uint64) (n int) {
	return (bits.Len64(x|1) + 6) / 7
}
func soz(x uint64) (n int) {
	return sov(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Config) UnmarshalVT(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflow
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
				return fmt.Errorf("proto: wrong wireType = %d for field ServerPeerIds", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflow
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
				return ErrInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ServerPeerIds = append(m.ServerPeerIds, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerServerBackoff", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflow
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
				return ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.PerServerBackoff == nil {
				m.PerServerBackoff = &backoff.Backoff{}
			}
			if unmarshal, ok := interface{}(m.PerServerBackoff).(interface {
				UnmarshalVT([]byte) error
			}); ok {
				if err := unmarshal.UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
					return err
				}
			} else {
				if err := proto.Unmarshal(dAtA[iNdEx:postIndex], m.PerServerBackoff); err != nil {
					return err
				}
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SrcPeerId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflow
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
				return ErrInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SrcPeerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TransportId", wireType)
			}
			m.TransportId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflow
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
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DrpcOpts", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflow
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
				return ErrInvalidLength
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.DrpcOpts == nil {
				m.DrpcOpts = &drpc.DrpcOpts{}
			}
			if unmarshal, ok := interface{}(m.DrpcOpts).(interface {
				UnmarshalVT([]byte) error
			}); ok {
				if err := unmarshal.UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
					return err
				}
			} else {
				if err := proto.Unmarshal(dAtA[iNdEx:postIndex], m.DrpcOpts); err != nil {
					return err
				}
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeoutDur", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflow
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
				return ErrInvalidLength
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLength
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TimeoutDur = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skip(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLength
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
func skip(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflow
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflow
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflow
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLength
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroup
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLength
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLength        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflow          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroup = fmt.Errorf("proto: unexpected end of group")
)
