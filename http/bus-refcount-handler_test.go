package bifrost_http

import (
	"context"
	"testing"

	"github.com/aperturerobotics/bifrost/testbed"
	"github.com/sirupsen/logrus"
)

func TestHTTPBusRefCountHandler(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	tb, err := testbed.NewTestbed(ctx, le, testbed.TestbedOpts{
		NoEcho: true,
		NoPeer: true,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer startMockHandler(t, tb)()

	// start the on-demand handler
	rc := NewBusRefCountHandler(ctx, tb.Bus, "/foo", "test-client")

	// perform a request
	checkMockRequest(t, rc)
}
