// Code generated by protoc-gen-srpc. DO NOT EDIT.
// protoc-gen-srpc version: v0.31.12
// source: github.com/aperturerobotics/bifrost/rpc/access/access.proto

package bifrost_rpc_access

import (
	context "context"

	rpcstream "github.com/aperturerobotics/starpc/rpcstream"
	srpc "github.com/aperturerobotics/starpc/srpc"
)

type SRPCAccessRpcServiceClient interface {
	SRPCClient() srpc.Client

	LookupRpcService(ctx context.Context, in *LookupRpcServiceRequest) (SRPCAccessRpcService_LookupRpcServiceClient, error)
	CallRpcService(ctx context.Context) (SRPCAccessRpcService_CallRpcServiceClient, error)
}

type srpcAccessRpcServiceClient struct {
	cc        srpc.Client
	serviceID string
}

func NewSRPCAccessRpcServiceClient(cc srpc.Client) SRPCAccessRpcServiceClient {
	return &srpcAccessRpcServiceClient{cc: cc, serviceID: SRPCAccessRpcServiceServiceID}
}

func NewSRPCAccessRpcServiceClientWithServiceID(cc srpc.Client, serviceID string) SRPCAccessRpcServiceClient {
	if serviceID == "" {
		serviceID = SRPCAccessRpcServiceServiceID
	}
	return &srpcAccessRpcServiceClient{cc: cc, serviceID: serviceID}
}

func (c *srpcAccessRpcServiceClient) SRPCClient() srpc.Client { return c.cc }

func (c *srpcAccessRpcServiceClient) LookupRpcService(ctx context.Context, in *LookupRpcServiceRequest) (SRPCAccessRpcService_LookupRpcServiceClient, error) {
	stream, err := c.cc.NewStream(ctx, c.serviceID, "LookupRpcService", in)
	if err != nil {
		return nil, err
	}
	strm := &srpcAccessRpcService_LookupRpcServiceClient{stream}
	if err := strm.CloseSend(); err != nil {
		return nil, err
	}
	return strm, nil
}

type SRPCAccessRpcService_LookupRpcServiceClient interface {
	srpc.Stream
	Recv() (*LookupRpcServiceResponse, error)
	RecvTo(*LookupRpcServiceResponse) error
}

type srpcAccessRpcService_LookupRpcServiceClient struct {
	srpc.Stream
}

func (x *srpcAccessRpcService_LookupRpcServiceClient) Recv() (*LookupRpcServiceResponse, error) {
	m := new(LookupRpcServiceResponse)
	if err := x.MsgRecv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *srpcAccessRpcService_LookupRpcServiceClient) RecvTo(m *LookupRpcServiceResponse) error {
	return x.MsgRecv(m)
}

func (c *srpcAccessRpcServiceClient) CallRpcService(ctx context.Context) (SRPCAccessRpcService_CallRpcServiceClient, error) {
	stream, err := c.cc.NewStream(ctx, c.serviceID, "CallRpcService", nil)
	if err != nil {
		return nil, err
	}
	strm := &srpcAccessRpcService_CallRpcServiceClient{stream}
	return strm, nil
}

type SRPCAccessRpcService_CallRpcServiceClient interface {
	srpc.Stream
	Send(*rpcstream.RpcStreamPacket) error
	Recv() (*rpcstream.RpcStreamPacket, error)
	RecvTo(*rpcstream.RpcStreamPacket) error
}

type srpcAccessRpcService_CallRpcServiceClient struct {
	srpc.Stream
}

func (x *srpcAccessRpcService_CallRpcServiceClient) Send(m *rpcstream.RpcStreamPacket) error {
	if m == nil {
		return nil
	}
	return x.MsgSend(m)
}

func (x *srpcAccessRpcService_CallRpcServiceClient) Recv() (*rpcstream.RpcStreamPacket, error) {
	m := new(rpcstream.RpcStreamPacket)
	if err := x.MsgRecv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *srpcAccessRpcService_CallRpcServiceClient) RecvTo(m *rpcstream.RpcStreamPacket) error {
	return x.MsgRecv(m)
}

type SRPCAccessRpcServiceServer interface {
	LookupRpcService(*LookupRpcServiceRequest, SRPCAccessRpcService_LookupRpcServiceStream) error
	CallRpcService(SRPCAccessRpcService_CallRpcServiceStream) error
}

type SRPCAccessRpcServiceUnimplementedServer struct{}

func (s *SRPCAccessRpcServiceUnimplementedServer) LookupRpcService(*LookupRpcServiceRequest, SRPCAccessRpcService_LookupRpcServiceStream) error {
	return srpc.ErrUnimplemented
}

