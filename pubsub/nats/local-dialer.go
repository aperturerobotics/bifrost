package nats

import (
	"net"

	nats_client "github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
	"github.com/pkg/errors"
)

const localNatsAddress = "local:4222"

// localNatsDialer dials the local nats server in-process.
type localNatsDialer struct {
	n  *Nats
	kp nkeys.KeyPair
}

// newLocalNatsDialer constructs a new localNatsDialer.
func newLocalNatsDialer(n *Nats, keyPair nkeys.KeyPair) *localNatsDialer {
	return &localNatsDialer{n: n, kp: keyPair}
}

// Dial dials the local server.
func (d *localNatsDialer) Dial(network, address string) (net.Conn, error) {
	if address != localNatsAddress {
		return nil, errors.Errorf("unexpected address for local nats dialer: %s", address)
	}

	var pubKey string
	var err error
	if d.kp != nil {
		pubKey, err = d.kp.PublicKey()
		if err != nil {
			return nil, err
		}
	}

	p1, p2 := net.Pipe()
	go func() {
		_ = d.n.natsServer.HandleClientConnection(p2, pubKey)
	}()
	return p1, nil
}

// _ is a type assertion
var _ nats_client.CustomDialer = ((*localNatsDialer)(nil))
