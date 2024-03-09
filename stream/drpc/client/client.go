package stream_drpc_client

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	stream_drpc "github.com/aperturerobotics/bifrost/stream/drpc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"storj.io/drpc/drpcconn"
)

// Client is a common drpc client implementation.
type Client struct {
	// le is the logger
	le *logrus.Entry
	// b is the bus
	b bus.Bus
	// c is the config
	c *Config

	// timeoutDur is the request timeout duration
	// if unset, has none
	timeoutDir time.Duration
	// srcPeer is the src peer id
	srcPeer peer.ID
	// destPeers is the dest peer ids
	destPeers []peer.ID
}

// NewClient constructs a new client.
func NewClient(le *logrus.Entry, b bus.Bus, c *Config) (*Client, error) {
	srcPeer, err := c.ParseSrcPeerId()
	if err != nil {
		return nil, errors.Wrap(err, "src_peer_id")
	}

	serverPeerIDs, err := c.ParseServerPeerIds()
	if err != nil {
		return nil, errors.Wrap(err, "src_peer_id")
	}

	timeoutDur, err := c.ParseTimeoutDur()
	if err != nil {
		return nil, errors.Wrap(err, "timeout_dur")
	}

	return &Client{
		le: le,
		b:  b,
		c:  c,

		timeoutDir: timeoutDur,
		srcPeer:    srcPeer,
		destPeers:  serverPeerIDs,
	}, nil
}

// BuildTimeoutCtx builds a context with the configured timeout.
func (c *Client) BuildTimeoutCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	to := c.timeoutDir
	if to <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, to)
}

// ExecuteConnection attempts to contact one of the configured servers and
// execute the given callback, which should construct & use drpc clients.
//
// Callback should return nextServer, err.
// If next=true is returned, tries another server.
func (c *Client) ExecuteConnection(
	ctx context.Context,
	protocolID protocol.ID,
	cb func(conn *drpcconn.Conn) (next bool, err error),
) error {
	var lastErr error
	for _, destPeer := range c.destPeers {
		estCtx, estCtxCancel := c.BuildTimeoutCtx(ctx)
		defer estCtxCancel()

		le := c.le.WithField("server-peer-id", destPeer.String())
		conn, connRel, err := stream_drpc.EstablishDrpcConn(
			estCtx,
			c.b,
			c.c.GetDrpcOpts(),
			protocolID,
			c.srcPeer, destPeer,
			c.c.GetTransportId(),
		)
		if err != nil {
			// detect deadline exceeded
			if err == context.Canceled && estCtx.Err() != nil && ctx.Err() == nil {
				err = context.DeadlineExceeded
			}
			le.WithError(err).Warn("unable to establish drpc conn")
			lastErr = err
			continue
		}

		var tryNext bool
		var tryErr error
		func() {
			// this also catches panic cases.
			defer connRel()
			defer conn.Close()

			tryNext, tryErr = cb(conn)
		}()
		if tryErr == nil {
			return nil
		}

		lastErr = tryErr
		if !tryNext {
			break
		}
	}

	if lastErr == nil {
		lastErr = errors.New("connection failed")
	}

	return lastErr
}
