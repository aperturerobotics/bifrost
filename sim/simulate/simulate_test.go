package simulate

import (
	"context"
	"testing"

	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/sirupsen/logrus"
)

// TestSimpleSimulate tests a simple simulation.
func TestSimpleSimulate(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	g := graph.NewGraph()

	addPeer := func() *graph.Peer {
		p, err := graph.GenerateAddPeer(ctx, g)
		if err != nil {
			t.Fatal(err.Error())
		}
		return p
	}

	p0 := addPeer()
	p1 := addPeer()
	lan1 := graph.AddLAN(g)
	lan1.AddPeer(g, p0)
	lan1.AddPeer(g, p1)

	p2 := addPeer()
	lan2 := graph.AddLAN(g)
	lan2.AddPeer(g, p2)
	lan2.AddConnectionToLAN(g, lan1)

	sim, err := NewSimulator(ctx, le, g)
	if err != nil {
		t.Fatal(err.Error())
	}
	le.Info("simulator startup complete")

	assertConnectivity := func(p0, p1 *graph.Peer) {
		px0 := sim.GetPeerByID(p0.GetPeerID())
		px1 := sim.GetPeerByID(p1.GetPeerID())
		if err := TestConnectivity(ctx, px0, px1); err != nil {
			t.Fatal(err.Error())
		}
		le.Infof(
			"successful connectivity test between %s and %s",
			p0.GetPeerID().Pretty(),
			p1.GetPeerID().Pretty(),
		)
	}

	assertConnectivity(p0, p1)
	assertConnectivity(p1, p0)
	assertConnectivity(p2, p0)

	le.Info("tests successful")
	_ = sim
}
