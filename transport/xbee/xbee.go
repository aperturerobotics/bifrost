package xbee

import (
	"context"
	"net"
	"strings"

	"github.com/aperturerobotics/bifrost/transport"
	pconn "github.com/aperturerobotics/bifrost/transport/common/kcp"
	"github.com/aperturerobotics/bifrost/transport/xbee/xbserial"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

// TransportID is the transport identifier
const TransportID = "xbee"

// Version is the version of the xbee implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the xbee controller ID.
const ControllerID = "bifrost/xbee/1"

// Link represents a xbee-based connection/link.
type Link = pconn.Link

// XBee implements a XBee transport.
type XBee struct {
	*pconn.Transport

	xbs *xbserial.XBeeSerial
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
	sp, err := serial.OpenPort(&serial.Config{
		Name: opts.GetDevicePath(),
		Baud: int(opts.GetDeviceBaud()),
		// ReadTimeout: time.Millisecond * 500,
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
		strings.Join([]string{TransportID, pc.LocalAddr().String()}, "/"),
	))
	le.
		WithField("device-address", pc.LocalAddr().String()).
		Info("opened xbee device successfully")
	conn := pconn.New(
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
	return &XBee{Transport: conn, xbs: xbs}, nil
}

// Execute executes the transport as configured, returning any fatal error.
func (c *XBee) Execute(ctx context.Context) error {
	return c.Transport.Execute(ctx)
}

// _ is a type assertion.
var _ transport.Transport = ((*XBee)(nil))

// _ is a type assertion
var _ transport.TransportDialer = ((*XBee)(nil))
