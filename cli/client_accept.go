package cli

import (
	"os"

	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/bifrost/stream/grpc/rpc"
	"github.com/aperturerobotics/bifrost/util/rwc"
	"github.com/urfave/cli"
)

// RunAccept runs the accept command.
func (a *ClientArgs) RunAccept(*cli.Context) error {
	ctx := a.GetContext()
	c, err := a.BuildClient()
	if err != nil {
		return err
	}

	client, err := c.AcceptStream(ctx)
	if err != nil {
		return err
	}

	if len(a.RemotePeerIdsCsv) != 0 {
		a.AcceptConf.RemotePeerIds = a.ParseRemotePeerIdsCsv()
	}
	err = client.Send(&stream_grpc.AcceptStreamRequest{
		Config: &a.AcceptConf,
	})
	if err != nil {
		return err
	}

	drpc := stream_grpc.NewAcceptStreamClientRPC(client)
	return stream_grpc_rpc.AttachRPCToStream(
		drpc,
		rwc.NewReadWriteCloser(os.Stdin, os.Stdout),
		nil,
	)
}
