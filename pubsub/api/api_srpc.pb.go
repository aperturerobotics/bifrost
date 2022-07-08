// Code generated by protoc-gen-srpc. DO NOT EDIT.
// protoc-gen-srpc version: v0.8.1
// source: github.com/aperturerobotics/bifrost/pubsub/api/api.proto

package pubsub_api

import (
	context "context"

	srpc "github.com/aperturerobotics/starpc/srpc"
)

type SRPCPubSubServiceClient interface {
	SRPCClient() srpc.Client

	Subscribe(ctx context.Context) (SRPCPubSubService_SubscribeClient, error)
}

type srpcPubSubServiceClient struct {
	cc srpc.Client
}

func NewSRPCPubSubServiceClient(cc srpc.Client) SRPCPubSubServiceClient {
	return &srpcPubSubServiceClient{cc}
}

func (c *srpcPubSubServiceClient) SRPCClient() srpc.Client { return c.cc }

func (c *srpcPubSubServiceClient) Subscribe(ctx context.Context) (SRPCPubSubService_SubscribeClient, error) {
	stream, err := c.cc.NewStream(ctx, "pubsub.api.PubSubService", "Subscribe", nil)
	if err != nil {
		return nil, err
	}
	strm := &srpcPubSubService_SubscribeClient{stream}
	return strm, nil
}

type SRPCPubSubService_SubscribeClient interface {
	srpc.Stream
	Send(*SubscribeRequest) error
	Recv() (*SubscribeResponse, error)
	RecvTo(*SubscribeResponse) error
}

type srpcPubSubService_SubscribeClient struct {
	srpc.Stream
}

func (x *srpcPubSubService_SubscribeClient) Send(m *SubscribeRequest) error {
	return x.MsgSend(m)
}

func (x *srpcPubSubService_SubscribeClient) Recv() (*SubscribeResponse, error) {
	m := new(SubscribeResponse)
	if err := x.MsgRecv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *srpcPubSubService_SubscribeClient) RecvTo(m *SubscribeResponse) error {
	return x.MsgRecv(m)
}

type SRPCPubSubServiceServer interface {
	Subscribe(SRPCPubSubService_SubscribeStream) error
}

type SRPCPubSubServiceUnimplementedServer struct{}

func (s *SRPCPubSubServiceUnimplementedServer) Subscribe(SRPCPubSubService_SubscribeStream) error {
	return srpc.ErrUnimplemented
}

const SRPCPubSubServiceServiceID = "pubsub.api.PubSubService"

type SRPCPubSubServiceHandler struct {
	impl SRPCPubSubServiceServer
}

func (SRPCPubSubServiceHandler) GetServiceID() string { return SRPCPubSubServiceServiceID }

func (SRPCPubSubServiceHandler) GetMethodIDs() []string {
	return []string{
		"Subscribe",
	}
}

func (d *SRPCPubSubServiceHandler) InvokeMethod(
	serviceID, methodID string,
	strm srpc.Stream,
) (bool, error) {
	if serviceID != "" && serviceID != d.GetServiceID() {
		return false, nil
	}

	switch methodID {
	case "Subscribe":
		return true, d.InvokeMethod_Subscribe(d.impl, strm)
	default:
		return false, nil
	}
}

func (SRPCPubSubServiceHandler) InvokeMethod_Subscribe(impl SRPCPubSubServiceServer, strm srpc.Stream) error {
	clientStrm := &srpcPubSubService_SubscribeStream{strm}
	return impl.Subscribe(clientStrm)
}

func SRPCRegisterPubSubService(mux srpc.Mux, impl SRPCPubSubServiceServer) error {
	return mux.Register(&SRPCPubSubServiceHandler{impl: impl})
}

type SRPCPubSubService_SubscribeStream interface {
	srpc.Stream
	Send(*SubscribeResponse) error
	Recv() (*SubscribeRequest, error)
}

type srpcPubSubService_SubscribeStream struct {
	srpc.Stream
}

func (x *srpcPubSubService_SubscribeStream) Send(m *SubscribeResponse) error {
	return x.MsgSend(m)
}

func (x *srpcPubSubService_SubscribeStream) Recv() (*SubscribeRequest, error) {
	m := new(SubscribeRequest)
	if err := x.MsgRecv(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *srpcPubSubService_SubscribeStream) RecvTo(m *SubscribeRequest) error {
	return x.MsgRecv(m)
}
