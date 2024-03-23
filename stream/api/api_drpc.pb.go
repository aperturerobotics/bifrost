// Code generated by protoc-gen-go-drpc. DO NOT EDIT.
// protoc-gen-go-drpc version: v0.0.34
// source: github.com/aperturerobotics/bifrost/stream/api/api.proto

package stream_api

import (
	context "context"
	errors "errors"

	drpc1 "github.com/planetscale/vtprotobuf/codec/drpc"
	drpc "storj.io/drpc"
	drpcerr "storj.io/drpc/drpcerr"
)

type drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto struct{}

func (drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto) Marshal(msg drpc.Message) ([]byte, error) {
	return drpc1.Marshal(msg)
}

func (drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto) Unmarshal(buf []byte, msg drpc.Message) error {
	return drpc1.Unmarshal(buf, msg)
}

type DRPCStreamServiceClient interface {
	DRPCConn() drpc.Conn

	ForwardStreams(ctx context.Context, in *ForwardStreamsRequest) (DRPCStreamService_ForwardStreamsClient, error)
	ListenStreams(ctx context.Context, in *ListenStreamsRequest) (DRPCStreamService_ListenStreamsClient, error)
	AcceptStream(ctx context.Context) (DRPCStreamService_AcceptStreamClient, error)
	DialStream(ctx context.Context) (DRPCStreamService_DialStreamClient, error)
}

type drpcStreamServiceClient struct {
	cc drpc.Conn
}

func NewDRPCStreamServiceClient(cc drpc.Conn) DRPCStreamServiceClient {
	return &drpcStreamServiceClient{cc}
}

func (c *drpcStreamServiceClient) DRPCConn() drpc.Conn { return c.cc }

