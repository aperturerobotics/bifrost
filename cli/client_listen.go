package cli

import (
	"os"

	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/urfave/cli"
)

// RunListen runs the listen command.
func (a *ClientArgs) RunListen(*cli.Context) error {
	ctx := a.GetContext()
	c, err := a.BuildClient()
	if err != nil {
		return err
	}
	req, err := c.ListenStreams(ctx, &stream_grpc.ListenStreamsRequest{
		ListeningConfig: &a.ListeningConf,
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
