package udp

import (
	"context"
	"net"

	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/pconn"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

// TransportID is the transport identifier
const TransportID = "udp"

// Version is the version of the udp implementation.
var Version = semver.MustParse("0.0.1")

// ExtendedSockBuf is the sockbuf parameter to set on udp sockets.
var ExtendedSockBuf = 16777217

// UDP implements a UDP transport.
type UDP = pconn.Transport

// NewUDP builds a new UDP transport, listening on the addr.
//
// uuid can be empty to generate based on peer id and local addr
func NewUDP(
	ctx context.Context,
	le *logrus.Entry,
	pKey crypto.PrivKey,
	c transport.TransportHandler,
	pconnOpts *pconn.Opts,
	uuid uint64,
	listenAddr string,
) (*UDP, error) {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return nil, err
	}
	if uc, ok := pc.(*net.UDPConn); ok {
		if err := uc.SetReadBuffer(ExtendedSockBuf); err != nil {
			le.WithError(err).Warn("unable to set read buffer on conn")
		}
		if err := uc.SetWriteBuffer(ExtendedSockBuf); err != nil {
			le.WithError(err).Warn("unable to set write buffer on conn")
		}
	}

	return pconn.NewTransport(
		ctx,
		le,
		pKey,
		c,
		pconnOpts,
		uuid,
		pc,
		func(addr string) (net.Addr, error) {
			return net.ResolveUDPAddr("udp", addr)
		},
	)
}

// _ is a type assertion.
var _ transport.Transport = ((*UDP)(nil))

// _ is a type assertion
var _ transport.TransportDialer = ((*UDP)(nil))
