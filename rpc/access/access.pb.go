// Code generated by protoc-gen-go-lite. DO NOT EDIT.
// protoc-gen-go-lite version: v0.7.0
// source: github.com/aperturerobotics/bifrost/rpc/access/access.proto

package bifrost_rpc_access

import (
	fmt "fmt"
	io "io"
	strconv "strconv"
	strings "strings"

	protobuf_go_lite "github.com/aperturerobotics/protobuf-go-lite"
	json "github.com/aperturerobotics/protobuf-go-lite/json"
	_ "github.com/aperturerobotics/starpc/rpcstream"
)

// LookupRpcServiceRequest is a request to lookup an rpc service.
type LookupRpcServiceRequest struct {
	unknownFields []byte
	// ServiceId is the service identifier.
	ServiceId string `protobuf:"bytes,1,opt,name=service_id,json=serviceId,proto3" json:"serviceId,omitempty"`
	// ServerId is the identifier of the server requesting the service.
	// Can be empty.
	ServerId string `protobuf:"bytes,2,opt,name=server_id,json=serverId,proto3" json:"serverId,omitempty"`
}

func (x *LookupRpcServiceRequest) Reset() {
	*x = LookupRpcServiceRequest{}
}

func (*LookupRpcServiceRequest) ProtoMessage() {}

func (x *LookupRpcServiceRequest) GetServiceId() string {
	if x != nil {
		return x.ServiceId
	}
	return ""
}

func (x *LookupRpcServiceRequest) GetServerId() string {
	if x != nil {
		return x.ServerId
	}
	return ""
}

// LookupRpcServiceResponse is a response to LookupRpcService
type LookupRpcServiceResponse struct {
	unknownFields []byte
	// Idle indicates the directive is now idle.
	Idle bool `protobuf:"varint,1,opt,name=idle,proto3" json:"idle,omitempty"`
	// Exists indicates we found the service on the remote.
	Exists bool `protobuf:"varint,2,opt,name=exists,proto3" json:"exists,omitempty"`
	// Removed indicates the value no longer exists.
	Removed bool `protobuf:"varint,3,opt,name=removed,proto3" json:"removed,omitempty"`
}

func (x *LookupRpcServiceResponse) Reset() {
	*x = LookupRpcServiceResponse{}
}

func (*LookupRpcServiceResponse) ProtoMessage() {}

func (x *LookupRpcServiceResponse) GetIdle() bool {
	if x != nil {
		return x.Idle
	}
	return false
}

func (x *LookupRpcServiceResponse) GetExists() bool {
	if x != nil {
		return x.Exists
	}
	return false
}

func (x *LookupRpcServiceResponse) GetRemoved() bool {
	if x != nil {
		return x.Removed
	}
	return false
}

func (m *LookupRpcServiceRequest) CloneVT() *LookupRpcServiceRequest {
	if m == nil {
		return (*LookupRpcServiceRequest)(nil)
	}
	r := new(LookupRpcServiceRequest)
	r.ServiceId = m.ServiceId
	r.ServerId = m.ServerId
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *LookupRpcServiceRequest) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (m *LookupRpcServiceResponse) CloneVT() *LookupRpcServiceResponse {
	if m == nil {
		return (*LookupRpcServiceResponse)(nil)
	}
	r := new(LookupRpcServiceResponse)
	r.Idle = m.Idle
	r.Exists = m.Exists
	r.Removed = m.Removed
	if len(m.unknownFields) > 0 {
		r.unknownFields = make([]byte, len(m.unknownFields))
		copy(r.unknownFields, m.unknownFields)
	}
	return r
}

func (m *LookupRpcServiceResponse) CloneMessageVT() protobuf_go_lite.CloneMessage {
	return m.CloneVT()
}

func (this *LookupRpcServiceRequest) EqualVT(that *LookupRpcServiceRequest) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.ServiceId != that.ServiceId {
		return false
	}
	if this.ServerId != that.ServerId {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *LookupRpcServiceRequest) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*LookupRpcServiceRequest)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}
func (this *LookupRpcServiceResponse) EqualVT(that *LookupRpcServiceResponse) bool {
	if this == that {
		return true
	} else if this == nil || that == nil {
		return false
	}
	if this.Idle != that.Idle {
		return false
	}
	if this.Exists != that.Exists {
		return false
	}
	if this.Removed != that.Removed {
		return false
	}
	return string(this.unknownFields) == string(that.unknownFields)
}

func (this *LookupRpcServiceResponse) EqualMessageVT(thatMsg any) bool {
	that, ok := thatMsg.(*LookupRpcServiceResponse)
	if !ok {
		return false
	}
	return this.EqualVT(that)
}

// MarshalProtoJSON marshals the LookupRpcServiceRequest message to JSON.
func (x *LookupRpcServiceRequest) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.ServiceId != "" || s.HasField("serviceId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("serviceId")
		s.WriteString(x.ServiceId)
	}
	if x.ServerId != "" || s.HasField("serverId") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("serverId")
		s.WriteString(x.ServerId)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the LookupRpcServiceRequest to JSON.
func (x *LookupRpcServiceRequest) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the LookupRpcServiceRequest message from JSON.
func (x *LookupRpcServiceRequest) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "service_id", "serviceId":
			s.AddField("service_id")
			x.ServiceId = s.ReadString()
		case "server_id", "serverId":
			s.AddField("server_id")
			x.ServerId = s.ReadString()
		}
	})
}

