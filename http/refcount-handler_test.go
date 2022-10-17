package bifrost_http

import (
	"context"
	"net/http"
	"testing"

	"github.com/aperturerobotics/bifrost/testbed"
	"github.com/sirupsen/logrus"
)

func TestHTTPRefCountHandler(t *testing.T) {
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
	rc := NewRefCountHandler(ctx, func(ctx context.Context) (*http.Handler, func(), error) {
		busHandler := NewBusHandler(tb.Bus, "test-client")
		httpHandler := http.Handler(busHandler)
		return &httpHandler, nil, nil
	})

	// perform a request
	checkMockRequest(t, rc)
}
