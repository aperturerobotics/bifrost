package peer_controller

import (
	"context"
	"testing"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/sirupsen/logrus"
)

func TestPrivKeyIntegrity(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	npeer, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	ctx := context.Background()
	privKey, err := npeer.GetPrivKey(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	privKeyID := npeer.GetPeerID()
	peerControllerConf, err := NewConfigWithPrivKey(privKey)
	if err != nil {
		t.Fatal(err.Error())
	}

	f := NewFactory()
	ctrl, err := f.Construct(peerControllerConf, controller.ConstructOpts{
		Logger: le,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	cctrl := ctrl.(*Controller)
	if privKeyID.Pretty() != cctrl.GetPeerID().Pretty() {
		t.Fatalf("priv key id mismatch: %s != %s", privKeyID.Pretty(), cctrl.GetPeerID().Pretty())
	}
}
