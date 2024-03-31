// Code generated by protoc-gen-srpc. DO NOT EDIT.
// protoc-gen-srpc version: v0.27.3
// source: github.com/aperturerobotics/bifrost/stream/drpc/e2e/e2e.proto

package drpc_e2e

import (
	context "context"

	srpc "github.com/aperturerobotics/starpc/srpc"
)

type SRPCEndToEndClient interface {
	SRPCClient() srpc.Client

	Mock(ctx context.Context, in *MockRequest) (*MockResponse, error)
}

type srpcEndToEndClient struct {
	cc        srpc.Client
	serviceID string
}

func NewSRPCEndToEndClient(cc srpc.Client) SRPCEndToEndClient {
	return &srpcEndToEndClient{cc: cc, serviceID: SRPCEndToEndServiceID}
}

func NewSRPCEndToEndClientWithServiceID(cc srpc.Client, serviceID string) SRPCEndToEndClient {
	if serviceID == "" {
		serviceID = SRPCEndToEndServiceID
	}
	return &srpcEndToEndClient{cc: cc, serviceID: serviceID}
}

func (c *srpcEndToEndClient) SRPCClient() srpc.Client { return c.cc }

func (c *srpcEndToEndClient) Mock(ctx context.Context, in *MockRequest) (*MockResponse, error) {
	out := new(MockResponse)
	err := c.cc.ExecCall(ctx, c.serviceID, "Mock", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type SRPCEndToEndServer interface {
	Mock(context.Context, *MockRequest) (*MockResponse, error)
}

type SRPCEndToEndUnimplementedServer struct{}

func (s *SRPCEndToEndUnimplementedServer) Mock(context.Context, *MockRequest) (*MockResponse, error) {
	return nil, srpc.ErrUnimplemented
}

const SRPCEndToEndServiceID = "drpc.e2e.EndToEnd"

type SRPCEndToEndHandler struct {
	serviceID string
	impl      SRPCEndToEndServer
}

// NewSRPCEndToEndHandler constructs a new RPC handler.
// serviceID: if empty, uses default: drpc.e2e.EndToEnd
func NewSRPCEndToEndHandler(impl SRPCEndToEndServer, serviceID string) srpc.Handler {
	if serviceID == "" {
		serviceID = SRPCEndToEndServiceID
	}
	return &SRPCEndToEndHandler{impl: impl, serviceID: serviceID}
}

// SRPCRegisterEndToEnd registers the implementation with the mux.
// Uses the default serviceID: drpc.e2e.EndToEnd
func SRPCRegisterEndToEnd(mux srpc.Mux, impl SRPCEndToEndServer) error {
	return mux.Register(NewSRPCEndToEndHandler(impl, ""))
}

func (d *SRPCEndToEndHandler) GetServiceID() string { return d.serviceID }

func (SRPCEndToEndHandler) GetMethodIDs() []string {
	return []string{
		"Mock",
	}
}

func (d *SRPCEndToEndHandler) InvokeMethod(
	serviceID, methodID string,
	strm srpc.Stream,
) (bool, error) {
	if serviceID != "" && serviceID != d.GetServiceID() {
		return false, nil
	}

	switch methodID {
	case "Mock":
		return true, d.InvokeMethod_Mock(d.impl, strm)
	default:
		return false, nil
	}
}

func (SRPCEndToEndHandler) InvokeMethod_Mock(impl SRPCEndToEndServer, strm srpc.Stream) error {
	req := new(MockRequest)
	if err := strm.MsgRecv(req); err != nil {
		return err
	}
	out, err := impl.Mock(strm.Context(), req)
	if err != nil {
		return err
	}
	return strm.MsgSend(out)
}

type SRPCEndToEnd_MockStream interface {
	srpc.Stream
}

type srpcEndToEnd_MockStream struct {
	srpc.Stream
}
