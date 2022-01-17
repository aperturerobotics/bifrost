package cli

import (
	"os"

	stream_grpc "github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/urfave/cli"
)

// RunForwarding runs the forwarding command.
func (a *ClientArgs) RunForwarding(_ *cli.Context) error {
	ctx := a.GetContext()
	c, err := a.BuildClient()
	if err != nil {
		return err
	}

	req, err := c.ForwardStreams(ctx, &stream_grpc.ForwardStreamsRequest{
		ForwardingConfig: &a.ForwardingConf,
	})
	if err != nil {
		return err
	}

	for {
		resp, err := req.Recv()
		if err != nil {
			return err
		}

		os.Stdout.WriteString(resp.GetControllerStatus().String())
		os.Stdout.WriteString("\n")
	}
}
