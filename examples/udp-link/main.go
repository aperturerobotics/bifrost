package main

import (
	"context"
	"crypto/rand"
	"net"
	"sync"

	"github.com/aperturerobotics/bifrost/daemon"
	link_holdopen_controller "github.com/aperturerobotics/bifrost/link/hold-open"
	"github.com/aperturerobotics/bifrost/peer"
	tptc "github.com/aperturerobotics/bifrost/transport/controller"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.New()
)

func init() {
	log.SetLevel(logrus.DebugLevel)
}

func genPeerIdentity() (peer.ID, crypto.PrivKey) {
	pk1, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	pid1, _ := peer.IDFromPrivateKey(pk1)
	log.Debugf("generated peer id: %s", pid1.String())

	return pid1, pk1
}

func execute() error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	_, pk1 := genPeerIdentity()
	p2, pk2 := genPeerIdentity()

	d1, err := daemon.NewDaemon(ctx, pk1, daemon.ConstructOpts{
		LogEntry: le,
	})
	if err != nil {
		return errors.Wrap(err, "construct daemon 1")
	}

	d2, err := daemon.NewDaemon(ctx, pk2, daemon.ConstructOpts{
		LogEntry: le,
	})
	if err != nil {
		return errors.Wrap(err, "construct daemon 2")
	}

	bus1 := d1.GetControllerBus()
	bus2 := d2.GetControllerBus()
	sr1 := d1.GetStaticResolver()
	sr2 := d2.GetStaticResolver()

	var wg sync.WaitGroup
	wg.Add(2)

	// Execute hold-open
	sr1.AddFactory(link_holdopen_controller.NewFactory(bus1))
	_, _, hr1, err := loader.WaitExecControllerRunning(
		ctx,
		bus1,
		resolver.NewLoadControllerWithConfig(&link_holdopen_controller.Config{}),
		nil,
	)
	if err != nil {
		return err
	}
	defer hr1.Release()

	sr2.AddFactory(link_holdopen_controller.NewFactory(bus2))
	_, _, hr2, err := loader.WaitExecControllerRunning(
		ctx,
		bus2,
		resolver.NewLoadControllerWithConfig(&link_holdopen_controller.Config{}),
		nil,
	)
	if err != nil {
		return err
	}
	defer hr2.Release()

	// Execute the UDP transport on the first daemon.
	tptc1, _, udpRef1, err := loader.WaitExecControllerRunning(
		ctx,
		bus1,
		resolver.NewLoadControllerWithConfig(&udptpt.Config{
			ListenAddr: ":5553",
		}),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "listen on udp 1")
	}
	defer udpRef1.Release()
	le.Info("UDP listening on: :5553")
	tpt1, _ := tptc1.(*tptc.Controller).GetTransport(ctx)

	// Execute the UDP transport on the second daemon.
	tptc2, _, udpRef2, err := loader.WaitExecControllerRunning(
		ctx,
		bus2,
		resolver.NewLoadControllerWithConfig(&udptpt.Config{
			ListenAddr: ":5554",
		}),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "listen on udp 2")
	}
	defer udpRef2.Release()
	le.Info("UDP listening on: :5554")
	_, _ = tptc2.(*tptc.Controller).GetTransport(ctx)

	tpt1.(*udptpt.UDP).DialPeer(ctx, p2, (&net.UDPAddr{
		IP:   net.IP{127, 0, 0, 1},
		Port: 5554,
	}).String())
	<-ctx.Done()
	return nil
}

func main() {
	if err := execute(); err != nil {
		panic(err)
	}
}
