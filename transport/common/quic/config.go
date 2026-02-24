package transport_quic

import (
	"crypto/tls"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	p2ptls "github.com/aperturerobotics/bifrost/crypto/tls"
	quic "github.com/quic-go/quic-go"
)

// BuildQuicConfig constructs the quic config.
func BuildQuicConfig(opts *Opts) *quic.Config {
	maxIdleTimeout := time.Second * 10
	if ntDur := opts.GetMaxIdleTimeoutDur(); ntDur != "" {
		nt, err := time.ParseDuration(ntDur)
		if err == nil && nt > time.Duration(0) && nt < time.Hour*2 {
			maxIdleTimeout = nt
		}
	}

	maxIncStreams := 100000
	if mis := opts.GetMaxIncomingStreams(); mis > 0 {
		maxIncStreams = int(mis)
	}

	keepAlivePeriod := maxIdleTimeout / 2
	if opts.GetDisableKeepAlive() {
		keepAlivePeriod = 0
	} else if keepAliveDur := opts.GetKeepAliveDur(); keepAliveDur != "" {
		kaDur, err := time.ParseDuration(keepAliveDur)
		if err == nil && kaDur > time.Duration(0) && kaDur < time.Hour*2 {
			keepAlivePeriod = kaDur
		}
	}

	return &quic.Config{
		// We don't use datagrams (yet), but this is necessary for WebTransport
		EnableDatagrams:         !opts.GetDisableDatagrams(),
		KeepAlivePeriod:         keepAlivePeriod,
		DisablePathMTUDiscovery: opts.GetDisablePathMtuDiscovery(),

		MaxIdleTimeout:        maxIdleTimeout,
		MaxIncomingStreams:    int64(maxIncStreams),
		MaxIncomingUniStreams: -1, // disable unidirectional streams

		Versions: []quic.Version{quic.Version2},
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
		// TODO: https://github.com/golang/go/issues/60506
		conf.SessionTicketsDisabled = true
		return conf, nil
	}

	// TODO: https://github.com/golang/go/issues/60506
	tlsConf.SessionTicketsDisabled = true
	return &tlsConf
}
