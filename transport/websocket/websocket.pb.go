// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.6.0
// source: github.com/aperturerobotics/bifrost/transport/websocket/websocket.proto

package websocket

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	dialer "github.com/aperturerobotics/bifrost/transport/common/dialer"
	quic "github.com/aperturerobotics/bifrost/transport/common/quic"
	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// Config is the configuration for the Websocket transport.
//
// Bifrost speaks Quic over the websocket. While this is not always necessary,
// especially when using wss transports, we still need to ensure end-to-end
// encryption to the peer that we handshake with on the other end, and to manage
// stream congestion control, multiplexing,
type Config struct {
	unknownFields []byte
	// TransportPeerID sets the peer ID to attach the transport to.
	// If unset, attaches to any running peer with a private key.
	TransportPeerId string `protobuf:"bytes,1,opt,name=transport_peer_id,json=transportPeerId,proto3" json:"transportPeerId,omitempty"`
	// ListenAddr contains the address to listen on.
	// Has no effect in the browser.
	ListenAddr string `protobuf:"bytes,2,opt,name=listen_addr,json=listenAddr,proto3" json:"listenAddr,omitempty"`
	// Quic contains the quic protocol options.
	//
	// The WebSocket transport always disables FEC and several other UDP-centric
	// features which are unnecessary due to the "reliable" nature of WebSockets.
	Quic *quic.Opts `protobuf:"bytes,3,opt,name=quic,proto3" json:"quic,omitempty"`
	// Dialers maps peer IDs to dialers.
	Dialers map[string]*dialer.DialerOpts `protobuf:"bytes,4,rep,name=dialers,proto3" json:"dialers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// HttpPath is the http path to expose the websocket.
	// If unset, ignores the incoming request path.
	HttpPath string `protobuf:"bytes,5,opt,name=http_path,json=httpPath,proto3" json:"httpPath,omitempty"`
	// DisableServePeerId disables serving the peer id.
	// If this is unset the peer ID is available at http_path+"/peer"
	// If http_path is unset the peer ID is available at /peer
	DisableServePeerId bool `protobuf:"varint,6,opt,name=disable_serve_peer_id,json=disableServePeerId,proto3" json:"disableServePeerId,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
}

func (*Config) ProtoMessage() {}

func (x *Config) GetTransportPeerId() string {
	if x != nil {
		return x.TransportPeerId
	}
	return ""
}

func (x *Config) GetListenAddr() string {
	if x != nil {
		return x.ListenAddr
	}
	return ""
}

func (x *Config) GetQuic() *quic.Opts {
	if x != nil {
		return x.Quic
	}
	return nil
}

func (x *Config) GetDialers() map[string]*dialer.DialerOpts {
	if x != nil {
		return x.Dialers
	}
	return nil
}

func (x *Config) GetHttpPath() string {
	if x != nil {
		return x.HttpPath
	}
	return ""
}

func (x *Config) GetDisableServePeerId() bool {
	if x != nil {
		return x.DisableServePeerId
	}
	return false
}

type Config_DialersEntry struct {
	unknownFields []byte
	Key           string             `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value         *dialer.DialerOpts `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Config_DialersEntry) Reset() {
	*x = Config_DialersEntry{}
}

func (*Config_DialersEntry) ProtoMessage() {}

func (x *Config_DialersEntry) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Config_DialersEntry) GetValue() *dialer.DialerOpts {
	if x != nil {
		return x.Value
	}
	return nil
}

