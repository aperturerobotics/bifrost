package cli

import (
	"os"

	stream_api "github.com/aperturerobotics/bifrost/stream/api"
	stream_api_rpc "github.com/aperturerobotics/bifrost/stream/api/rpc"
	"github.com/aperturerobotics/bifrost/util/rwc"
	"github.com/urfave/cli/v2"
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
	err = client.Send(&stream_api.AcceptStreamRequest{
		Config: &a.AcceptConf,
	})
	if err != nil {
		return err
	}

	drpc := stream_api.NewAcceptStreamClientRPC(client)
	return stream_api_rpc.AttachRPCToStream(
		drpc,
		rwc.NewReadWriteCloser(os.Stdin, os.Stdout),
		nil,
	)
}
