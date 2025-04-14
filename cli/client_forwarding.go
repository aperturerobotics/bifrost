package cli

import (
	"os"

	stream_api "github.com/aperturerobotics/bifrost/stream/api"
	"github.com/aperturerobotics/cli"
)

// RunForwarding runs the forwarding command.
func (a *ClientArgs) RunForwarding(_ *cli.Context) error {
	ctx := a.GetContext()
	c, err := a.BuildClient()
	if err != nil {
		return err
	}

	req, err := c.ForwardStreams(ctx, &stream_api.ForwardStreamsRequest{
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
