// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.8.0
// source: github.com/aperturerobotics/bifrost/http/listener/config.proto

package bifrost_http_listener

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// Config configures a http server that listens on a port.
//
// Handles incoming requests with LookupHTTPHandler.
type Config struct {
	unknownFields []byte
	// Addr is the address to listen.
	//
	// Example: 0.0.0.0:8080
	Addr string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	// ClientId is the client id to set on LookupHTTPHandler.
	ClientId string `protobuf:"bytes,2,opt,name=client_id,json=clientId,proto3" json:"clientId,omitempty"`
	// CertFile is the path to the certificate file to use for https.
	// Can be unset to use HTTP.
	CertFile string `protobuf:"bytes,3,opt,name=cert_file,json=certFile,proto3" json:"certFile,omitempty"`
	// KeyFile is the path to the key file to use for https.
	// Cannot be unset if cert_file is set.
	// Otherwise can be unset.
	KeyFile string `protobuf:"bytes,4,opt,name=key_file,json=keyFile,proto3" json:"keyFile,omitempty"`
	// Wait indicates to wait for LookupHTTPHandler even if it becomes idle.
	// If false: returns 404 not found if LookupHTTPHandler becomes idle.
	Wait bool `protobuf:"varint,5,opt,name=wait,proto3" json:"wait,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
}

func (*Config) ProtoMessage() {}

func (x *Config) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *Config) GetClientId() string {
	if x != nil {
		return x.ClientId
	}
	return ""
}

func (x *Config) GetCertFile() string {
	if x != nil {
		return x.CertFile
	}
	return ""
}

func (x *Config) GetKeyFile() string {
	if x != nil {
		return x.KeyFile
	}
	return ""
}

func (x *Config) GetWait() bool {
	if x != nil {
		return x.Wait
	}
	return false
}

func (m *Config) CloneVT() *Config {
	if m == nil {
		return (*Config)(nil)
	}
	r := new(Config)
	r.Addr = m.Addr
	r.ClientId = m.ClientId
	r.CertFile = m.CertFile
	r.KeyFile = m.KeyFile
	r.Wait = m.Wait
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
	if this.Addr != that.Addr {
		return false
	}
	if this.ClientId != that.ClientId {
		return false
	}
	if this.CertFile != that.CertFile {
		return false
	}
	if this.KeyFile != that.KeyFile {
		return false
	}
	if this.Wait != that.Wait {
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
	if x.Addr != "" || s.HasField("addr") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("addr")
		s.WriteString(x.Addr)
	}
	if x.ClientId != "" || s.HasField("clientId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("clientId")
		s.WriteString(x.ClientId)
	}
	if x.CertFile != "" || s.HasField("certFile") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("certFile")
		s.WriteString(x.CertFile)
	}
	if x.KeyFile != "" || s.HasField("keyFile") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("keyFile")
		s.WriteString(x.KeyFile)
	}
	if x.Wait || s.HasField("wait") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("wait")
		s.WriteBool(x.Wait)
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
		case "addr":
			s.AddField("addr")
			x.Addr = s.ReadString()
		case "client_id", "clientId":
			s.AddField("client_id")
			x.ClientId = s.ReadString()
		case "cert_file", "certFile":
			s.AddField("cert_file")
			x.CertFile = s.ReadString()
		case "key_file", "keyFile":
			s.AddField("key_file")
			x.KeyFile = s.ReadString()
		case "wait":
			s.AddField("wait")
			x.Wait = s.ReadBool()
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
	if m.Wait {
		i--
		if m.Wait {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x28
	}
	if len(m.KeyFile) > 0 {
		i -= len(m.KeyFile)
		copy(dAtA[i:], m.KeyFile)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.KeyFile)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.CertFile) > 0 {
		i -= len(m.CertFile)
		copy(dAtA[i:], m.CertFile)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.CertFile)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ClientId) > 0 {
		i -= len(m.ClientId)
		copy(dAtA[i:], m.ClientId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ClientId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Addr) > 0 {
		i -= len(m.Addr)
		copy(dAtA[i:], m.Addr)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.Addr)))
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
	l = len(m.Addr)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.ClientId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.CertFile)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.KeyFile)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.Wait {
		n += 2
	}
	n += len(m.unknownFields)
	return n
}

func (x *Config) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Config {")
	if x.Addr != "" {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("addr: ")
		sb.WriteString(strconv.Quote(x.Addr))
	}
	if x.ClientId != "" {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("client_id: ")
		sb.WriteString(strconv.Quote(x.ClientId))
	}
	if x.CertFile != "" {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("cert_file: ")
		sb.WriteString(strconv.Quote(x.CertFile))
	}
	if x.KeyFile != "" {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("key_file: ")
		sb.WriteString(strconv.Quote(x.KeyFile))
	}
	if x.Wait != false {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("wait: ")
		sb.WriteString(strconv.FormatBool(x.Wait))
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
				return fmt.Errorf("proto: wrong wireType = %d for field Addr", wireType)
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
			m.Addr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClientId", wireType)
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
			m.ClientId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CertFile", wireType)
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
			m.CertFile = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field KeyFile", wireType)
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
			m.KeyFile = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Wait", wireType)
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
			m.Wait = bool(v != 0)
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
