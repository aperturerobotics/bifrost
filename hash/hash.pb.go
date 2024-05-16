// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.6.5
// source: github.com/aperturerobotics/bifrost/hash/hash.proto

package hash

import (
	base64 "encoding/base64"
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// HashType identifies the hash type in use.
type HashType int32

const (
	// HashType_UNKNOWN is an unknown hash type.
	HashType_HashType_UNKNOWN HashType = 0
	// HashType_SHA256 is the sha256 hash type.
	HashType_HashType_SHA256 HashType = 1
	// HashType_SHA1 is the sha1 hash type.
	// NOTE: Do not use SHA1 unless you absolutely have to for backwards compat! (Git)
	HashType_HashType_SHA1 HashType = 2
	// HashType_BLAKE3 is the blake3 hash type.
	// Uses a 32-byte digest size.
	HashType_HashType_BLAKE3 HashType = 3
)

// Enum value maps for HashType.
var (
	HashType_name = map[int32]string{
		0: "HashType_UNKNOWN",
		1: "HashType_SHA256",
		2: "HashType_SHA1",
		3: "HashType_BLAKE3",
	}
	HashType_value = map[string]int32{
		"HashType_UNKNOWN": 0,
		"HashType_SHA256":  1,
		"HashType_SHA1":    2,
		"HashType_BLAKE3":  3,
	}
)

func (x HashType) Enum() *HashType {
	p := new(HashType)
	*p = x
	return p
}

func (x HashType) String() string {
	name, valid := HashType_name[int32(x)]
	if valid {
		return name
	}
	return strconv.Itoa(int(x))
}

// Hash is a hash of a binary blob.
type Hash struct {
	unknownFields []byte
	// HashType is the hash type in use.
	HashType HashType `protobuf:"varint,1,opt,name=hash_type,json=hashType,proto3" json:"hashType,omitempty"`
	// Hash is the hash value.
	Hash []byte `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *Hash) Reset() {
	*x = Hash{}
}

func (*Hash) ProtoMessage() {}

func (x *Hash) GetHashType() HashType {
	if x != nil {
		return x.HashType
	}
	return HashType_HashType_UNKNOWN
}

func (x *Hash) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (m *Hash) CloneVT() *Hash {
	if m == nil {
		return (*Hash)(nil)
	}
	r := new(Hash)
	r.HashType = m.HashType
	if rhs := m.Hash; rhs != nil {
		tmpBytes := make([]byte, len(rhs))
		copy(tmpBytes, rhs)
		r.Hash = tmpBytes
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Hash) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *Hash) EqualVT(that *Hash) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.HashType != that.HashType {
		return false
	}
	if string(this.Hash) != string(that.Hash) {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Hash) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*Hash)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the HashType to JSON.
func (x HashType) MarshalProtoJSON(s *json.MarshalState) {
	s.WriteEnumString(int32(x), HashType_name)
}

// MarshalText marshals the HashType to text.
func (x HashType) MarshalText() ([]byte, error) {
	return []byte(json.GetEnumString(int32(x), HashType_name)), nil
}

// MarshalJSON marshals the HashType to JSON.
func (x HashType) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the HashType from JSON.
func (x *HashType) UnmarshalProtoJSON(s *json.UnmarshalState) {
	v := s.ReadEnum(HashType_value)
	if err := s.Err(); err != nil {
		s.SetErrorf("could not read HashType enum: %v", err)
		return
	}
	*x = HashType(v)
}

// UnmarshalText unmarshals the HashType from text.
func (x *HashType) UnmarshalText(b []byte) error {
	i, err := json.ParseEnumString(string(b), HashType_value)
	if err != nil {
		return err
	}
	*x = HashType(i)
	return nil
}

// UnmarshalJSON unmarshals the HashType from JSON.
func (x *HashType) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the Hash message to JSON.
func (x *Hash) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.HashType != 0 || s.HasField("hashType") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("hashType")
		x.HashType.MarshalProtoJSON(s)
	}
	if len(x.Hash) > 0 || s.HasField("hash") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("hash")
		s.WriteBytes(x.Hash)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the Hash to JSON.
func (x *Hash) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the Hash message from JSON.
func (x *Hash) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "hash_type", "hashType":
			s.AddField("hash_type")
			x.HashType.UnmarshalProtoJSON(s)
		case "hash":
			s.AddField("hash")
			x.Hash = s.ReadBytes()
		}
	})
}

// UnmarshalJSON unmarshals the Hash from JSON.
func (x *Hash) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *Hash) MarshalVT() (dAtA []byte, err error) {
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

func (m *Hash) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Hash) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if len(m.Hash) > 0 {
		i -= len(m.Hash)
		copy(dAtA[i:], m.Hash)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.Hash)))
		i--
		dAtA[i] = 0x12
	}
	if m.HashType != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.HashType))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Hash) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.HashType != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.HashType))
	}
	l = len(m.Hash)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (x HashType) MarshalProtoText() string {
	return x.String()
}
func (x *Hash) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Hash { ")
	if x.HashType != 0 {
		sb.WriteString(" hash_type: ")
		sb.WriteString(HashType(x.HashType).String())
	}
	if len(x.Hash) > 0 {
		sb.WriteString(" hash: ")
		sb.WriteString("\"")
		sb.WriteString(base64.StdEncoding.EncodeToString(x.Hash))
		sb.WriteString("\"")
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *Hash) String() string {
	return x.MarshalProtoText()
}
func (m *Hash) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: Hash: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Hash: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
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
				m.HashType |= HashType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hash", wireType)
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
			m.Hash = append(m.Hash[:0], dAtA[iNdEx:postIndex]...)
			if m.Hash == nil {
				m.Hash = []byte{}
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
