package common

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/aperturerobotics/bifrost/keypem"
	nctr "github.com/aperturerobotics/bifrost/node/controller"
	"github.com/aperturerobotics/bifrost/peer"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var le = logrus.NewEntry(log)

func init() {
	log.SetLevel(logrus.DebugLevel)
}

func genPeerIdentity() (peer.ID, crypto.PrivKey) {
	pk1, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	pid1, _ := peer.IDFromPrivateKey(pk1)
	log.Debugf("generated peer id: %s", pid1.Pretty())

	return pid1, pk1
}

// BuildCommonBus builds a common bus.
// Also returns a cancel function.
func BuildCommonBus(ctx context.Context) (bus.Bus, crypto.PrivKey, error) {
	peerID, peerPrivKey := genPeerIdentity()

	peerPrivKeyPem, err := keypem.MarshalPrivKeyPem(peerPrivKey)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(peerPrivKeyPem))

	// Construct the bus with the websocket transport and node factory attached.
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		panic(err)
	}
	sr.AddFactory(wtpt.NewFactory(b))
	sr.AddFactory(nctr.NewFactory(b))

	le.
		WithField("peer-id", peerID.Pretty()).
		Debug("constructing node")
	_, _, err = b.AddDirective(
		resolver.NewLoadControllerWithConfigSingleton(&nctr.Config{
			PrivKey: string(peerPrivKeyPem),
		}),
		bus.NewCallbackHandler(func(val directive.Value) {
			le.Infof("node controller resolved: %#v", val)
		}, nil, nil),
	)
	if err != nil {
		return nil, nil, err
	}

	return b, peerPrivKey, nil
}

// GetLogEntry returns the root log entry.
func GetLogEntry() *logrus.Entry {
	return le
}
