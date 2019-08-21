package udp

import (
	"net"
	"strings"
	"time"

	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/pconn"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// TransportID is the transport identifier
const TransportID = "udp"

// Version is the version of the udp implementation.
var Version = semver.MustParse("0.0.1")

// ExtendedSockBuf is the sockbuf parameter to set on udp sockets.
var ExtendedSockBuf = 16777217

// handshakeTimeout is the time after which a handshake expires
var handshakeTimeout = time.Second * 8

// UDP implements a UDP transport.
type UDP = pconn.Transport

// NewUDP builds a new UDP transport, listening on the addr.
func NewUDP(
	le *logrus.Entry,
	listenAddr string,
	pKey crypto.PrivKey,
	c transport.TransportHandler,
	pconnOpts *pconn.Opts,
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

	uuid := scrc.Crc64([]byte(
		strings.Join([]string{TransportID, listenAddr}, "/"),
	))
	conn := pconn.New(
		le,
		uuid,
		pc,
		pKey,
		func(addr string) (net.Addr, error) {
			return net.ResolveUDPAddr("udp", addr)
		},
		c,
		pconnOpts,
	)
	return conn, nil
}

// _ is a type assertion.
var _ transport.Transport = ((*UDP)(nil))

// _ is a type assertion
var _ transport.TransportDialer = ((*UDP)(nil))
