// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.8.0
// source: github.com/aperturerobotics/bifrost/peer/peer.proto

package peer

import (
	base64 "encoding/base64"
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	hash "github.com/aperturerobotics/bifrost/hash"
	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// Signature contains a signature by a peer.
type Signature struct {
	unknownFields []byte
	// PubKey is the public key of the peer.
	// May be empty if the public key is to be inferred from context.
	PubKey []byte `protobuf:"bytes,1,opt,name=pub_key,json=pubKey,proto3" json:"pubKey,omitempty"`
	// HashType is the hash type used to hash the data.
	// The signature is then of the hash bytes (usually 32).
	HashType hash.HashType `protobuf:"varint,2,opt,name=hash_type,json=hashType,proto3" json:"hashType,omitempty"`
	// SigData contains the signature data.
	// The format is defined by the key type.
	SigData []byte `protobuf:"bytes,3,opt,name=sig_data,json=sigData,proto3" json:"sigData,omitempty"`
}

func (x *Signature) Reset() {
	*x = Signature{}
}

func (*Signature) ProtoMessage() {}

func (x *Signature) GetPubKey() []byte {
	if x != nil {
		return x.PubKey
	}
	return nil
}

func (x *Signature) GetHashType() hash.HashType {
	if x != nil {
		return x.HashType
	}
	return hash.HashType(0)
}

func (x *Signature) GetSigData() []byte {
	if x != nil {
		return x.SigData
	}
	return nil
}

// SignedMsg is a message from a peer with a signature.
type SignedMsg struct {
	unknownFields []byte
	// FromPeerId is the peer identifier of the sender.
	FromPeerId string `protobuf:"bytes,1,opt,name=from_peer_id,json=fromPeerId,proto3" json:"fromPeerId,omitempty"`
	// Signature is the sender signature.
	// Should not contain PubKey, which is inferred from peer id.
	Signature *Signature `protobuf:"bytes,2,opt,name=signature,proto3" json:"signature,omitempty"`
	// Data is the signed data.
	Data []byte `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *SignedMsg) Reset() {
	*x = SignedMsg{}
}

func (*SignedMsg) ProtoMessage() {}

func (x *SignedMsg) GetFromPeerId() string {
	if x != nil {
		return x.FromPeerId
	}
	return ""
}

func (x *SignedMsg) GetSignature() *Signature {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *SignedMsg) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (m *Signature) CloneVT() *Signature {
	if m == nil {
		return (*Signature)(nil)
	}
	r := new(Signature)
	r.HashType = m.HashType
	if rhs := m.PubKey; rhs != nil {
		tmpBytes := make([]byte, len(rhs))
		copy(tmpBytes, rhs)
		r.PubKey = tmpBytes
	}
	if rhs := m.SigData; rhs != nil {
		tmpBytes := make([]byte, len(rhs))
		copy(tmpBytes, rhs)
		r.SigData = tmpBytes
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Signature) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *SignedMsg) CloneVT() *SignedMsg {
	if m == nil {
		return (*SignedMsg)(nil)
	}
	r := new(SignedMsg)
	r.FromPeerId = m.FromPeerId
	r.Signature = m.Signature.CloneVT()
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

func (m *SignedMsg) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *Signature) EqualVT(that *Signature) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if string(this.PubKey) != string(that.PubKey) {
		return false
	}
	if this.HashType != that.HashType {
		return false
	}
	if string(this.SigData) != string(that.SigData) {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Signature) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*Signature)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *SignedMsg) EqualVT(that *SignedMsg) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.FromPeerId != that.FromPeerId {
		return false
	}
	if !this.Signature.EqualVT(that.Signature) {
		return false
	}
	if string(this.Data) != string(that.Data) {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *SignedMsg) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*SignedMsg)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the Signature message to JSON.
func (x *Signature) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if len(x.PubKey) > 0 || s.HasField("pubKey") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("pubKey")
		s.WriteBytes(x.PubKey)
	}
	if x.HashType != 0 || s.HasField("hashType") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("hashType")
		x.HashType.MarshalProtoJSON(s)
	}
	if len(x.SigData) > 0 || s.HasField("sigData") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("sigData")
		s.WriteBytes(x.SigData)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the Signature to JSON.
func (x *Signature) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the Signature message from JSON.
func (x *Signature) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "pub_key", "pubKey":
			s.AddField("pub_key")
			x.PubKey = s.ReadBytes()
		case "hash_type", "hashType":
			s.AddField("hash_type")
			x.HashType.UnmarshalProtoJSON(s)
		case "sig_data", "sigData":
			s.AddField("sig_data")
			x.SigData = s.ReadBytes()
		}
	})
}

// UnmarshalJSON unmarshals the Signature from JSON.
func (x *Signature) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the SignedMsg message to JSON.
func (x *SignedMsg) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.FromPeerId != "" || s.HasField("fromPeerId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("fromPeerId")
		s.WriteString(x.FromPeerId)
	}
	if x.Signature != nil || s.HasField("signature") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("signature")
		x.Signature.MarshalProtoJSON(s.WithField("signature"))
	}
	if len(x.Data) > 0 || s.HasField("data") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("data")
		s.WriteBytes(x.Data)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the SignedMsg to JSON.
func (x *SignedMsg) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the SignedMsg message from JSON.
func (x *SignedMsg) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "from_peer_id", "fromPeerId":
			s.AddField("from_peer_id")
			x.FromPeerId = s.ReadString()
		case "signature":
			if s.ReadNil() {
				x.Signature = nil
				return
			}
			x.Signature = &Signature{}
			x.Signature.UnmarshalProtoJSON(s.WithField("signature", true))
		case "data":
			s.AddField("data")
			x.Data = s.ReadBytes()
		}
	})
}

// UnmarshalJSON unmarshals the SignedMsg from JSON.
func (x *SignedMsg) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *Signature) MarshalVT() (dAtA []byte, err error) {
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

func (m *Signature) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Signature) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if len(m.SigData) > 0 {
		i -= len(m.SigData)
		copy(dAtA[i:], m.SigData)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.SigData)))
		i--
		dAtA[i] = 0x1a
	}
	if m.HashType != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.HashType))
		i--
		dAtA[i] = 0x10
	}
	if len(m.PubKey) > 0 {
		i -= len(m.PubKey)
		copy(dAtA[i:], m.PubKey)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.PubKey)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SignedMsg) MarshalVT() (dAtA []byte, err error) {
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

func (m *SignedMsg) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *SignedMsg) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0x1a
	}
	if m.Signature != nil {
		size, err := m.Signature.MarshalToSizedBufferVT(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
		i--
		dAtA[i] = 0x12
	}
	if len(m.FromPeerId) > 0 {
		i -= len(m.FromPeerId)
		copy(dAtA[i:], m.FromPeerId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.FromPeerId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Signature) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PubKey)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.HashType != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.HashType))
	}
	l = len(m.SigData)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (m *SignedMsg) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.FromPeerId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.Signature != nil {
		l = m.Signature.SizeVT()
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (x *Signature) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Signature {")
	if x.PubKey != nil {
		if sb.Len() > 11 {
			sb.WriteString(" ")
		}
		sb.WriteString("pub_key: ")
		sb.WriteString("\"")
		sb.WriteString(base64.StdEncoding.EncodeToString(x.PubKey))
		sb.WriteString("\"")
	}
	if x.HashType != 0 {
		if sb.Len() > 11 {
			sb.WriteString(" ")
		}
		sb.WriteString("hash_type: ")
		sb.WriteString("\"")
		sb.WriteString(hash.HashType(x.HashType).String())
		sb.WriteString("\"")
	}
	if x.SigData != nil {
		if sb.Len() > 11 {
			sb.WriteString(" ")
		}
		sb.WriteString("sig_data: ")
		sb.WriteString("\"")
		sb.WriteString(base64.StdEncoding.EncodeToString(x.SigData))
		sb.WriteString("\"")
	}
	sb.WriteString("}")
	return sb.String()
}

func (x *Signature) String() string {
	return x.MarshalProtoText()
}
func (x *SignedMsg) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("SignedMsg {")
	if x.FromPeerId != "" {
		if sb.Len() > 11 {
			sb.WriteString(" ")
		}
		sb.WriteString("from_peer_id: ")
		sb.WriteString(strconv.Quote(x.FromPeerId))
	}
	if x.Signature != nil {
		if sb.Len() > 11 {
			sb.WriteString(" ")
		}
		sb.WriteString("signature: ")
		sb.WriteString(x.Signature.MarshalProtoText())
	}
	if x.Data != nil {
		if sb.Len() > 11 {
			sb.WriteString(" ")
		}
		sb.WriteString("data: ")
		sb.WriteString("\"")
		sb.WriteString(base64.StdEncoding.EncodeToString(x.Data))
		sb.WriteString("\"")
	}
	sb.WriteString("}")
	return sb.String()
}

func (x *SignedMsg) String() string {
	return x.MarshalProtoText()
}
func (m *Signature) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: Signature: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Signature: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PubKey", wireType)
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
			m.PubKey = append(m.PubKey[:0], dAtA[iNdEx:postIndex]...)
			if m.PubKey == nil {
				m.PubKey = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field HashType", wireType)
			}
			m.HashType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.HashType |= hash.HashType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SigData", wireType)
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
			m.SigData = append(m.SigData[:0], dAtA[iNdEx:postIndex]...)
			if m.SigData == nil {
				m.SigData = []byte{}
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
func (m *SignedMsg) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: SignedMsg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SignedMsg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FromPeerId", wireType)
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
			m.FromPeerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signature", wireType)
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
			if m.Signature == nil {
				m.Signature = &Signature{}
			}
			if err := m.Signature.UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
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
