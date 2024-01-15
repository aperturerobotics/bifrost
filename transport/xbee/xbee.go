package xbee

import (
	"context"
	"net"
	"strings"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	kcp "github.com/aperturerobotics/bifrost/transport/common/kcp"
	"github.com/aperturerobotics/bifrost/transport/xbee/xbserial"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

// Version is the version of the xbee implementation.
var Version = semver.MustParse("0.0.1")

// TransportType is the transport type identifier for this transport.
const TransportType = "xbee"

// ControllerID is the xbee controller ID.
const ControllerID = "bifrost/xbee"

// Link represents a xbee-based connection/link.
type Link = kcp.Link

// XBee implements a XBee transport.
type XBee struct {
	*kcp.Transport

	xbs *xbserial.XBeeSerial
	// staticPeerMap is the static peer map
	staticPeerMap map[string]*dialer.DialerOpts
}

// NewXBee builds a new XBee transport, opening the serial device.
func NewXBee(
	ctx context.Context,
	le *logrus.Entry,
	opts *Config,
	pKey crypto.PrivKey,
	c transport.TransportHandler,
) (*XBee, error) {
	le = le.
		WithField("device-path", opts.GetDevicePath()).
		WithField("device-baud", opts.GetDeviceBaud())
	le.Debug("opening device")
	sp, err := serial.Open(opts.GetDevicePath(), &serial.Mode{
		BaudRate: int(opts.GetDeviceBaud()),
	})
	if err != nil {
		return nil, errors.Wrap(err, "open xbee serial")
	}

	xbs := xbserial.NewXBeeSerial(le, sp)
	go func() {
		_ = xbs.ReadPump()
	}()

	le.Debug("reading device address")
	pc, err := xbserial.NewPacketConn(ctx, xbs)
	if err != nil {
		sp.Close()
		return nil, err
	}

	uuid := scrc.Crc64([]byte(
		strings.Join([]string{ControllerID, pc.LocalAddr().String()}, "/"),
	))
	le.
		WithField("device-address", pc.LocalAddr().String()).
		Info("opened xbee device successfully")
	conn := kcp.New(
		le,
		uuid,
		pc,
		pKey,
		func(addr string) (net.Addr, error) {
			return xbserial.ParseXBeeAddr(addr)
		},
		c,
		opts.GetPacketOpts(),
	)
	return &XBee{Transport: conn, xbs: xbs, staticPeerMap: opts.GetDialers()}, nil
}

// MatchTransportType checks if the given transport type ID matches this transport.
// If returns true, the transport controller will call DialPeer with that tptaddr.
// E.x.: "udp-quic" or "ws"
func (c *XBee) MatchTransportType(transportType string) bool {
	return transportType == TransportType
}

// GetPeerDialer returns the dialing information for a peer.
// Called when resolving EstablishLink.
// Return nil, nil to indicate not found or unavailable.
func (c *XBee) GetPeerDialer(ctx context.Context, peerID peer.ID) (*dialer.DialerOpts, error) {
	return c.staticPeerMap[peerID.String()], nil
}

// Execute executes the transport as configured, returning any fatal error.
func (c *XBee) Execute(ctx context.Context) error {
	return c.Transport.Execute(ctx)
}

var (
	// _ is a type assertion.
	_ transport.Transport = ((*XBee)(nil))

	// _ is a type assertion
	_ dialer.TransportDialer = ((*XBee)(nil))
)
