package cli

import (
	"os"

	stream_api "github.com/aperturerobotics/bifrost/stream/api"
	stream_api_rpc "github.com/aperturerobotics/bifrost/stream/api/rpc"
	"github.com/aperturerobotics/bifrost/util/rwc"
	"github.com/urfave/cli/v2"
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
	err = client.Send(&stream_api.DialStreamRequest{
		Config: &a.DialConf,
	})
	if err != nil {
		return err
	}

	rpcClient := stream_api.NewDialStreamClientRPC(client)
	return stream_api_rpc.AttachRPCToStream(
		rpcClient,
		rwc.NewReadWriteCloser(os.Stdin, os.Stdout),
		nil,
	)
}
