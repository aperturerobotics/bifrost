package webrtc

import (
	"context"
	"errors"
	"testing"

	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	pion "github.com/pion/webrtc/v4"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/s4wave/spacewave/net/testbed"
	"github.com/sirupsen/logrus"
)

// testICEProvider serves mutable credentials to one test peer.
type testICEProvider struct {
	// peerID restricts the provider to the intended local identity.
	peerID peer.ID
	// password is the next credential returned during a synchronous request.
	password string
	// err refuses the next credential request when set.
	err error
}

// GetControllerInfo identifies the fixture adapter.
func (p *testICEProvider) GetControllerInfo() *controller.Info {
	return controller.NewInfo("test/ice", controller.MustParseVersion("0.0.1"), "ICE test")
}

// Execute leaves the fixture attached until cleanup.
func (p *testICEProvider) Execute(context.Context) error { return nil }

// Close releases no external resources.
func (p *testICEProvider) Close() error { return nil }

// HandleDirective confines credentials to the fixture peer.
func (p *testICEProvider) HandleDirective(_ context.Context, inst directive.Instance) ([]directive.Resolver, error) {
	d, ok := inst.GetDirective().(LookupICEProvider)
	if !ok || d.ICEPeerID() != p.peerID {
		return nil, nil
	}
	return directive.Resolvers(directive.NewValueResolver([]ICEProvider{p})), nil
}

// ICEConfig returns the current credential or a deliberate refusal.
func (p *testICEProvider) ICEConfig(context.Context) (*WebRtcConfig, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &WebRtcConfig{IceServers: []*IceServerConfig{{
		Urls: []string{"turn:127.0.0.1:3478"}, Username: "tester",
		Credential: &IceServerConfig_Password{Password: p.password},
	}}}, nil
}

// TestICEProviderRefreshAndIsolation exercises the real Pion construction boundary.
func TestICEProviderRefreshAndIsolation(t *testing.T) {
	// Attach an application provider to the actual controller bus.
	tb, err := testbed.NewTestbed(t.Context(), logrus.NewEntry(logrus.New()), testbed.TestbedOpts{NoEcho: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tb.Release)
	provider := &testICEProvider{peerID: tb.PeerID, password: "first"}
	release, err := tb.Bus.AddController(t.Context(), provider, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(release)
	w := &WebRTC{b: tb.Bus, peerID: tb.PeerID, webrtcApi: pion.NewAPI(), webrtcConf: &pion.Configuration{}}

	// Every new connection uses fresh credentials without mutating the prior one.
	first, err := w.newPeerConnection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Close() })
	provider.password = "second"
	second, err := w.newPeerConnection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = second.Close() })
	if first.GetConfiguration().ICEServers[0].Credential != "first" || second.GetConfiguration().ICEServers[0].Credential != "second" {
		t.Fatal("credentials were retained across independent connection attempts")
	}

	// A provider refusal must not silently fall back to direct-only transport.
	provider.err = errors.New("credentials unavailable")
	if conn, err := w.newPeerConnection(t.Context()); err == nil || conn != nil {
		t.Fatal("credential refusal created a connection")
	}

	// Other session identities retain their configured transport without this provider.
	other, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err)
	}
	w.peerID = other.GetPeerID()
	conn, err := w.newPeerConnection(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if len(conn.GetConfiguration().ICEServers) != 0 {
		t.Fatal("another peer inherited application credentials")
	}
}

// _ is a type assertion
var _ controller.Controller = (*testICEProvider)(nil)
