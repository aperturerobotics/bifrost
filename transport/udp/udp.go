package udp

import (
	"context"
	"net"
	"time"

	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/pconn"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// TransportID is the transport identifier
const TransportID = "udp"

// Version is the version of the udp implementation.
var Version = semver.MustParse("0.0.1")

// handshakeTimeout is the time after which a handshake expires
var handshakeTimeout = time.Second * 8

// UDP implements a UDP transport.
// It is unordered, unreliable, and unencrypted.
type UDP = pconn.Transport

// NewUDP builds a new UDP transport, listening on the addr.
func NewUDP(
	le *logrus.Entry,
	listenAddr string,
	dialAddrs []string,
	pKey crypto.PrivKey,
	c transport.TransportHandler,
) (*UDP, error) {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	conn := pconn.New(le, pc, pKey, c)
	for _, addr := range dialAddrs {
		da, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			le.WithError(err).Warnf("cannot resolve address: %s", addr)
			continue
		}

		go conn.Dial(context.TODO(), da)
	}

	return conn, nil
}

// _ is a type assertion.
var _ transport.Transport = ((*UDP)(nil))
