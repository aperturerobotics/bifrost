package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/stream/forwarding"
	"github.com/urfave/cli"
)

var forwardingConf stream_forwarding.Config

// runForwardController runs a forwarding controller.
func runForwardController(cctx *cli.Context) error {
	ctx := context.Background()
	c, err := GetClient()
	if err != nil {
		return err
	}

	req, err := c.ForwardStreams(ctx, &api.ForwardStreamsRequest{
		ForwardingConfig: &forwardingConf,
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
