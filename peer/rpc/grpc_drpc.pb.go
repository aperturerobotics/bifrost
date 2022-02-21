// Code generated by protoc-gen-go-drpc. DO NOT EDIT.
// protoc-gen-go-drpc version: v0.0.29
// source: github.com/aperturerobotics/bifrost/peer/rpc/grpc.proto

package peer_grpc

import (
	context "context"
	errors "errors"

	proto "github.com/golang/protobuf/proto"
	drpc "storj.io/drpc"
	drpcerr "storj.io/drpc/drpcerr"
)

type drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto struct{}

func (drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto) Marshal(msg drpc.Message) ([]byte, error) {
	return proto.Marshal(msg)
}

func (drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto) Unmarshal(buf []byte, msg drpc.Message) error {
	return proto.Unmarshal(buf, msg)
}

type DRPCPeerServiceClient interface {
	DRPCConn() drpc.Conn

	Identify(ctx context.Context, in *IdentifyRequest) (DRPCPeerService_IdentifyClient, error)
	GetPeerInfo(ctx context.Context, in *GetPeerInfoRequest) (*GetPeerInfoResponse, error)
}

type drpcPeerServiceClient struct {
	cc drpc.Conn
}

func NewDRPCPeerServiceClient(cc drpc.Conn) DRPCPeerServiceClient {
	return &drpcPeerServiceClient{cc}
}

func (c *drpcPeerServiceClient) DRPCConn() drpc.Conn { return c.cc }

func (c *drpcPeerServiceClient) Identify(ctx context.Context, in *IdentifyRequest) (DRPCPeerService_IdentifyClient, error) {
	stream, err := c.cc.NewStream(ctx, "/peer.grpc.PeerService/Identify", drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{})
	if err != nil {
		return nil, err
	}
	x := &drpcPeerService_IdentifyClient{stream}
	if err := x.MsgSend(in, drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{}); err != nil {
		return nil, err
	}
	if err := x.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type DRPCPeerService_IdentifyClient interface {
	drpc.Stream
	Recv() (*IdentifyResponse, error)
}

type drpcPeerService_IdentifyClient struct {
	drpc.Stream
}

func (x *drpcPeerService_IdentifyClient) Recv() (*IdentifyResponse, error) {
	m := new(IdentifyResponse)
	if err := x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{}); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *drpcPeerService_IdentifyClient) RecvMsg(m *IdentifyResponse) error {
	return x.MsgRecv(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{})
}

func (c *drpcPeerServiceClient) GetPeerInfo(ctx context.Context, in *GetPeerInfoRequest) (*GetPeerInfoResponse, error) {
	out := new(GetPeerInfoResponse)
	err := c.cc.Invoke(ctx, "/peer.grpc.PeerService/GetPeerInfo", drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{}, in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type DRPCPeerServiceServer interface {
	Identify(*IdentifyRequest, DRPCPeerService_IdentifyStream) error
	GetPeerInfo(context.Context, *GetPeerInfoRequest) (*GetPeerInfoResponse, error)
}

type DRPCPeerServiceUnimplementedServer struct{}

func (s *DRPCPeerServiceUnimplementedServer) Identify(*IdentifyRequest, DRPCPeerService_IdentifyStream) error {
	return drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

func (s *DRPCPeerServiceUnimplementedServer) GetPeerInfo(context.Context, *GetPeerInfoRequest) (*GetPeerInfoResponse, error) {
	return nil, drpcerr.WithCode(errors.New("Unimplemented"), drpcerr.Unimplemented)
}

type DRPCPeerServiceDescription struct{}

func (DRPCPeerServiceDescription) NumMethods() int { return 2 }

func (DRPCPeerServiceDescription) Method(n int) (string, drpc.Encoding, drpc.Receiver, interface{}, bool) {
	switch n {
	case 0:
		return "/peer.grpc.PeerService/Identify", drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return nil, srv.(DRPCPeerServiceServer).
					Identify(
						in1.(*IdentifyRequest),
						&drpcPeerService_IdentifyStream{in2.(drpc.Stream)},
					)
			}, DRPCPeerServiceServer.Identify, true
	case 1:
		return "/peer.grpc.PeerService/GetPeerInfo", drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{},
			func(srv interface{}, ctx context.Context, in1, in2 interface{}) (drpc.Message, error) {
				return srv.(DRPCPeerServiceServer).
					GetPeerInfo(
						ctx,
						in1.(*GetPeerInfoRequest),
					)
			}, DRPCPeerServiceServer.GetPeerInfo, true
	default:
		return "", nil, nil, nil, false
	}
}

func DRPCRegisterPeerService(mux drpc.Mux, impl DRPCPeerServiceServer) error {
	return mux.Register(impl, DRPCPeerServiceDescription{})
}

type DRPCPeerService_IdentifyStream interface {
	drpc.Stream
	Send(*IdentifyResponse) error
}

type drpcPeerService_IdentifyStream struct {
	drpc.Stream
}

func (x *drpcPeerService_IdentifyStream) Send(m *IdentifyResponse) error {
	return x.MsgSend(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{})
}

type DRPCPeerService_GetPeerInfoStream interface {
	drpc.Stream
	SendAndClose(*GetPeerInfoResponse) error
}

type drpcPeerService_GetPeerInfoStream struct {
	drpc.Stream
}

func (x *drpcPeerService_GetPeerInfoStream) SendAndClose(m *GetPeerInfoResponse) error {
	if err := x.MsgSend(m, drpcEncoding_File_github_com_aperturerobotics_bifrost_peer_rpc_grpc_proto{}); err != nil {
		return err
	}
	return x.CloseSend()
}
