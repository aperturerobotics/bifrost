package nats

import (
	"sync"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/nats-io/nats.go"
)

// natsClient contains a nats client with extra data
type natsClient struct {
	*nats.Conn
	npeerID              peer.ID
	npeerIDPretty        string
	subscriptionRefCount int
}

// newNatsClient builds a new nats client.
func newNatsClient(npeerID peer.ID, nc *nats.Conn) *natsClient {
	return &natsClient{npeerID: npeerID, npeerIDPretty: npeerID.Pretty(), Conn: nc}
}

// addRef adds a reference to the nats client.
func (n *natsClient) addRef() func() {
	n.subscriptionRefCount++
	var so sync.Once
	return func() {
		so.Do(func() {
			if n.subscriptionRefCount > 0 {
				n.subscriptionRefCount--
			}
		})
	}
}
