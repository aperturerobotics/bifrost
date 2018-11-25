package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/stream/listening"
	"github.com/urfave/cli"
)

var listeningConf stream_listening.Config

// runListeningController runs a listening controller.
func runListeningController(cctx *cli.Context) error {
	ctx := context.Background()
	c, err := GetClient()
	if err != nil {
		return err
	}

	req, err := c.ListenStreams(ctx, &api.ListenStreamsRequest{
		ListeningConfig: &listeningConf,
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
