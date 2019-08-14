package cli

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/urfave/cli"
)

// RunForwarding runs the forwarding command.
func (a *ClientArgs) RunForwarding(ctx context.Context, _ *cli.Context) error {
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

	return nil
}
