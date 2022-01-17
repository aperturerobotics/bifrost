package stream_drpc_server

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream/drpc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"

	"storj.io/drpc"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
)

// RegisterFn is a callback to register apis to a mux.
type RegisterFn func(mux drpc.Mux) error

// Server handles incoming streams for a peer id.
type Server struct {
	// b is the bus
	b bus.Bus
	// drpcOpts are opts passed to drpc
	drpcOpts *stream_drpc.DrpcOpts
	// protocolIDs is list of protocol id to listen on.
	// cannot be empty
	protocolIDs []protocol.ID
	// peerIDs is the list of local peer ids to listen on
	// if empty, allows any
	peerIDs []string

	// mux is the drpc mux
	mux *drpcmux.Mux
	// server is the drpc server
	server *drpcserver.Server
}

// NewServer constructs a common drpc controller.
// If peerIDs and/or domainIDs are empty, matches any.
func NewServer(
	b bus.Bus,
	opts *stream_drpc.DrpcOpts,
	protocolIDs []protocol.ID,
	peerIDs []string,
	registerFns []RegisterFn,
) (*Server, error) {
	mux := drpcmux.New()
	for _, rf := range registerFns {
		if err := rf(mux); err != nil {
			return nil, err
		}
	}

	server := drpcserver.New(mux)
	return &Server{
		b:           b,
		drpcOpts:    opts,
		protocolIDs: protocolIDs,
		peerIDs:     peerIDs,

		mux:    mux,
		server: server,
	}, nil
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
func (s *Server) HandleDirective(ctx context.Context, di directive.Instance) (directive.Resolver, error) {
	dir := di.GetDirective()
	switch d := dir.(type) {
	case link.HandleMountedStream:
		return s.ResolveHandleMountedStream(ctx, di, d)
	}

	return nil, nil
}

// ResolveHandleMountedStream resolves a HandleMountedStream directive by dialing a target.
func (s *Server) ResolveHandleMountedStream(
	ctx context.Context,
	di directive.Instance,
	dir link.HandleMountedStream,
) (directive.Resolver, error) {
	inProtocol := dir.HandleMountedStreamProtocolID()
	var match bool
	for _, pr := range s.protocolIDs {
		if pr == inProtocol {
			match = true
			break
		}
	}
	if !match {
		return nil, nil
	}

	inPeerID := dir.HandleMountedStreamLocalPeerID()
	inPeerIDPretty := inPeerID.Pretty()
	if len(s.peerIDs) != 0 {
		match = false
		for _, pid := range s.peerIDs {
			if pid == inPeerIDPretty {
				match = true
			}
		}
		if !match {
			return nil, nil
		}
	}

	return newMountedStreamResolver(s), nil
}

// HandleMountedStream handles an incoming mounted stream.
// Any returned error indicates the stream should be closed.
// This function should return as soon as possible, and start
// additional goroutines to manage the lifecycle of the stream.
// Typically EstablishLink is asserted in HandleMountedStream.
func (s *Server) HandleMountedStream(ctx context.Context, ms link.MountedStream) error {
	// keep the link open
	_, elRef, err := s.b.AddDirective(
		link.NewEstablishLinkWithPeer(ms.GetLink().GetLocalPeer(), ms.GetPeerID()),
		nil,
	)
	if err != nil {
		return err
	}
	go func() {
		subCtx, subCtxCancel := context.WithCancel(ctx)
		strm := ms.GetStream()
		_ = s.server.ServeOne(subCtx, strm)
		subCtxCancel()
		strm.Close()
		elRef.Release()
	}()
	return nil
}

// _ is a type assertion
var _ link.MountedStreamHandler = ((*Server)(nil))
