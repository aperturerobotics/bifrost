package cli

import (
	"os"

	stream_grpc "github.com/aperturerobotics/bifrost/stream/grpc"
	stream_grpc_rpc "github.com/aperturerobotics/bifrost/stream/grpc/rpc"
	"github.com/aperturerobotics/bifrost/util/rwc"
	"github.com/urfave/cli"
)

// RunDial runs the dial command.
func (a *ClientArgs) RunDial(*cli.Context) error {
	ctx := a.GetContext()
	c, err := a.BuildClient()
	if err != nil {
		return err
	}

	client, err := c.DialStream(ctx)
	if err != nil {
		return err
	}

	if len(a.RemotePeerIdsCsv) != 0 {
		a.AcceptConf.RemotePeerIds = a.ParseRemotePeerIdsCsv()
	}
	err = client.Send(&stream_grpc.DialStreamRequest{
		Config: &a.DialConf,
	})
	if err != nil {
		return err
	}

	drpc := stream_grpc.NewDialStreamClientRPC(client)
	return stream_grpc_rpc.AttachRPCToStream(
		drpc,
		rwc.NewReadWriteCloser(os.Stdin, os.Stdout),
		nil,
	)
}