// UnmarshalJSON unmarshals the LookupRpcServiceRequest from JSON.
func (x *LookupRpcServiceRequest) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the LookupRpcServiceResponse message to JSON.
func (x *LookupRpcServiceResponse) MarshalProtoJSON(s *json.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Idle || s.HasField("idle") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("idle")
		s.WriteBool(x.Idle)
	}
	if x.Exists || s.HasField("exists") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("exists")
		s.WriteBool(x.Exists)
	}
	if x.Removed || s.HasField("removed") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("removed")
		s.WriteBool(x.Removed)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the LookupRpcServiceResponse to JSON.
func (x *LookupRpcServiceResponse) MarshalJSON() ([]byte, error) {
	return json.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the LookupRpcServiceResponse message from JSON.
func (x *LookupRpcServiceResponse) UnmarshalProtoJSON(s *json.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.Skip() // ignore unknown field
		case "idle":
			s.AddField("idle")
			x.Idle = s.ReadBool()
		case "exists":
			s.AddField("exists")
			x.Exists = s.ReadBool()
		case "removed":
			s.AddField("removed")
			x.Removed = s.ReadBool()
		}
	})
}

// UnmarshalJSON unmarshals the LookupRpcServiceResponse from JSON.
func (x *LookupRpcServiceResponse) UnmarshalJSON(b []byte) error {
	return json.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

func (m *LookupRpcServiceRequest) MarshalVT() (dAtA []byte, err error) {
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

func (m *LookupRpcServiceRequest) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *LookupRpcServiceRequest) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if len(m.ServerId) > 0 {
		i -= len(m.ServerId)
		copy(dAtA[i:], m.ServerId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ServerId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ServiceId) > 0 {
		i -= len(m.ServiceId)
		copy(dAtA[i:], m.ServiceId)
		i = protobuf_go_lite.EncodeVarint(dAtA, i, uint64(len(m.ServiceId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *LookupRpcServiceResponse) MarshalVT() (dAtA []byte, err error) {
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

func (m *LookupRpcServiceResponse) MarshalToVT(dAtA []byte) (int, error) {
	size := m.SizeVT()
	return m.MarshalToSizedBufferVT(dAtA[:size])
}

func (m *LookupRpcServiceResponse) MarshalToSizedBufferVT(dAtA []byte) (int, error) {
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
	if m.Removed {
		i--
		if m.Removed {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x18
	}
	if m.Exists {
		i--
		if m.Exists {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x10
	}
	if m.Idle {
		i--
		if m.Idle {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *LookupRpcServiceRequest) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ServiceId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	l = len(m.ServerId)
	if l > 0 {
		n += 1 + l + protobuf_go_lite.SizeOfVarint(uint64(l))
	}
	n += len(m.unknownFields)
	return n
}

func (m *LookupRpcServiceResponse) SizeVT() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Idle {
		n += 2
	}
	if m.Exists {
		n += 2
	}
	if m.Removed {
		n += 2
	}
	n += len(m.unknownFields)
	return n
}

func (x *LookupRpcServiceRequest) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("LookupRpcServiceRequest {")
	if x.ServiceId != "" {
		if sb.Len() > 25 {
			sb.WriteString(" ")
		}
		sb.WriteString("service_id: ")
		sb.WriteString(strconv.Quote(x.ServiceId))
	}
	if x.ServerId != "" {
		if sb.Len() > 25 {
			sb.WriteString(" ")
		}
		sb.WriteString("server_id: ")
		sb.WriteString(strconv.Quote(x.ServerId))
	}
	sb.WriteString("}")
	return sb.String()
}

func (x *LookupRpcServiceRequest) String() string {
	return x.MarshalProtoText()
}
func (x *LookupRpcServiceResponse) MarshalProtoText() string {
	var sb strings.Builder
	sb.WriteString("LookupRpcServiceResponse {")
	if x.Idle != false {
		if sb.Len() > 26 {
			sb.WriteString(" ")
		}
		sb.WriteString("idle: ")
		sb.WriteString(strconv.FormatBool(x.Idle))
	}
	if x.Exists != false {
		if sb.Len() > 26 {
			sb.WriteString(" ")
		}
		sb.WriteString("exists: ")
		sb.WriteString(strconv.FormatBool(x.Exists))
	}
	if x.Removed != false {
		if sb.Len() > 26 {
			sb.WriteString(" ")
		}
		sb.WriteString("removed: ")
		sb.WriteString(strconv.FormatBool(x.Removed))
	}
	sb.WriteString("}")
	return sb.String()
}

func (x *LookupRpcServiceResponse) String() string {
	return x.MarshalProtoText()
}
func (m *LookupRpcServiceRequest) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: LookupRpcServiceRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LookupRpcServiceRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ServiceId", wireType)
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
			m.ServiceId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ServerId", wireType)
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
			m.ServerId = string(dAtA[iNdEx:postIndex])
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
func (m *LookupRpcServiceResponse) UnmarshalVT(dAtA []byte) error {
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
			return fmt.Errorf("proto: LookupRpcServiceResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: LookupRpcServiceResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Idle", wireType)
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
			m.Idle = bool(v != 0)
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Exists", wireType)
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
			m.Exists = bool(v != 0)
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Removed", wireType)
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
			m.Removed = bool(v != 0)
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