func (m *Config) CloneVT() *Config {
	if m == nil {
		return (*Config)(nil)
	}
	r := new(Config)
	r.TransportPeerId = m.TransportPeerId
	r.ListenAddr = m.ListenAddr
	r.HttpPath = m.HttpPath
	r.DisableServePeerId = m.DisableServePeerId
	if rhs := m.Quic; rhs != nil {
		r.Quic = rhs.CloneVT()
	}
	if rhs := m.Dialers; rhs != nil {
		tmpContainer := make(map[string]*dialer.DialerOpts, len(rhs))
		for k, v := range rhs {
			tmpContainer[k] = v.CloneVT()
		}
		r.Dialers = tmpContainer
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
	if this.TransportPeerId != that.TransportPeerId {
		return false
	}
	if this.ListenAddr != that.ListenAddr {
		return false
	}
	if !this.Quic.EqualVT(that.Quic) {
		return false
	}
	if len(this.Dialers) != len(that.Dialers) {
		return false
	}
	for i, vx := range this.Dialers {
		vy, ok := that.Dialers[i]
		if !ok {
			return false
		}
		if p, q := vx, vy; p != q {
			if p == nil {
				p = &dialer.DialerOpts{}
			}
			if q == nil {
				q = &dialer.DialerOpts{}
			}
			if !p.EqualVT(q) {
				return false
			}
		}
	}
	if this.HttpPath != that.HttpPath {
		return false
	}
	if this.DisableServePeerId != that.DisableServePeerId {
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

// MarshalProtoJSON marshals the Config_DialersEntry message to JSON.
func (x *Config_DialersEntry) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Key != "" || s.HasField("key") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("key")
		s.WriteString(x.Key)
	}
	if x.Value != nil || s.HasField("value") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("value")
		x.Value.MarshalProtoJSON(s.WithField("value"))
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the Config_DialersEntry to JSON.
func (x *Config_DialersEntry) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the Config_DialersEntry message from JSON.
func (x *Config_DialersEntry) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "key":
			s.AddField("key")
			x.Key = s.ReadString()
		case "value":
			if s.ReadNil() {
				x.Value = nil
				return
			}
			x.Value = &dialer.DialerOpts{}
			x.Value.UnmarshalProtoJSON(s.WithField("value", true))
		}
	})
}