func (c *drpcStreamServiceClient) ForwardStreams(ctx context.Context, in *ForwardStreamsRequest) (DRPCStreamService_ForwardStreamsClient, error) {
	stream, err := c.cc.NewStream(ctx, "/stream.api.StreamService/ForwardStreams", drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
	if err != nil {
		return nil, err
	}
	x := &drpcStreamService_ForwardStreamsClient{stream}
	if err := x.MsgSend(in, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{}); err != nil {
		return nil, err
	}
	if err := x.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DRPCStreamService_ForwardStreamsClient interface {
	drpc.Stream
	Recv() (*ForwardStreamsResponse, error)
}

type drpcStreamService_ForwardStreamsClient struct {
	drpc.Stream
}

func (x *drpcStreamService_ForwardStreamsClient) GetStream() drpc.Stream {
	return x.Stream
}

func (x *drpcStreamService_ForwardStreamsClient) Recv() (*ForwardStreamsResponse, error) {
	m := new(ForwardStreamsResponse)
	if err := x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{}); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *drpcStreamService_ForwardStreamsClient) RecvMsg(m *ForwardStreamsResponse) error {
	return x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

func (c *drpcStreamServiceClient) ListenStreams(ctx context.Context, in *ListenStreamsRequest) (DRPCStreamService_ListenStreamsClient, error) {
	stream, err := c.cc.NewStream(ctx, "/stream.api.StreamService/ListenStreams", drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
	if err != nil {
		return nil, err
	}
	x := &drpcStreamService_ListenStreamsClient{stream}
	if err := x.MsgSend(in, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{}); err != nil {
		return nil, err
	}
	if err := x.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DRPCStreamService_ListenStreamsClient interface {
	drpc.Stream
	Recv() (*ListenStreamsResponse, error)
}

type drpcStreamService_ListenStreamsClient struct {
	drpc.Stream
}

func (x *drpcStreamService_ListenStreamsClient) GetStream() drpc.Stream {
	return x.Stream
}

func (x *drpcStreamService_ListenStreamsClient) Recv() (*ListenStreamsResponse, error) {
	m := new(ListenStreamsResponse)
	if err := x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{}); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *drpcStreamService_ListenStreamsClient) RecvMsg(m *ListenStreamsResponse) error {
	return x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

func (c *drpcStreamServiceClient) AcceptStream(ctx context.Context) (DRPCStreamService_AcceptStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, "/stream.api.StreamService/AcceptStream", drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
	if err != nil {
		return nil, err
	}
	x := &drpcStreamService_AcceptStreamClient{stream}
	return x, nil
}

type DRPCStreamService_AcceptStreamClient interface {
	drpc.Stream
	Send(*AcceptStreamRequest) error
	Recv() (*AcceptStreamResponse, error)
}

type drpcStreamService_AcceptStreamClient struct {
	drpc.Stream
}

func (x *drpcStreamService_AcceptStreamClient) GetStream() drpc.Stream {
	return x.Stream
}

func (x *drpcStreamService_AcceptStreamClient) Send(m *AcceptStreamRequest) error {
	return x.MsgSend(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

func (x *drpcStreamService_AcceptStreamClient) Recv() (*AcceptStreamResponse, error) {
	m := new(AcceptStreamResponse)
	if err := x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{}); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *drpcStreamService_AcceptStreamClient) RecvMsg(m *AcceptStreamResponse) error {
	return x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

func (c *drpcStreamServiceClient) DialStream(ctx context.Context) (DRPCStreamService_DialStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, "/stream.api.StreamService/DialStream", drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
	if err != nil {
		return nil, err
	}
	x := &drpcStreamService_DialStreamClient{stream}
	return x, nil
}

type DRPCStreamService_DialStreamClient interface {
	drpc.Stream
	Send(*DialStreamRequest) error
	Recv() (*DialStreamResponse, error)
}

type drpcStreamService_DialStreamClient struct {
	drpc.Stream
}

func (x *drpcStreamService_DialStreamClient) GetStream() drpc.Stream {
	return x.Stream
}

func (x *drpcStreamService_DialStreamClient) Send(m *DialStreamRequest) error {
	return x.MsgSend(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

func (x *drpcStreamService_DialStreamClient) Recv() (*DialStreamResponse, error) {
	m := new(DialStreamResponse)
	if err := x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{}); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *drpcStreamService_DialStreamClient) RecvMsg(m *DialStreamResponse) error {
	return x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

type DRPCStreamServiceServer interface {
	ForwardStreams(*ForwardStreamsRequest, DRPCStreamService_ForwardStreamsStream) error
	ListenStreams(*ListenStreamsRequest, DRPCStreamService_ListenStreamsStream) error
	AcceptStream(DRPCStreamService_AcceptStreamStream) error
	DialStream(DRPCStreamService_DialStreamStream) error
}

type DRPCStreamServiceUnimplementedServer struct{}

func (s *DRPCStreamServiceUnimplementedServer) ForwardStreams(*ForwardStreamsRequest, DRPCStreamService_ForwardStreamsStream) error {
	return drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

func (s *DRPCStreamServiceUnimplementedServer) ListenStreams(*ListenStreamsRequest, DRPCStreamService_ListenStreamsStream) error {
	return drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

func (s *DRPCStreamServiceUnimplementedServer) AcceptStream(DRPCStreamService_AcceptStreamStream) error {
	return drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

func (s *DRPCStreamServiceUnimplementedServer) DialStream(DRPCStreamService_DialStreamStream) error {
	return drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

type DRPCStreamServiceDescription struct{}

func (DRPCStreamServiceDescription) NumMethods() int { return 4 }

func (DRPCStreamServiceDescription) Method(n int) (string, drpc.Encoding, drpc.Receiver, interface{}, bool) {
	switch n {
	case 0:
		return "/stream.api.StreamService/ForwardStreams", drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return nil, srv.(DRPCStreamServiceServer).
					ForwardStreams(
						in1.(*ForwardStreamsRequest),
						&drpcStreamService_ForwardStreamsStream{in2.(drpc.Stream)},
					)
			}, DRPCStreamServiceServer.ForwardStreams, true
	case 1:
		return "/stream.api.StreamService/ListenStreams", drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return nil, srv.(DRPCStreamServiceServer).
					ListenStreams(
						in1.(*ListenStreamsRequest),
						&drpcStreamService_ListenStreamsStream{in2.(drpc.Stream)},
					)
			}, DRPCStreamServiceServer.ListenStreams, true
	case 2:
		return "/stream.api.StreamService/AcceptStream", drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return nil, srv.(DRPCStreamServiceServer).
					AcceptStream(
						&drpcStreamService_AcceptStreamStream{in1.(drpc.Stream)},
					)
			}, DRPCStreamServiceServer.AcceptStream, true
	case 3:
		return "/stream.api.StreamService/DialStream", drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return nil, srv.(DRPCStreamServiceServer).
					DialStream(
						&drpcStreamService_DialStreamStream{in1.(drpc.Stream)},
					)
			}, DRPCStreamServiceServer.DialStream, true
	default:
		return "", nil, nil, nil, false
	}
}

func DRPCRegisterStreamService(mux drpc.Mux, impl DRPCStreamServiceServer) error {
	return mux.Register(impl, DRPCStreamServiceDescription{})
}

type DRPCStreamService_ForwardStreamsStream interface {
	drpc.Stream
	Send(*ForwardStreamsResponse) error
}

type drpcStreamService_ForwardStreamsStream struct {
	drpc.Stream
}

func (x *drpcStreamService_ForwardStreamsStream) Send(m *ForwardStreamsResponse) error {
	return x.MsgSend(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

type DRPCStreamService_ListenStreamsStream interface {
	drpc.Stream
	Send(*ListenStreamsResponse) error
}

type drpcStreamService_ListenStreamsStream struct {
	drpc.Stream
}

func (x *drpcStreamService_ListenStreamsStream) Send(m *ListenStreamsResponse) error {
	return x.MsgSend(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

type DRPCStreamService_AcceptStreamStream interface {
	drpc.Stream
	Send(*AcceptStreamResponse) error
	Recv() (*AcceptStreamRequest, error)
}

type drpcStreamService_AcceptStreamStream struct {
	drpc.Stream
}

func (x *drpcStreamService_AcceptStreamStream) Send(m *AcceptStreamResponse) error {
	return x.MsgSend(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

func (x *drpcStreamService_AcceptStreamStream) Recv() (*AcceptStreamRequest, error) {
	m := new(AcceptStreamRequest)
	if err := x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{}); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *drpcStreamService_AcceptStreamStream) RecvMsg(m *AcceptStreamRequest) error {
	return x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

type DRPCStreamService_DialStreamStream interface {
	drpc.Stream
	Send(*DialStreamResponse) error
	Recv() (*DialStreamRequest, error)
}

type drpcStreamService_DialStreamStream struct {
	drpc.Stream
}

func (x *drpcStreamService_DialStreamStream) Send(m *DialStreamResponse) error {
	return x.MsgSend(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}

func (x *drpcStreamService_DialStreamStream) Recv() (*DialStreamRequest, error) {
	m := new(DialStreamRequest)
	if err := x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{}); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *drpcStreamService_DialStreamStream) RecvMsg(m *DialStreamRequest) error {
	return x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_stream_api_api_proto{})
}
