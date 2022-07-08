package transport_quic

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	p2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	quic "github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
)

// BuildQuicConfig constructs the quic config.
func BuildQuicConfig(le *logrus.Entry, opts *Opts) *quic.Config {
	var enableMaxIdleTimeout bool
	maxIdleTimeout := time.Second * 40
	if ntDur := opts.GetMaxIdleTimeoutDur(); ntDur != "" {
		nt, err := time.ParseDuration(ntDur)
		if err == nil && nt != time.Duration(0) && nt < time.Hour*2 {
			maxIdleTimeout = nt
			enableMaxIdleTimeout = true
		}
	}

	maxIncStreams := 100000
	if mis := opts.GetMaxIncomingStreams(); mis <= 0 {
		maxIncStreams = int(mis)
	}

	if opts.GetVerbose() {
		le.Level = logrus.DebugLevel
	} else {
		le.Level = logrus.InfoLevel
	}

	return &quic.Config{
		Logger: le,

		EnableDatagrams:         !opts.GetDisableDatagrams(),
		KeepAlive:               !opts.GetDisableKeepAlive(),
		DisablePathMTUDiscovery: opts.GetDisablePathMtuDiscovery(),

		MaxIdleTimeout:        maxIdleTimeout,
		DisableIdleTimeout:    !enableMaxIdleTimeout,
		MaxIncomingStreams:    int64(maxIncStreams),
		MaxIncomingUniStreams: -1, // disable unidirectional streams

		AcceptToken: func(clientAddr net.Addr, _ *quic.Token) bool {
			// unconditionally accept any quic token
			return true
		},

		// MaxReceiveStreamFlowControlWindow:     3 * (1 << 20),   // 3 MB
		// MaxReceiveConnectionFlowControlWindow: 4.5 * (1 << 20), // 4.5 MB
	}
}

// BuildIncomingTlsConf builds the tls config for incoming conns.
//
// rpeer can be empty to indicate accepting any remote peer
func BuildIncomingTlsConf(identity *p2ptls.Identity, rpeer peer.ID) *tls.Config {
	var tlsConf tls.Config
	tlsConf.NextProtos = []string{Alpn}
	tlsConf.GetConfigForClient = func(_ *tls.ClientHelloInfo) (*tls.Config, error) {
		// note: if rpeer is empty, allows any incoming peer id.
		conf, _ := identity.ConfigForPeer(rpeer)
		conf.NextProtos = []string{Alpn}
		return conf, nil
	}
	return &tlsConf
}