// UnmarshalJSON unmarshals the Config_DialersEntry from JSON.
func (x *Config_DialersEntry) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the Config message to JSON.
func (x *Config) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.TransportPeerId != "" || s.HasField("transportPeerId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("transportPeerId")
		s.WriteString(x.TransportPeerId)
	}
	if x.ListenAddr != "" || s.HasField("listenAddr") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("listenAddr")
		s.WriteString(x.ListenAddr)
	}
	if x.Quic != nil || s.HasField("quic") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("quic")
		x.Quic.MarshalProtoJSON(s.WithField("quic"))
	}
	if x.Dialers != nil || s.HasField("dialers") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("dialers")
		s.WriteObjectStart()
		var wroteElement bool
		for k, v := range x.Dialers {
			s.WriteMoreIf(&wroteElement)
			s.WriteObjectStringField(k)
			v.MarshalProtoJSON(s.WithField("dialers"))
		}
		s.WriteObjectEnd()
	}
	if x.HttpPath != "" || s.HasField("httpPath") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("httpPath")
		s.WriteString(x.HttpPath)
	}
	if x.DisableServePeerId || s.HasField("disableServePeerId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("disableServePeerId")
		s.WriteBool(x.DisableServePeerId)
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
		case "transport_peer_id", "transportPeerId":
			s.AddField("transport_peer_id")
			x.TransportPeerId = s.ReadString()
		case "listen_addr", "listenAddr":
			s.AddField("listen_addr")
			x.ListenAddr = s.ReadString()
		case "quic":
			if s.ReadNil() {
				x.Quic = nil
				return
			}
			x.Quic = &quic.Opts{}
			x.Quic.UnmarshalProtoJSON(s.WithField("quic", true))
		case "dialers":
			s.AddField("dialers")
			if s.ReadNil() {
				x.Dialers = nil
				return
			}
			x.Dialers = make(map[string]*dialer.DialerOpts)
			s.ReadStringMap(func(key string) {
				var v dialer.DialerOpts
				v.UnmarshalProtoJSON(s)
				x.Dialers[key] = &v
			})
		case "http_path", "httpPath":
			s.AddField("http_path")
			x.HttpPath = s.ReadString()
		case "disable_serve_peer_id", "disableServePeerId":
			s.AddField("disable_serve_peer_id")
			x.DisableServePeerId = s.ReadBool()
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
	if m.DisableServePeerId {
		i--
		if m.DisableServePeerId {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x30
	}
	if len(m.HttpPath) > 0 {
		i -= len(m.HttpPath)
		copy(dAtA[i:], m.HttpPath)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.HttpPath)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Dialers) > 0 {
		for k := range m.Dialers {
			v := m.Dialers[k]
			baseI := i
			size, err := v.MarshalToSizedBufferVT(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x22
		}
	}
	if m.Quic != nil {
		size, err := m.Quic.MarshalToSizedBufferVT(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.ListenAddr) > 0 {
		i -= len(m.ListenAddr)
		copy(dAtA[i:], m.ListenAddr)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ListenAddr)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.TransportPeerId) > 0 {
		i -= len(m.TransportPeerId)
		copy(dAtA[i:], m.TransportPeerId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.TransportPeerId)))
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
	l = len(m.TransportPeerId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.ListenAddr)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.Quic != nil {
		l = m.Quic.SizeVT()
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if len(m.Dialers) > 0 {
		for k, v := range m.Dialers {
			_ = k
			_ = v
			l = 0
			if v != nil {
				l = v.SizeVT()
			}
			l += 1 + protobuf_go_lite.SizeOfVarint(uint64(l))
			mapEntrySize := 1 + len(k) + protobuf_go_lite.SizeOfVarint(uint64(len(k))) + l
			n += mapEntrySize + 1 + protobuf_go_lite.SizeOfVarint(uint64(mapEntrySize))
		}
	}
	l = len(m.HttpPath)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	if m.DisableServePeerId {
		n += 2
	}
	n += len(m.unknownFields)
	return n
}

func (x *Config_DialersEntry) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("DialersEntry { ")
	if x.Key != "" {
		sb.WriteString(" key: ")
		sb.WriteString(strconv.Quote(x.Key))
	}
	if x.Value != nil {
		sb.WriteString(" value: ")
		sb.WriteString(x.Value.MarshalProtoText())
	}
	sb.WriteString("}")
	return sb.String()
}
func (x *Config_DialersEntry) String() string {
	return x.MarshalProtoText()
}
func (x *Config) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Config { ")
	if x.TransportPeerId != "" {
		sb.WriteString(" transport_peer_id: ")
		sb.WriteString(strconv.Quote(x.TransportPeerId))
	}
	if x.ListenAddr != "" {
		sb.WriteString(" listen_addr: ")
		sb.WriteString(strconv.Quote(x.ListenAddr))
	}
	if x.Quic != nil {
		sb.WriteString(" quic: ")
		sb.WriteString(x.Quic.MarshalProtoText())
	}
	if len(x.Dialers) > 0 {
		sb.WriteString(" dialers: {")
		for k, v := range x.Dialers {
			sb.WriteString(" ")
			sb.WriteString(strconv.Quote(k))
			sb.WriteString(": ")
			sb.WriteString(v.MarshalProtoText())
		}
		sb.WriteString(" }")
	}
	if x.HttpPath != "" {
		sb.WriteString(" http_path: ")
		sb.WriteString(strconv.Quote(x.HttpPath))
	}
	if x.DisableServePeerId {
		sb.WriteString(" disable_serve_peer_id: ")
		sb.WriteString(strconv.FormatBool(x.DisableServePeerId))
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
				return fmt.Errorf("proto: wrong wireType = %d for field TransportPeerId", wireType)
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
			m.TransportPeerId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ListenAddr", wireType)
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
			m.ListenAddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
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
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Dialers", wireType)
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
			if m.Dialers == nil {
				m.Dialers = make(map[string]*dialer.DialerOpts)
			}
			var mapkey string
			var mapvalue *dialer.DialerOpts
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
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
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return protobuf_go_lite.ErrIntOverflow
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return protobuf_go_lite.ErrIntOverflow
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					postmsgIndex := iNdEx + mapmsglen
					if postmsgIndex < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &dialer.DialerOpts{}
					if err := mapvalue.UnmarshalVT(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := protobuf_go_lite.Skip(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return protobuf_go_lite.ErrInvalidLength
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.Dialers[mapkey] = mapvalue
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HttpPath", wireType)
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
			m.HttpPath = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DisableServePeerId", wireType)
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
			m.DisableServePeerId = bool(v != 0)
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