func (s *SRPCAccessRpcServiceUnimplementedServer) CallRpcService(SRPCAccessRpcService_CallRpcServiceStream) error {
	return srpc.ErrUnimplemented
}

const SRPCAccessRpcServiceServiceID = "bifrost.rpc.access.AccessRpcService"

type SRPCAccessRpcServiceHandler struct {
	serviceID string
	impl      SRPCAccessRpcServiceServer
}

// NewSRPCAccessRpcServiceHandler constructs a new RPC handler.
// serviceID: if empty, uses default: bifrost.rpc.access.AccessRpcService
func NewSRPCAccessRpcServiceHandler(impl SRPCAccessRpcServiceServer, serviceID string) srpc.Handler {
	if serviceID == "" {
		serviceID = SRPCAccessRpcServiceServiceID
	}
	return &SRPCAccessRpcServiceHandler{impl: impl, serviceID: serviceID}
}

// SRPCRegisterAccessRpcService registers the implementation with the mux.
// Uses the default serviceID: bifrost.rpc.access.AccessRpcService
func SRPCRegisterAccessRpcService(mux srpc.Mux, impl SRPCAccessRpcServiceServer) error {
	return mux.Register(NewSRPCAccessRpcServiceHandler(impl, ""))
}

func (d *SRPCAccessRpcServiceHandler) GetServiceID() string { return d.serviceID }

func (SRPCAccessRpcServiceHandler) GetMethodIDs() []string {
	return []string{
		"LookupRpcService",
		"CallRpcService",
	}
}

func (d *SRPCAccessRpcServiceHandler) InvokeMethod(
	serviceID, methodID string,
	strm srpc.Stream,
) (bool, error) {
	if serviceID != "" && serviceID != d.GetServiceID() {
		return false, nil
	}

	switch methodID {
	case "LookupRpcService":
		return true, d.InvokeMethod_LookupRpcService(d.impl, strm)
	case "CallRpcService":
		return true, d.InvokeMethod_CallRpcService(d.impl, strm)
	default:
		return false, nil
	}
}

func (SRPCAccessRpcServiceHandler) InvokeMethod_LookupRpcService(impl SRPCAccessRpcServiceServer, strm srpc.Stream) error {
	req := new(LookupRpcServiceRequest)
	if err := strm.MsgRecv(req); err != nil {
		return err
	}
	serverStrm := &srpcAccessRpcService_LookupRpcServiceStream{strm}
	return impl.LookupRpcService(req, serverStrm)
}

func (SRPCAccessRpcServiceHandler) InvokeMethod_CallRpcService(impl SRPCAccessRpcServiceServer, strm srpc.Stream) error {
	clientStrm := &srpcAccessRpcService_CallRpcServiceStream{strm}
	return impl.CallRpcService(clientStrm)
}

type SRPCAccessRpcService_LookupRpcServiceStream interface {
	srpc.Stream
	Send(*LookupRpcServiceResponse) error
	SendAndClose(*LookupRpcServiceResponse) error
}

type srpcAccessRpcService_LookupRpcServiceStream struct {
	srpc.Stream
}

func (x *srpcAccessRpcService_LookupRpcServiceStream) Send(m *LookupRpcServiceResponse) error {
	return x.MsgSend(m)
}

func (x *srpcAccessRpcService_LookupRpcServiceStream) SendAndClose(m *LookupRpcServiceResponse) error {
	if m != nil {
		if err := x.MsgSend(m); err != nil {
			return err
		}
	}
	return x.CloseSend()
}

type SRPCAccessRpcService_CallRpcServiceStream interface {
	srpc.Stream
	Send(*rpcstream.RpcStreamPacket) error
	SendAndClose(*rpcstream.RpcStreamPacket) error
	Recv() (*rpcstream.RpcStreamPacket, error)
	RecvTo(*rpcstream.RpcStreamPacket) error
}

type srpcAccessRpcService_CallRpcServiceStream struct {
	srpc.Stream
}

func (x *srpcAccessRpcService_CallRpcServiceStream) Send(m *rpcstream.RpcStreamPacket) error {
	return x.MsgSend(m)
}

func (x *srpcAccessRpcService_CallRpcServiceStream) SendAndClose(m *rpcstream.RpcStreamPacket) error {
	if m != nil {
		if err := x.MsgSend(m); err != nil {
			return err
		}
	}
	return x.CloseSend()
}

func (x *srpcAccessRpcService_CallRpcServiceStream) Recv() (*rpcstream.RpcStreamPacket, error) {
	m := new(rpcstream.RpcStreamPacket)
	if err := x.MsgRecv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *srpcAccessRpcService_CallRpcServiceStream) RecvTo(m *rpcstream.RpcStreamPacket) error {
	return x.MsgRecv(m)
}
