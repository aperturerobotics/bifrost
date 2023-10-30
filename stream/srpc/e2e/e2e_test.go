package srpc_e2e

import (
	"context"
	"testing"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	stream_srpc_client "github.com/aperturerobotics/bifrost/stream/srpc/client"
	stream_srpc_server "github.com/aperturerobotics/bifrost/stream/srpc/server"
	"github.com/aperturerobotics/bifrost/testbed"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/bifrost/transport/inproc"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/starpc/echo"
	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var ProtocolID = protocol.ID("bifrost/stream/srpc/e2e")

// TestStarpc tests a srpc service end-to-end.
func TestStarpc(t *testing.T) {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	// tb1: client
	tb1, err := testbed.NewTestbed(ctx, le.WithField("testbed", 1), testbed.TestbedOpts{NoEcho: true})
	if err != nil {
		t.Fatal(err.Error())
	}

	// tb2: server
	tb2, err := testbed.NewTestbed(ctx, le.WithField("testbed", 2), testbed.TestbedOpts{})
	if err != nil {
		t.Fatal(err.Error())
	}

	tb1PeerID, err := peer.IDFromPrivateKey(tb1.PrivKey)
	if err != nil {
		t.Fatal(err.Error())
	}

	tb2PeerID, err := peer.IDFromPrivateKey(tb2.PrivKey)
	if err != nil {
		t.Fatal(err.Error())
	}

	// tb1 -> tb2 inproc
	tp2 := inproc.BuildInprocController(tb2.Logger, tb2.Bus, tb2PeerID, &inproc.Config{
		TransportPeerId: tb2PeerID.String(),
	})
	go func() {
		_ = tb2.Bus.ExecuteController(ctx, tp2)
	}()

	// tb2 -> tb1 inproc
	tpt2dialer := &dialer.DialerOpts{
		Address: inproc.NewAddr(tb2PeerID).String(),
	}
	tp1 := inproc.BuildInprocController(tb1.Logger, tb1.Bus, tb1PeerID, &inproc.Config{
		TransportPeerId: tb1PeerID.String(),
		Dialers: map[string]*dialer.DialerOpts{
			tb2PeerID.String(): tpt2dialer,
		},
	})
	go func() {
		_ = tb1.Bus.ExecuteController(ctx, tp1)
	}()

	// connect tb1 <-> tb2
	tpt2, _ := tp2.GetTransport(ctx)
	tpt1, _ := tp1.GetTransport(ctx)
	tpt1.(*inproc.Inproc).ConnectToInproc(ctx, tpt2.(*inproc.Inproc))
	tpt2.(*inproc.Inproc).ConnectToInproc(ctx, tpt1.(*inproc.Inproc))

	// tb2: Run the server (stream handler)
	mockServer := echo.NewEchoServer(nil)
	server, err := stream_srpc_server.NewServer(
		tb2.Bus,
		le,
		controller.NewInfo(
			string(ProtocolID)+"/server",
			semver.MustParse("0.0.1"),
			"test of srpc server",
		),
		[]protocol.ID{ProtocolID},
		[]string{tb2PeerID.String()},
		[]stream_srpc_server.RegisterFn{mockServer.Register},
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	go func() {
		_ = tb2.Bus.ExecuteController(ctx, server)
	}()

	// tb1: construct client
	cl, err := stream_srpc_client.NewClient(tb1.Logger, tb1.Bus, &stream_srpc_client.Config{
		ServerPeerIds: []string{tb2PeerID.String()},
		SrcPeerId:     tb1PeerID.String(),
	}, ProtocolID)
	if err != nil {
		t.Fatal(err.Error())
	}

	// run a request
	mockMsg := "hello world 123"
	serviceClient := echo.NewSRPCEchoerClient(cl)
	resp, err := serviceClient.Echo(ctx, &echo.EchoMsg{
		Body: mockMsg,
	})
	if err == nil {
		respBody := resp.GetBody()
		if respBody != mockMsg {
			err = errors.Errorf("expected response %s but got %s", mockMsg, respBody)
		}
	}
	if err != nil {
		t.Fatal(err.Error())
	}
}
