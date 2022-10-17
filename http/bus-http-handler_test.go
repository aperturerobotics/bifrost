package bifrost_http

import (
	"context"
	"testing"

	"github.com/aperturerobotics/bifrost/testbed"
	"github.com/sirupsen/logrus"
)

func TestHTTPBusHTTPHandler(t *testing.T) {
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
	rc := NewBusHTTPHandler(ctx, tb.Bus, "/foo", "test-client", false)

	// perform a request
	checkMockRequest(t, rc)
}
