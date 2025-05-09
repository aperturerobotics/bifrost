// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.9.1
// source: github.com/aperturerobotics/bifrost/pubsub/floodsub/floodsub.proto

package floodsub

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	hash "github.com/aperturerobotics/bifrost/hash"
	peer "github.com/aperturerobotics/bifrost/peer"
	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
)

// Config configures the floodsub router.
type Config struct {
	unknownFields []byte
	// PublishHashType is the hash type to use when signing published messages.
	// Defaults to sha256
	PublishHashType hash.HashType `protobuf:"varint,1,opt,name=publish_hash_type,json=publishHashType,proto3" json:"publishHashType,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
}

func (*Config) ProtoMessage() {}

func (x *Config) GetPublishHashType() hash.HashType {
	if x != nil {
		return x.PublishHashType
	}
	return hash.HashType(0)
}

// Packet is the floodsub packet.
type Packet struct {
	unknownFields []byte
	// Subscriptions contains any new subscription changes.
	Subscriptions []*SubscriptionOpts `protobuf:"bytes,1,rep,name=subscriptions,proto3" json:"subscriptions,omitempty"`
	// Publish contains messages we are publishing.
	Publish []*peer.SignedMsg `protobuf:"bytes,2,rep,name=publish,proto3" json:"publish,omitempty"`
}

func (x *Packet) Reset() {
	*x = Packet{}
}

func (*Packet) ProtoMessage() {}

func (x *Packet) GetSubscriptions() []*SubscriptionOpts {
	if x != nil {
		return x.Subscriptions
	}
	return nil
}

func (x *Packet) GetPublish() []*peer.SignedMsg {
	if x != nil {
		return x.Publish
	}
	return nil
}

// SubscriptionOpts are subscription options.
type SubscriptionOpts struct {
	unknownFields []byte
	// Subscribe indicates if we are subscribing to this channel ID.
	Subscribe bool `protobuf:"varint,1,opt,name=subscribe,proto3" json:"subscribe,omitempty"`
	// ChannelId is the channel to subscribe to.
	ChannelId string `protobuf:"bytes,2,opt,name=channel_id,json=channelId,proto3" json:"channelId,omitempty"`
}

func (x *SubscriptionOpts) Reset() {
	*x = SubscriptionOpts{}
}

func (*SubscriptionOpts) ProtoMessage() {}

func (x *SubscriptionOpts) GetSubscribe() bool {
	if x != nil {
		return x.Subscribe
	}
	return false
}

func (x *SubscriptionOpts) GetChannelId() string {
	if x != nil {
		return x.ChannelId
	}
	return ""
}

func (m *Config) CloneVT() *Config {
	if m == nil {
		return (*Config)(nil)
	}
	r := new(Config)
	r.PublishHashType = m.PublishHashType
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Config) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *Packet) CloneVT() *Packet {
	if m == nil {
		return (*Packet)(nil)
	}
	r := new(Packet)
	if rhs := m.Subscriptions; rhs != nil {
		tmpContainer := make([]*SubscriptionOpts, len(rhs))
		for k, v := range rhs {
			tmpContainer[k] = v.CloneVT()
		}
		r.Subscriptions = tmpContainer
	}
	if rhs := m.Publish; rhs != nil {
		tmpContainer := make([]*peer.SignedMsg, len(rhs))
		for k, v := range rhs {
			tmpContainer[k] = v.CloneVT()
		}
		r.Publish = tmpContainer
	}
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *Packet) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *SubscriptionOpts) CloneVT() *SubscriptionOpts {
	if m == nil {
		return (*SubscriptionOpts)(nil)
	}
	r := new(SubscriptionOpts)
	r.Subscribe = m.Subscribe
	r.ChannelId = m.ChannelId
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *SubscriptionOpts) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *Config) EqualVT(that *Config) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.PublishHashType != that.PublishHashType {
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
func (this *Packet) EqualVT(that *Packet) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if len(this.Subscriptions) != len(that.Subscriptions) {
		return false
	}
	for i, vx := range this.Subscriptions {
		vy := that.Subscriptions[i]
		if p, q := vx, vy; p != q {
			if p == nil {
				p = &SubscriptionOpts{}
			}
			if q == nil {
				q = &SubscriptionOpts{}
			}
			if !p.EqualVT(q) {
				return false
			}
		}
	}
	if len(this.Publish) != len(that.Publish) {
		return false
	}
	for i, vx := range this.Publish {
		vy := that.Publish[i]
		if p, q := vx, vy; p != q {
			if p == nil {
				p = &peer.SignedMsg{}
			}
			if q == nil {
				q = &peer.SignedMsg{}
			}
			if !p.EqualVT(q) {
				return false
			}
		}
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *Packet) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*Packet)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *SubscriptionOpts) EqualVT(that *SubscriptionOpts) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.Subscribe != that.Subscribe {
		return false
	}
	if this.ChannelId != that.ChannelId {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *SubscriptionOpts) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*SubscriptionOpts)
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
	if x.PublishHashType != 0 || s.HasField("publishHashType") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("publishHashType")
		x.PublishHashType.MarshalProtoJSON(s)
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
		case "publish_hash_type", "publishHashType":
			s.AddField("publish_hash_type")
			x.PublishHashType.UnmarshalProtoJSON(s)
		}
	})
}

// UnmarshalJSON unmarshals the Config from JSON.
func (x *Config) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the Packet message to JSON.
func (x *Packet) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if len(x.Subscriptions) > 0 || s.HasField("subscriptions") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("subscriptions")
		s.WriteArrayStart()
		var wroteElement bool
		for _, element := range x.Subscriptions {
			s.WriteMoreIf(&wroteElement)
			element.MarshalProtoJSON(s.WithField("subscriptions"))
		}
		s.WriteArrayEnd()
	}
	if len(x.Publish) > 0 || s.HasField("publish") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("publish")
		s.WriteArrayStart()
		var wroteElement bool
		for _, element := range x.Publish {
			s.WriteMoreIf(&wroteElement)
			element.MarshalProtoJSON(s.WithField("publish"))
		}
		s.WriteArrayEnd()
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the Packet to JSON.
func (x *Packet) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the Packet message from JSON.
func (x *Packet) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "subscriptions":
			s.AddField("subscriptions")
			if s.ReadNil() {
				x.Subscriptions = nil
				return
			}
			s.ReadArray(func() {
				if s.ReadNil() {
					x.Subscriptions = append(x.Subscriptions, nil)
					return
				}
				v := &SubscriptionOpts{}
				v.UnmarshalProtoJSON(s.WithField("subscriptions", false))
				if s.Err() != nil {
					return
				}
				x.Subscriptions = append(x.Subscriptions, v)
			})
		case "publish":
			s.AddField("publish")
			if s.ReadNil() {
				x.Publish = nil
				return
			}
			s.ReadArray(func() {
				if s.ReadNil() {
					x.Publish = append(x.Publish, nil)
					return
				}
				v := &peer.SignedMsg{}
				v.UnmarshalProtoJSON(s.WithField("publish", false))
				if s.Err() != nil {
					return
				}
				x.Publish = append(x.Publish, v)
			})
		}
	})
}

// UnmarshalJSON unmarshals the Packet from JSON.
func (x *Packet) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the SubscriptionOpts message to JSON.
func (x *SubscriptionOpts) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Subscribe || s.HasField("subscribe") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("subscribe")
		s.WriteBool(x.Subscribe)
	}
	if x.ChannelId != "" || s.HasField("channelId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("channelId")
		s.WriteString(x.ChannelId)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the SubscriptionOpts to JSON.
func (x *SubscriptionOpts) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the SubscriptionOpts message from JSON.
func (x *SubscriptionOpts) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "subscribe":
			s.AddField("subscribe")
			x.Subscribe = s.ReadBool()
		case "channel_id", "channelId":
			s.AddField("channel_id")
			x.ChannelId = s.ReadString()
		}
	})
}

// UnmarshalJSON unmarshals the SubscriptionOpts from JSON.
func (x *SubscriptionOpts) UnmarshalJSON(b []byte) error {
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
	if m.PublishHashType != 0 {
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(m.PublishHashType))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Packet) MarshalVT() (dAtA []byte, err error) {
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

func (m *Packet) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *Packet) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if len(m.Publish) > 0 {
		for iNdEx := len(m.Publish) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.Publish[iNdEx].MarshalToSizedBufferVT(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.Subscriptions) > 0 {
		for iNdEx := len(m.Subscriptions) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.Subscriptions[iNdEx].MarshalToSizedBufferVT(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(size))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *SubscriptionOpts) MarshalVT() (dAtA []byte, err error) {
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

func (m *SubscriptionOpts) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *SubscriptionOpts) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if len(m.ChannelId) > 0 {
		i -= len(m.ChannelId)
		copy(dAtA[i:], m.ChannelId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ChannelId)))
		i--
		dAtA[i] = 0x12
	}
	if m.Subscribe {
		i--
		if m.Subscribe {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *Config) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PublishHashType != 0 {
		n += 1 + protobuf_go_lite.SizeOfVarint(uint64(m.PublishHashType))
	}
	n += len(m.unknownFields)
	return n
}

func (m *Packet) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Subscriptions) > 0 {
		for _, e := range m.Subscriptions {
			l = e.SizeVT()
			n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
		}
	}
	if len(m.Publish) > 0 {
		for _, e := range m.Publish {
			l = e.SizeVT()
			n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
		}
	}
	n += len(m.unknownFields)
	return n
}

func (m *SubscriptionOpts) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Subscribe {
		n += 2
	}
	l = len(m.ChannelId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (x *Config) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Config {")
	if x.PublishHashType != 0 {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("publish_hash_type: ")
		sb.WriteString("\"")
		sb.WriteString(hash.HashType(x.PublishHashType).String())
		sb.WriteString("\"")
	}
	sb.WriteString("}")
	return sb.String()
}

func (x *Config) String() string {
	return x.MarshalProtoText()
}
func (x *Packet) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("Packet {")
	if len(x.Subscriptions) > 0 {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("subscriptions: [")
		for i, v := range x.Subscriptions {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(v.MarshalProtoText())
		}
		sb.WriteString("]")
	}
	if len(x.Publish) > 0 {
		if sb.Len() > 8 {
			sb.WriteString(" ")
		}
		sb.WriteString("publish: [")
		for i, v := range x.Publish {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(v.MarshalProtoText())
		}
		sb.WriteString("]")
	}
	sb.WriteString("}")
	return sb.String()
}

func (x *Packet) String() string {
	return x.MarshalProtoText()
}
func (x *SubscriptionOpts) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("SubscriptionOpts {")
	if x.Subscribe != false {
		if sb.Len() > 18 {
			sb.WriteString(" ")
		}
		sb.WriteString("subscribe: ")
		sb.WriteString(strconv.FormatBool(x.Subscribe))
	}
	if x.ChannelId != "" {
		if sb.Len() > 18 {
			sb.WriteString(" ")
		}
		sb.WriteString("channel_id: ")
		sb.WriteString(strconv.Quote(x.ChannelId))
	}
	sb.WriteString("}")
	return sb.String()
}

func (x *SubscriptionOpts) String() string {
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
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublishHashType", wireType)
			}
			m.PublishHashType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protobuf_go_lite.ErrIntOverflow
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PublishHashType |= hash.HashType(b&0x7F) << shift
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
func (m *Packet) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: Packet: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Packet: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subscriptions", wireType)
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
			m.Subscriptions = append(m.Subscriptions, &SubscriptionOpts{})
			if err := m.Subscriptions[len(m.Subscriptions)-1].UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Publish", wireType)
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
			m.Publish = append(m.Publish, &peer.SignedMsg{})
			if err := m.Publish[len(m.Publish)-1].UnmarshalVT(dAtA[iNdEx:postIndex]); err != nil {
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
func (m *SubscriptionOpts) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: SubscriptionOpts: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SubscriptionOpts: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subscribe", wireType)
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
			m.Subscribe = bool(v != 0)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChannelId", wireType)
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
			m.ChannelId = string(dAtA[iNdEx:postIndex])
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
