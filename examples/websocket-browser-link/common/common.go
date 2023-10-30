package common

import (
	"context"
	"fmt"

	"github.com/aperturerobotics/bifrost/keypem"
	link_holdopen_controller "github.com/aperturerobotics/bifrost/link/hold-open"
	"github.com/aperturerobotics/bifrost/peer"
	nctr "github.com/aperturerobotics/bifrost/peer/controller"
	"github.com/aperturerobotics/bifrost/pubsub/nats"
	nats_controller "github.com/aperturerobotics/bifrost/pubsub/nats/controller"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/prng"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var le = logrus.NewEntry(log)

func init() {
	log.SetLevel(logrus.DebugLevel)
}

func genPeerIdentity(peerSeed string) (peer.ID, crypto.PrivKey) {
	// NOT SECURE: for demo purposes only
	r := prng.BuildSeededRand([]byte(peerSeed))
	pk1, _, err := crypto.GenerateEd25519Key(r)
	if err != nil {
		log.Fatal(err)
	}
	pid1, _ := peer.IDFromPrivateKey(pk1)
	log.Debugf("generated insecure peer id: %s", pid1.String())

	return pid1, pk1
}

// BuildCommonBus builds a common bus.
// Also returns a cancel function.
func BuildCommonBus(ctx context.Context, peerSeed string) (bus.Bus, crypto.PrivKey, error) {
	peerID, peerPrivKey := genPeerIdentity(peerSeed)

	peerPrivKeyPem, err := keypem.MarshalPrivKeyPem(peerPrivKey)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println(string(peerPrivKeyPem))

	// Construct the bus with the websocket transport and node factory attached.
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return nil, nil, err
	}
	sr.AddFactory(wtpt.NewFactory(b))
	sr.AddFactory(nctr.NewFactory())
	sr.AddFactory(link_holdopen_controller.NewFactory(b))
	sr.AddFactory(nats_controller.NewFactory(b))

	le = le.WithField("peer-id", peerID.String())
	le.Debug("constructing node")
	_, _, err = b.AddDirective(
		resolver.NewLoadControllerWithConfig(&nctr.Config{
			PrivKey: string(peerPrivKeyPem),
		}),
		bus.NewCallbackHandler(func(val directive.AttachedValue) {
			le.Debug("node controller resolved")
		}, nil, nil),
	)
	if err != nil {
		return nil, nil, err
	}

	// keep links open
	holdOpen, err := link_holdopen_controller.NewController(b, le)
	if err != nil {
		return nil, nil, err
	}
	_, err = b.AddController(ctx, holdOpen, nil)
	if err != nil {
		return nil, nil, err
	}

	// use pubsub: nats
	_, _, err = b.AddDirective(resolver.NewLoadControllerWithConfig(&nats_controller.Config{
		PeerId:     peerID.String(),
		NatsConfig: &nats.Config{LogTrace: true},
	}), nil)
	if err != nil {
		return nil, nil, err
	}

	return b, peerPrivKey, nil
}

// GetLogEntry returns the root log entry.
func GetLogEntry() *logrus.Entry {
	return le
}
