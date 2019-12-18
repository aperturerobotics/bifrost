package testbed

import (
	"context"

	core "github.com/aperturerobotics/bifrost/core/test"
	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/peer/controller"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	srr "github.com/aperturerobotics/controllerbus/controller/resolver/static"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
)

// Testbed is a constructed testbed.
type Testbed struct {
	// Context is the root context.
	Context context.Context
	// Logger is the logger
	Logger *logrus.Entry
	// StaticResolver is the static resolver.
	StaticResolver *srr.Resolver
	// Bus is the controller bus
	Bus bus.Bus
	// PrivKey is the private key.
	PrivKey crypto.PrivKey
	// Release releases the testbed.
	Release func()
}

// TestbedOpts are extra options to construct the testbed.
type TestbedOpts struct {
	// PrivKey overrides the private key.
	PrivKey crypto.PrivKey
	// NoPeer disables generating + starting the peer and filling PrivKey.
	NoPeer bool
}

// NewTestbed constructs a new core bus with a attached kvtx in-memory volume,
// logger, and other core controllers required for a test to function.
func NewTestbed(ctx context.Context, le *logrus.Entry, opts TestbedOpts) (*Testbed, error) {
	var rels []func()
	t := &Testbed{
		Context: ctx,
		Logger:  le,
		PrivKey: opts.PrivKey,
		Release: func() {
			for _, rel := range rels {
				rel()
			}
		},
	}

	b, sr, err := core.NewTestingBus(ctx, le)
	if err != nil {
		return nil, err
	}
	t.StaticResolver = sr
	t.Bus = b

	if !opts.NoPeer && t.PrivKey == nil {
		t.PrivKey, _, err = keypem.GeneratePrivKey()
		if err != nil {
			return nil, err
		}
	}

	if !opts.NoPeer {
		// start peer controller
		peerConfig, err := peer_controller.NewConfigWithPrivKey(t.PrivKey)
		if err != nil {
			return nil, err
		}
		_, peerRef, err := bus.ExecOneOff(
			ctx,
			t.Bus,
			resolver.NewLoadControllerWithConfig(peerConfig),
			nil,
		)
		if err != nil {
			t.Release()
			return nil, err
		}
		rels = append(rels, peerRef.Release)
	}

	return t, nil
}
