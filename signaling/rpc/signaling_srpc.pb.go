// Code generated by protoc-gen-srpc. DO NOT EDIT.
// protoc-gen-srpc version: v0.35.1
// source: github.com/aperturerobotics/bifrost/signaling/rpc/signaling.proto

package signaling_rpc

import (
	context "context"

	srpc "github.com/aperturerobotics/starpc/srpc"
)

type SRPCSignalingClient interface {
	SRPCClient() srpc.Client

	Listen(ctx context.Context, in *ListenRequest) (SRPCSignaling_ListenClient, error)
	Session(ctx context.Context) (SRPCSignaling_SessionClient, error)
}

type srpcSignalingClient struct {
	cc        srpc.Client
	serviceID string
}

func NewSRPCSignalingClient(cc srpc.Client) SRPCSignalingClient {
	return &srpcSignalingClient{cc: cc, serviceID: SRPCSignalingServiceID}
}

func NewSRPCSignalingClientWithServiceID(cc srpc.Client, serviceID string) SRPCSignalingClient {
	if serviceID == "" {
		serviceID = SRPCSignalingServiceID
	}
	return &srpcSignalingClient{cc: cc, serviceID: serviceID}
}

func (c *srpcSignalingClient) SRPCClient() srpc.Client { return c.cc }

func (c *srpcSignalingClient) Listen(ctx context.Context, in *ListenRequest) (SRPCSignaling_ListenClient, error) {
	stream, err := c.cc.NewStream(ctx, c.serviceID, "Listen", in)
	if err != nil {
		return nil, err
	}
	strm := &srpcSignaling_ListenClient{stream}
	if err := strm.CloseSend(); err != nil {
		return nil, err
	}
	return strm, nil
}

type SRPCSignaling_ListenClient interface {
	srpc.Stream
	Recv() (*ListenResponse, error)
	RecvTo(*ListenResponse) error
}

type srpcSignaling_ListenClient struct {
	srpc.Stream
}

func (x *srpcSignaling_ListenClient) Recv() (*ListenResponse, error) {
	m := new(ListenResponse)
	if err := x.MsgRecv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *srpcSignaling_ListenClient) RecvTo(m *ListenResponse) error {
	return x.MsgRecv(m)
}

func (c *srpcSignalingClient) Session(ctx context.Context) (SRPCSignaling_SessionClient, error) {
	stream, err := c.cc.NewStream(ctx, c.serviceID, "Session", nil)
	if err != nil {
		return nil, err
	}
	strm := &srpcSignaling_SessionClient{stream}
	return strm, nil
}

type SRPCSignaling_SessionClient interface {
	srpc.Stream
	Send(*SessionRequest) error
	Recv() (*SessionResponse, error)
	RecvTo(*SessionResponse) error
}

type srpcSignaling_SessionClient struct {
	srpc.Stream
}

func (x *srpcSignaling_SessionClient) Send(m *SessionRequest) error {
	if m == nil {
		return nil
	}
	return x.MsgSend(m)
}

func (x *srpcSignaling_SessionClient) Recv() (*SessionResponse, error) {
	m := new(SessionResponse)
	if err := x.MsgRecv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *srpcSignaling_SessionClient) RecvTo(m *SessionResponse) error {
	return x.MsgRecv(m)
}

type SRPCSignalingServer interface {
	Listen(*ListenRequest, SRPCSignaling_ListenStream) error
	Session(SRPCSignaling_SessionStream) error
}

const SRPCSignalingServiceID = "signaling.rpc.Signaling"

type SRPCSignalingHandler struct {
	serviceID string
	impl      SRPCSignalingServer
}

// NewSRPCSignalingHandler constructs a new RPC handler.
// serviceID: if empty, uses default: signaling.rpc.Signaling
func NewSRPCSignalingHandler(impl SRPCSignalingServer, serviceID string) srpc.Handler {
	if serviceID == "" {
		serviceID = SRPCSignalingServiceID
	}
	return &SRPCSignalingHandler{impl: impl, serviceID: serviceID}
}

// SRPCRegisterSignaling registers the implementation with the mux.
// Uses the default serviceID: signaling.rpc.Signaling
func SRPCRegisterSignaling(mux srpc.Mux, impl SRPCSignalingServer) error {
	return mux.Register(NewSRPCSignalingHandler(impl, ""))
}

func (d *SRPCSignalingHandler) GetServiceID() string { return d.serviceID }

func (SRPCSignalingHandler) GetMethodIDs() []string {
	return []string{
		"Listen",
		"Session",
	}
}

func (d *SRPCSignalingHandler) InvokeMethod(
	serviceID, methodID string,
	strm srpc.Stream,
) (bool, error) {
	if serviceID != "" && serviceID != d.GetServiceID() {
		return false, nil
	}

	switch methodID {
	case "Listen":
		return true, d.InvokeMethod_Listen(d.impl, strm)
	case "Session":
		return true, d.InvokeMethod_Session(d.impl, strm)
	default:
		return false, nil
	}
}

func (SRPCSignalingHandler) InvokeMethod_Listen(impl SRPCSignalingServer, strm srpc.Stream) error {
	req := new(ListenRequest)
	if err := strm.MsgRecv(req); err != nil {
		return err
	}
	serverStrm := &srpcSignaling_ListenStream{strm}
	return impl.Listen(req, serverStrm)
}

func (SRPCSignalingHandler) InvokeMethod_Session(impl SRPCSignalingServer, strm srpc.Stream) error {
	clientStrm := &srpcSignaling_SessionStream{strm}
	return impl.Session(clientStrm)
}

type SRPCSignaling_ListenStream interface {
	srpc.Stream
	Send(*ListenResponse) error
	SendAndClose(*ListenResponse) error
}

type srpcSignaling_ListenStream struct {
	srpc.Stream
}

func (x *srpcSignaling_ListenStream) Send(m *ListenResponse) error {
	return x.MsgSend(m)
}

func (x *srpcSignaling_ListenStream) SendAndClose(m *ListenResponse) error {
	if m != nil {
		if err := x.MsgSend(m); err != nil {
			return err
		}
	}
	return x.CloseSend()
}

type SRPCSignaling_SessionStream interface {
	srpc.Stream
	Send(*SessionResponse) error
	SendAndClose(*SessionResponse) error
	Recv() (*SessionRequest, error)
	RecvTo(*SessionRequest) error
}

type srpcSignaling_SessionStream struct {
	srpc.Stream
}

func (x *srpcSignaling_SessionStream) Send(m *SessionResponse) error {
	return x.MsgSend(m)
}

func (x *srpcSignaling_SessionStream) SendAndClose(m *SessionResponse) error {
	if m != nil {
		if err := x.MsgSend(m); err != nil {
			return err
		}
	}
	return x.CloseSend()
}

func (x *srpcSignaling_SessionStream) Recv() (*SessionRequest, error) {
	m := new(SessionRequest)
	if err := x.MsgRecv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *srpcSignaling_SessionStream) RecvTo(m *SessionRequest) error {
	return x.MsgRecv(m)
}
