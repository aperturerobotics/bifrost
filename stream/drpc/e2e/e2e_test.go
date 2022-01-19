package drpc_e2e

import (
	"context"
	"testing"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	stream_drpc_client "github.com/aperturerobotics/bifrost/stream/drpc/client"
	stream_drpc_server "github.com/aperturerobotics/bifrost/stream/drpc/server"
	"github.com/aperturerobotics/bifrost/testbed"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/bifrost/transport/inproc"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"storj.io/drpc/drpcconn"
)

// TestDrpc tests a drpc service end-to-end.
func TestDrpc(t *testing.T) {
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
		TransportPeerId: tb2PeerID.Pretty(),
	})
	go tb2.Bus.ExecuteController(ctx, tp2)

	// tb2 -> tb1 inproc
	tpt2dialer := &dialer.DialerOpts{
		Address: inproc.NewAddr(tb2PeerID).String(),
	}
	tp1 := inproc.BuildInprocController(tb1.Logger, tb1.Bus, tb1PeerID, &inproc.Config{
		TransportPeerId: tb1PeerID.Pretty(),
		Dialers: map[string]*dialer.DialerOpts{
			tb2PeerID.Pretty(): tpt2dialer,
		},
	})
	go tb1.Bus.ExecuteController(ctx, tp1)

	// connect tb1 <-> tb2
	tpt2, _ := tp2.GetTransport(ctx)
	tpt1, _ := tp1.GetTransport(ctx)
	tpt1.(*inproc.Inproc).ConnectToInproc(ctx, tpt2.(*inproc.Inproc))
	tpt2.(*inproc.Inproc).ConnectToInproc(ctx, tpt1.(*inproc.Inproc))

	// tb2: Run the server (stream handler)
	mockServer := NewServer()
	server, err := stream_drpc_server.NewServer(
		tb2.Bus,
		controller.Info{
			Id:          string(ProtocolID) + "/server",
			Version:     "0.0.1",
			Description: "test of drpc server",
		},
		nil,
		[]protocol.ID{ProtocolID},
		[]string{tb2PeerID.Pretty()},
		[]stream_drpc_server.RegisterFn{mockServer.Register},
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	go tb2.Bus.ExecuteController(ctx, server)

	// tb1: construct client
	cl, err := stream_drpc_client.NewClient(tb1.Logger, tb1.Bus, &stream_drpc_client.Config{
		ServerPeerIds: []string{tb2PeerID.Pretty()},
		SrcPeerId:     tb1PeerID.Pretty(),
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// run a request
	mockMsg := "hello world 123"
	err = cl.ExecuteConnection(ctx, ProtocolID, func(conn *drpcconn.Conn) (next bool, err error) {
		client := NewDRPCEndToEndClient(conn)
		resp, err := client.Mock(ctx, &MockRequest{
			Body: mockMsg,
		})
		if err == nil {
			respBody := resp.GetReqBody()
			if respBody != mockMsg {
				err = errors.Errorf("expected response %s but got %s", mockMsg, respBody)
			}
		}
		return false, err
	})
	if err != nil {
		t.Fatal(err.Error())
	}
}
