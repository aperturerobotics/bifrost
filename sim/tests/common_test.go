package tests

import (
	"context"
	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/sirupsen/logrus"
	"testing"
)

var addPeer = AddPeer
var initSimulator = InitSimulator

func TestInitSimulator(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	g := graph.NewGraph()
	addPeer(t, g)
	sim := initSimulator(t, ctx, le, g)
	_ = sim
}
